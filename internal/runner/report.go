package runner

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// RenderOptions carries everything Render needs to print one run.
type RenderOptions struct {
	Out  io.Writer
	Rows []Row
	// NormalizeDuration replaces every measured duration with a fixed
	// placeholder so a rendering can be hashed for determinism.
	NormalizeDuration bool
	// Width is the destination's rune width, or zero when it has none.
	// Terminal detection belongs to cmd/jigctl (ADR-0013); this package is
	// only ever told the answer.
	Width int
	Style Style
}

// Render prints the human-facing shape of a run: a legend when the run needs
// one, one scan-list line per record, and a detail block repeating every
// record the reader has to act on.
func Render(opts RenderOptions) error {
	records := GroupRecords(opts.Rows)
	layout := ComputeLayout(opts.Width, records)

	renderLegend(opts, records)

	for i := range records {
		renderScanLine(opts, layout, &records[i])
	}

	detailed := make([]*Record, 0, len(records))
	for i := range records {
		if needsDetail(&records[i]) {
			detailed = append(detailed, &records[i])
		}
	}

	if len(detailed) > 0 {
		fmt.Fprintln(opts.Out)
		for i, rec := range detailed {
			if i > 0 {
				fmt.Fprintln(opts.Out)
			}
			renderDetail(opts, layout, rec)
		}
	}

	return nil
}

// needsDetail reports whether a record's outcome asks somebody to do
// something now, which is the only reason the detail block exists.
//
// Expected-unchecked is absent in full. Repeating one in detail — glyph, id,
// title, summary — says nothing the scan list did not, because nothing about
// it is wrong. A draft declares that a rule is not enforced yet; an
// inferential or agent-review binding declares that a judgement is owed to a
// human or an agent. Both will declare the same thing on every run until
// somebody acts, so both would sit in the block permanently, spending exactly
// the attention it exists to direct at the records that are broken. The scan
// line names them and their record is one glob from their id.
func needsDetail(rec *Record) bool {
	return rec.Projection != ProjectionPass && rec.Projection != ProjectionExpectedUnchecked
}

// renderLegend prints the vocabulary the run actually used, once, above the
// list. A run that is entirely passing and entirely enforced needs none: both
// its glyphs are the ones a reader expects to see and neither carries news.
func renderLegend(opts RenderOptions, records []Record) {
	projections := make([]Projection, 0, len(records))
	states := make([]string, 0, len(records))
	for i := range records {
		projections = append(projections, records[i].Projection)
		states = append(states, records[i].State)
	}

	if !needsLegend(projections, states) {
		return
	}

	lines := LegendLines(opts.Style, projections)
	lines = append(lines, StateLegendLines(opts.Style, states)...)
	if len(lines) == 0 {
		return
	}

	for _, line := range lines {
		fmt.Fprintf(opts.Out, "  %s\n", line)
	}
	fmt.Fprintln(opts.Out)
}

func needsLegend(projections []Projection, states []string) bool {
	for _, p := range projections {
		if p != ProjectionPass {
			return true
		}
	}
	for _, s := range states {
		if _, known := stateVocabulary[s]; known && s != "enforced" {
			return true
		}
	}
	return false
}

// truncate bounds s to maxRunes, marking the cut with an ellipsis.
func truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes-1]) + "…"
}

// clip bounds s the same way truncate does, except that a maxRunes of zero
// or less means the column is unbounded rather than empty.
func clip(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	return truncate(s, maxRunes)
}

func projectionCode(proj Projection, r Reason) string {
	switch proj {
	case ProjectionPass:
		return "pass"
	case ProjectionViolation:
		return "violation"
	case ProjectionExpectedUnchecked, ProjectionBlockedUnchecked, ProjectionOperational:
		return reasonCode(r)
	case ProjectionInvalid:
		return "unknown"
	}
	return "unknown"
}
