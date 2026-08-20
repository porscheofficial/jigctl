package runner

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// paint returns the escape sequence that redraws the whole block in place.
// The block's height never changes, so a single relative cursor move back to
// its first line is enough; each line is then erased and rewritten. The
// caller must hold v.mu.
func (v *LiveView) paint() string {
	var b strings.Builder
	if v.painted > 0 {
		fmt.Fprintf(&b, cursorUpPattern, v.painted)
	}
	for i := range v.records {
		b.WriteString(eraseLine)
		b.WriteString(v.liveLine(&v.records[i]))
		b.WriteString("\n")
	}
	v.painted = len(v.records)
	return b.String()
}

// write emits one frame. A destination that stops accepting writes stops the
// animation rather than failing the run: the view is decoration, and the
// settled output is printed by Render either way. The caller must hold v.mu.
func (v *LiveView) write(frame string) {
	if v.failed || frame == "" {
		return
	}
	if _, err := io.WriteString(v.out, frame); err != nil {
		v.failed = true
	}
}

func (v *LiveView) liveLine(r *liveRecord) string {
	if r.settled() {
		settled := newRecord(r.path, r.rows)
		return scanLine(v.style, v.layout, &settled, false)
	}

	state := v.style.ColorizeState(r.state, StateGlyph(r.state))
	if r.running > 0 {
		return composeScanLine(v.layout,
			v.style.accent(pendingGlyph), state, r.recordID, r.title,
			liveDurationCell(v.frame, time.Since(r.started)), r.evidence())
	}
	return composeScanLine(v.layout,
		v.style.dim(pendingGlyph), state, r.recordID, r.title, pendingDuration, r.evidence())
}

// liveDurationCell renders the spinner and the elapsed time into the same
// seven cells a settled duration occupies. Elapsed time is truncated to a
// tenth of a second, which bounds it to five runes and leaves exactly the two
// the spinner and its separator need.
func liveDurationCell(frame int, elapsed time.Duration) string {
	spinner := spinnerFrames[frame%len(spinnerFrames)]
	return fmt.Sprintf("%7s", spinner+" "+formatDuration(elapsed.Truncate(100*time.Millisecond)))
}
