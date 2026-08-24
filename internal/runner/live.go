package runner

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

// Terminal control sequences. Only relative cursor movement is used, so the
// view never needs to know where on the screen it was started.
const (
	cursorHide      = "\x1b[?25l"
	cursorShow      = "\x1b[?25h"
	eraseLine       = "\r\x1b[2K"
	eraseBelow      = "\x1b[0J"
	cursorUpPattern = "\x1b[%dA"
)

const (
	pendingGlyph = "·"
	defaultTick  = 100 * time.Millisecond
	// minLiveWidth is the narrowest terminal the view will paint into. Below
	// it the columns no longer fit on one physical line, and a line that
	// wraps breaks the relative cursor movement the view repaints with.
	minLiveWidth = 60
	// liveHeadroom is the number of lines the view leaves free below its
	// block so that scrolling never carries the block off the screen.
	liveHeadroom = 2
)

// spinnerFrames are single-cell braille glyphs, chosen over clock faces
// because emoji are rendered two cells wide by some terminals and one by
// others, which would shift every column to their right.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// liveRecord is one record's line in flight. It tracks its bindings by count
// rather than individually because the line does not distinguish them: a
// record is running while any of its bindings is, and settles only once all
// of them have, which is what lets the finished line be the same bytes the
// settled report will print.
type liveRecord struct {
	path     string
	recordID string
	title    string
	state    string
	bindings int
	running  int
	started  time.Time
	planned  []string
	rows     []Row
}

func (r *liveRecord) settled() bool { return len(r.rows) == r.bindings }

func (r *liveRecord) evidence() string { return strings.Join(r.planned, "; ") }

func (r *liveRecord) addPlanned(text string) {
	if text == "" {
		return
	}
	for _, p := range r.planned {
		if p == text {
			return
		}
	}
	r.planned = append(r.planned, text)
}

// LiveOptions configures the transient progress view.
type LiveOptions struct {
	Out     io.Writer
	Plan    *hcr.Plan
	Style   Style
	Width   int
	Height  int
	Cadence CadenceSet
	// Tick is the repaint interval; zero selects a default.
	Tick time.Duration
}

// LiveView paints the scan list in place while a run executes, so a reader
// can see which record is in flight and how long it has been running. It is
// deliberately transient: Close erases the whole block, and the settled
// output is then printed by Render, from the same scanLine function the view
// paints its finished records with. Nothing the view writes survives the run,
// so no live frame can enter a determinism hash.
type LiveView struct {
	out     io.Writer
	style   Style
	layout  Layout
	builder *RowBuilder
	tick    time.Duration

	mu      sync.Mutex
	records []liveRecord
	index   map[BindingIdentity]int
	frame   int
	painted int
	failed  bool

	stop     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewLiveView starts a live view over the plan's records. It reports false
// when the destination cannot host one — no records to show, or a terminal
// too narrow or too short to repaint without wrapping or scrolling — in which
// case the run should simply print its settled output when it finishes.
func NewLiveView(opts LiveOptions) (*LiveView, bool) {
	if opts.Out == nil || opts.Plan == nil {
		return nil, false
	}

	records, index := liveSkeleton(opts.Plan, opts.Cadence)
	if len(records) == 0 || opts.Width < minLiveWidth || opts.Height < len(records)+liveHeadroom {
		return nil, false
	}

	tick := opts.Tick
	if tick <= 0 {
		tick = defaultTick
	}

	longest := 0
	for i := range records {
		if n := utf8.RuneCountInString(records[i].title); n > longest {
			longest = n
		}
	}

	view := &LiveView{
		out:     opts.Out,
		style:   opts.Style,
		layout:  layoutFor(opts.Width, longest),
		builder: NewRowBuilder(opts.Plan),
		tick:    tick,
		records: records,
		index:   index,
		stop:    make(chan struct{}),
	}

	view.mu.Lock()
	view.write(cursorHide + view.paint())
	view.mu.Unlock()

	view.wg.Add(1)
	go view.loop()

	return view, true
}

// Start reports that the binding is now executing.
func (v *LiveView) Start(id BindingIdentity) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if i, ok := v.index[id]; ok {
		r := &v.records[i]
		if r.running == 0 {
			r.started = time.Now()
		}
		r.running++
	}
	v.write(v.paint())
}

// Done reports the binding's settled verdict.
func (v *LiveView) Done(id BindingIdentity, verdict *Verdict) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if i, ok := v.index[id]; ok {
		r := &v.records[i]
		if r.running > 0 {
			r.running--
		}
		if verdict != nil {
			r.rows = append(r.rows, v.builder.Row(verdict))
			sort.Slice(r.rows, func(i, j int) bool {
				return r.rows[i].Identity.BindingIndex < r.rows[j].Identity.BindingIndex
			})
		}
	}
	v.write(v.paint())
}

// Close stops the animation and erases the block, returning the cursor to
// where the view began so the settled output can be printed in its place.
func (v *LiveView) Close() {
	v.stopOnce.Do(func() { close(v.stop) })
	v.wg.Wait()

	v.mu.Lock()
	defer v.mu.Unlock()

	var b strings.Builder
	if v.painted > 0 {
		fmt.Fprintf(&b, cursorUpPattern, v.painted)
		b.WriteString(eraseBelow)
	}
	b.WriteString(cursorShow)
	v.painted = 0
	v.write(b.String())
}

func (v *LiveView) loop() {
	defer v.wg.Done()
	ticker := time.NewTicker(v.tick)
	defer ticker.Stop()

	for {
		select {
		case <-v.stop:
			return
		case <-ticker.C:
			v.mu.Lock()
			v.frame++
			v.write(v.paint())
			v.mu.Unlock()
		}
	}
}

// liveSkeleton lays out one line per record, in the order the settled report
// prints them, and returns the binding-to-line index progress is reported
// through. A record bound twice occupies one line, so the block the view
// erases is the same height as the list Render prints into its place.
func liveSkeleton(plan *hcr.Plan, cadence CadenceSet) (records []liveRecord, index map[BindingIdentity]int) {
	byPath := make(map[string]*liveRecord)
	for i := range plan.Targets {
		t := &plan.Targets[i]
		for j := range t.Bindings {
			b := &t.Bindings[j]
			rec, ok := byPath[b.RecordPath]
			if !ok {
				rec = &liveRecord{
					path:     b.RecordPath,
					recordID: b.RecordID,
					title:    b.Title,
					state:    b.State,
				}
				byPath[b.RecordPath] = rec
			}
			rec.bindings++
			rec.addPlanned(plannedEvidence(b, cadence))
		}
	}

	records = make([]liveRecord, 0, len(byPath))
	for _, rec := range byPath {
		records = append(records, *rec)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].path < records[j].path })

	line := make(map[string]int, len(records))
	for i := range records {
		line[records[i].path] = i
	}

	index = make(map[BindingIdentity]int, len(records))
	for i := range plan.Targets {
		t := &plan.Targets[i]
		for j := range t.Bindings {
			b := &t.Bindings[j]
			id := BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex}
			index[id] = line[b.RecordPath]
		}
	}

	return records, index
}
