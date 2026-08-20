package runner

import (
	"fmt"
	"io"
	"sort"
	"unicode/utf8"
)

type RenderOptions struct {
	Out               io.Writer
	Rows              []Row
	NormalizeDuration bool
	Style             Style
}

func Render(opts RenderOptions) error {
	rows := make([]Row, len(opts.Rows))
	copy(rows, opts.Rows)

	// Sort rows
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Identity.RecordPath != rows[j].Identity.RecordPath {
			return rows[i].Identity.RecordPath < rows[j].Identity.RecordPath
		}
		return rows[i].Identity.BindingIndex < rows[j].Identity.BindingIndex
	})

	var nonPassing []*Row
	projections := make([]Projection, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		projections = append(projections, r.Projection)
		if r.Projection != ProjectionPass {
			nonPassing = append(nonPassing, r)
		}
	}

	// 1. Legend (if non-passing exists)
	if len(nonPassing) > 0 {
		legend := LegendLines(projections)
		for _, l := range legend {
			fmt.Fprintf(opts.Out, "  %s\n", l)
		}
		if len(legend) > 0 {
			fmt.Fprintln(opts.Out)
		}
	}

	// 2. Scan list
	for i := range rows {
		renderScanLine(opts, &rows[i])
	}

	// 3. Detail block
	if len(nonPassing) > 0 {
		fmt.Fprintln(opts.Out)
		for i, r := range nonPassing {
			if i > 0 {
				fmt.Fprintln(opts.Out)
			}
			renderDetail(opts, r)
		}
	}

	// 4. Ledger
	fmt.Fprintln(opts.Out)
	renderLedger(opts.Out, rows)

	return nil
}

// truncate to fixed runes.
func truncate(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes-1]) + "…"
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
