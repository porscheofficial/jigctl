package runner

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/patricebouillet/jigctl/internal/hcr"
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

type liveRow struct {
	identity BindingIdentity
	recordID string
	title    string
	started  time.Time
	running  bool
	row      Row
	settled  bool
}

// LiveOptions configures the transient progress view.
type LiveOptions struct {
	Out    io.Writer
	Plan   *hcr.Plan
	Style  Style
	Width  int
	Height int
	// Tick is the repaint interval; zero selects a default.
	Tick time.Duration
}

// LiveView paints the scan list in place while a run executes, so a reader
// can see which binding is in flight and how long it has been running. It is
// deliberately transient: Close erases the whole block, and the settled
// output is then printed by Render, from the same scanLine function the view
// paints its finished rows with. Nothing the view writes survives the run, so
// no live frame can enter a determinism hash.
type LiveView struct {
	out     io.Writer
	style   Style
	layout  Layout
	builder *RowBuilder
	tick    time.Duration

	mu      sync.Mutex
	rows    []liveRow
	index   map[BindingIdentity]int
	frame   int
	painted int
	failed  bool

	stop     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewLiveView starts a live view over the plan's bindings. It reports false
// when the destination cannot host one — no bindings to show, or a terminal
// too narrow or too short to repaint without wrapping or scrolling — in which
// case the run should simply print its settled output when it finishes.
func NewLiveView(opts LiveOptions) (*LiveView, bool) {
	if opts.Out == nil || opts.Plan == nil {
		return nil, false
	}

	rows := liveSkeleton(opts.Plan)
	if len(rows) == 0 || opts.Width < minLiveWidth || opts.Height < len(rows)+liveHeadroom {
		return nil, false
	}

	tick := opts.Tick
	if tick <= 0 {
		tick = defaultTick
	}

	longest := 0
	index := make(map[BindingIdentity]int, len(rows))
	for i := range rows {
		index[rows[i].identity] = i
		if n := utf8.RuneCountInString(rows[i].title); n > longest {
			longest = n
		}
	}

	view := &LiveView{
		out:     opts.Out,
		style:   opts.Style,
		layout:  layoutFor(opts.Width, longest),
		builder: NewRowBuilder(opts.Plan),
		tick:    tick,
		rows:    rows,
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
		v.rows[i].running = true
		v.rows[i].started = time.Now()
	}
	v.write(v.paint())
}

// Done reports the binding's settled verdict.
func (v *LiveView) Done(id BindingIdentity, verdict *Verdict) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if i, ok := v.index[id]; ok {
		r := &v.rows[i]
		r.running = false
		if verdict != nil {
			r.row = v.builder.Row(verdict)
			r.settled = true
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

func liveSkeleton(plan *hcr.Plan) []liveRow {
	total := 0
	for i := range plan.Targets {
		total += len(plan.Targets[i].Bindings)
	}

	rows := make([]liveRow, 0, total)
	for i := range plan.Targets {
		t := &plan.Targets[i]
		for j := range t.Bindings {
			b := &t.Bindings[j]
			rows = append(rows, liveRow{
				identity: BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex},
				recordID: b.RecordID,
				title:    b.Title,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].identity.RecordPath != rows[j].identity.RecordPath {
			return rows[i].identity.RecordPath < rows[j].identity.RecordPath
		}
		return rows[i].identity.BindingIndex < rows[j].identity.BindingIndex
	})

	return rows
}
