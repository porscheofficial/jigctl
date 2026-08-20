package runner

import (
	"fmt"
	"io"
	"sort"
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
// one, one scan-list line per binding, and a detail block repeating every
// binding that did not pass.
func Render(opts RenderOptions) error {
	rows := sortRows(opts.Rows)
	layout := ComputeLayout(opts.Width, rows)

	var nonPassing []*Row
	projections := make([]Projection, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		projections = append(projections, r.Projection)
		if r.Projection != ProjectionPass {
			nonPassing = append(nonPassing, r)
		}
	}

	if len(nonPassing) > 0 {
		legend := LegendLines(projections)
		for _, l := range legend {
			fmt.Fprintf(opts.Out, "  %s\n", l)
		}
		if len(legend) > 0 {
			fmt.Fprintln(opts.Out)
		}
	}

	for i := range rows {
		renderScanLine(opts, layout, &rows[i])
	}

	if len(nonPassing) > 0 {
		fmt.Fprintln(opts.Out)
		for i, r := range nonPassing {
			if i > 0 {
				fmt.Fprintln(opts.Out)
			}
			renderDetail(opts, layout, r)
		}
	}

	return nil
}

// sortRows returns a copy of rows in the order a run is always printed in:
// by record path, then by binding index within a record.
func sortRows(in []Row) []Row {
	rows := make([]Row, len(in))
	copy(rows, in)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Identity.RecordPath != rows[j].Identity.RecordPath {
			return rows[i].Identity.RecordPath < rows[j].Identity.RecordPath
		}
		return rows[i].Identity.BindingIndex < rows[j].Identity.BindingIndex
	})
	return rows
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
