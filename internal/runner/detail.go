package runner

import (
	"fmt"
	"strings"
)

func renderDetail(opts RenderOptions, r *Row) {
	glyph := opts.Style.Colorize(r.Projection, Glyph(r.Projection))
	id := r.RecordID
	if id == "" {
		id = "--------"
	}

	titleLines := wrapText(r.Title)
	if len(titleLines) == 0 {
		fmt.Fprintf(opts.Out, "  %s  %-8s  \n", glyph, id)
	} else {
		for i, line := range titleLines {
			if i == 0 {
				fmt.Fprintf(opts.Out, "  %s  %-8s  %s\n", glyph, id, line)
			} else {
				fmt.Fprintf(opts.Out, "               %s\n", line)
			}
		}
	}

	renderDetailSummary(opts, r)

	for _, f := range r.Findings {
		if len(f.WaivedBy) > 0 {
			continue
		}

		locusStr := formatLocus(r.Kind, f.Locus)
		if locusStr != "" {
			fmt.Fprintf(opts.Out, "       %s\n", locusStr)
		}
	}
}

func renderDetailSummary(opts RenderOptions, r *Row) {
	// First, if it's non-passing, print the reason
	if r.IsUnknown {
		wrappedLines := wrapText("[unknown] " + reasonMessage(r.Reason))
		for _, line := range wrappedLines {
			fmt.Fprintf(opts.Out, "       %s\n", line)
		}
	} else if r.Reason != 0 && r.Projection != ProjectionViolation && r.Projection != ProjectionPass {
		wrappedLines := wrapText(reasonMessage(r.Reason))
		for _, line := range wrappedLines {
			fmt.Fprintf(opts.Out, "       %s\n", line)
		}
	}

	// Then, if there is a summary, print it as well.
	if r.Summary != "" {
		lines := strings.Split(strings.TrimSpace(r.Summary), "\n")
		for _, rawLine := range lines {
			wrappedLines := wrapText(rawLine)
			for _, line := range wrappedLines {
				fmt.Fprintf(opts.Out, "       %s\n", line)
			}
		}
	}
}

func formatLocus(kind string, locus Locus) string {
	switch kind {
	case "command":
		return ""
	case "grep":
		if locus.Pointer == "" {
			return fmt.Sprintf("[pattern] %s", locus.File)
		}
		ptr := strings.TrimPrefix(locus.Pointer, "L")
		return fmt.Sprintf("%s:%s", locus.File, ptr)
	case "config-assert":
		return fmt.Sprintf("%s %s", locus.File, locus.Pointer)
	default:
		if locus.File != "" && locus.Pointer != "" {
			return fmt.Sprintf("%s %s", locus.File, locus.Pointer)
		} else if locus.File != "" {
			return locus.File
		}
		return ""
	}
}
