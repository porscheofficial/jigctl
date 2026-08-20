package runner

import (
	"fmt"
	"strings"
)

// renderDetail repeats one record the reader has to act on: its heading, why
// its bindings did not complete, the summary the author wrote, and every
// unwaived finding. A record is repeated once however many bindings verify
// it, since the reader is being pointed back at a rule, not at a mechanism.
func renderDetail(opts RenderOptions, layout Layout, rec *Record) {
	renderDetailHeading(opts, layout, rec)
	renderDetailReasons(opts, layout, rec)
	renderDetailSummary(opts, layout, rec)
	renderDetailLoci(opts, rec)
}

func renderDetailHeading(opts RenderOptions, layout Layout, rec *Record) {
	glyph := opts.Style.Colorize(rec.Projection, Glyph(rec.Projection))
	state := opts.Style.ColorizeState(rec.State, StateGlyph(rec.State))
	id := rec.RecordID
	if id == "" {
		id = "--------"
	}

	lines := wrapText(rec.Title, layout.Heading)
	if len(lines) == 0 {
		fmt.Fprintf(opts.Out, "  %s %s  %-8s\n", glyph, state, id)
		return
	}

	for i, line := range lines {
		if i == 0 {
			fmt.Fprintf(opts.Out, "  %s %s  %-8s  %s\n", glyph, state, id, line)
			continue
		}
		fmt.Fprintf(opts.Out, "%s%s\n", strings.Repeat(" ", headingIndent), line)
	}
}

func renderDetailReasons(opts RenderOptions, layout Layout, rec *Record) {
	seen := make(map[string]struct{}, len(rec.Rows))
	for i := range rec.Rows {
		text := detailReason(&rec.Rows[i])
		if text == "" {
			continue
		}
		if _, duplicate := seen[text]; duplicate {
			continue
		}
		seen[text] = struct{}{}
		printProse(opts, layout, text)
	}
}

func detailReason(r *Row) string {
	if r.IsUnknown {
		return "[unknown] " + reasonPhrase(r.Reason)
	}
	if r.Reason != ReasonNone && r.Projection != ProjectionViolation && r.Projection != ProjectionPass {
		return notRunPhrase(r)
	}
	return ""
}

func renderDetailSummary(opts RenderOptions, layout Layout, rec *Record) {
	if rec.Summary == "" {
		return
	}
	for _, rawLine := range strings.Split(strings.TrimSpace(rec.Summary), "\n") {
		printProse(opts, layout, rawLine)
	}
}

// renderDetailLoci prints every unwaived finding of every binding. Each is
// formatted against the kind of the binding that produced it, not the
// record's, because a record may bind more than one kind and a locus means a
// different thing per kind (ADR-0013's first Note).
func renderDetailLoci(opts RenderOptions, rec *Record) {
	for i := range rec.Rows {
		r := &rec.Rows[i]
		for _, f := range r.Findings {
			if len(f.WaivedBy) > 0 {
				continue
			}
			if locus := formatLocus(r.Kind, f.Locus); locus != "" {
				fmt.Fprintf(opts.Out, "       %s\n", locus)
			}
		}
	}
}

func printProse(opts RenderOptions, layout Layout, text string) {
	for _, line := range wrapText(text, layout.Prose) {
		fmt.Fprintf(opts.Out, "       %s\n", line)
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
