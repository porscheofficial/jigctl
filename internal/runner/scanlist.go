package runner

import (
	"fmt"
	"strings"
)

func renderScanLine(opts RenderOptions, r *Row) {
	glyph := opts.Style.Colorize(r.Projection, Glyph(r.Projection))

	id := r.RecordID
	if id == "" {
		id = "--------" // placeholder if unknown
	}

	title := truncate(r.Title, 40)

	var durationStr string
	if opts.NormalizeDuration {
		if r.Execution != nil {
			durationStr = "   <dur>"
		} else {
			durationStr = "       "
		}
	} else {
		if r.Execution != nil {
			durationStr = DurationColumn(r.Execution)
		} else {
			durationStr = "       "
		}
	}

	fmt.Fprintf(opts.Out, "  %s  %-8s  %-40s  %s  %s\n",
		glyph, id, title, durationStr, scanLineEvidence(r))
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
