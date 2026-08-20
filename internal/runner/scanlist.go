package runner

import (
	"fmt"
	"strings"
)

// pendingDuration is the blank the duration column occupies when a binding
// has no execution to time.
const pendingDuration = "       "

func renderScanLine(opts RenderOptions, layout Layout, r *Row) {
	fmt.Fprintln(opts.Out, scanLine(opts.Style, layout, r, opts.NormalizeDuration))
}

// scanLine formats one settled binding as a single scan-list line. It is the
// only place that shape exists: the live view formats its settled rows
// through this same function so that the frame it paints during a run and the
// output printed after it are the same bytes.
func scanLine(style Style, layout Layout, r *Row, normalize bool) string {
	glyph := style.Colorize(r.Projection, Glyph(r.Projection))
	return composeScanLine(
		layout,
		glyph,
		r.RecordID,
		r.Title,
		durationCell(r.Execution, normalize),
		scanLineEvidence(r),
	)
}

// composeScanLine assembles the fixed column geometry every scan-list line
// shares, settled or in flight. Only the caller knows what belongs in each
// cell; the widths belong here.
func composeScanLine(layout Layout, glyph, id, title, duration, evidence string) string {
	if id == "" {
		id = "--------" // placeholder if unknown
	}
	line := fmt.Sprintf("  %s  %-8s  %-*s  %s  %s",
		glyph, id, layout.Title, truncate(title, layout.Title),
		duration, clip(evidence, layout.Evidence))
	return strings.TrimRight(line, " ")
}

func durationCell(exec *Execution, normalize bool) string {
	if exec == nil {
		return pendingDuration
	}
	if normalize {
		return "   <dur>"
	}
	return DurationColumn(exec)
}

func scanLineEvidence(r *Row) string {
	switch {
	case r.Projection != ProjectionPass && r.Projection != ProjectionViolation && r.Reason != 0:
		return reasonCode(r.Reason)
	case r.Kind == "command" && r.Execution != nil:
		if r.Execution.ExitCode != 0 {
			return fmt.Sprintf("%s (exit: %d)", strings.Join(r.Execution.Argv, " "), r.Execution.ExitCode)
		}
		return strings.Join(r.Execution.Argv, " ")
	case r.Kind == "external" && !r.IsUnknown:
		return fmt.Sprintf("tool: %s, docs: %s", r.Tool, r.Docs)
	case r.UnwaivedCount > 0:
		if r.UnwaivedCount == 1 {
			return "1 finding"
		}
		return fmt.Sprintf("%d findings", r.UnwaivedCount)
	case r.IsUnknown:
		return reasonCode(r.Reason)
	}
	return ""
}
