package runner

import (
	"fmt"
	"strings"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

// pendingDuration is the blank the duration column occupies when a binding
// has no execution to time.
const pendingDuration = "       "

func renderScanLine(opts RenderOptions, layout Layout, rec *Record) {
	fmt.Fprintln(opts.Out, scanLine(opts.Style, layout, rec, opts.NormalizeDuration))
}

// scanLine formats one settled record as a single scan-list line. It is the
// only place that shape exists: the live view formats its settled records
// through this same function so that the frame it paints during a run and the
// output printed after it are the same bytes.
func scanLine(style Style, layout Layout, rec *Record, normalize bool) string {
	return composeScanLine(
		layout,
		style.Colorize(rec.Projection, Glyph(rec.Projection)),
		style.ColorizeState(rec.State, StateGlyph(rec.State)),
		rec.RecordID,
		rec.Title,
		durationCell(rec, normalize),
		recordEvidence(rec),
	)
}

// composeScanLine assembles the fixed column geometry every scan-list line
// shares, settled or in flight. Only the caller knows what belongs in each
// cell; the widths belong here.
func composeScanLine(layout Layout, glyph, state, id, title, duration, evidence string) string {
	if id == "" {
		id = "--------" // placeholder if unknown
	}
	line := fmt.Sprintf("  %s %s  %-8s  %-*s  %s  %s",
		glyph, state, id, layout.Title, truncate(title, layout.Title),
		duration, clip(evidence, layout.Evidence))
	return strings.TrimRight(line, " ")
}

func durationCell(rec *Record, normalize bool) string {
	total, executed := recordDuration(rec)
	if !executed {
		return pendingDuration
	}
	if normalize {
		return "   <dur>"
	}
	return fmt.Sprintf("%7s", formatDuration(total))
}

// bindingEvidence describes what one binding did, or why it did not, in the
// words the human report uses. It deliberately never returns a finding count:
// how many findings a record has is a question only the record can answer,
// and recordEvidence totals it across the bindings this function left empty.
func bindingEvidence(r *Row) string {
	switch {
	case r.IsUnknown:
		return reasonPhrase(r.Reason)
	case r.Reason != ReasonNone && r.Projection != ProjectionPass && r.Projection != ProjectionViolation:
		return notRunPhrase(r)
	case r.Kind == "external":
		return externalEvidence(r.Tool, r.Docs)
	case r.Kind == "command" && r.Execution != nil && r.Execution.ExitCode != 0:
		return fmt.Sprintf("%s (exit: %d)", strings.Join(r.Execution.Argv, " "), r.Execution.ExitCode)
	case r.Kind == "command" && r.Execution != nil:
		return strings.Join(r.Execution.Argv, " ")
	}
	return ""
}

// notRunPhrase says why a binding did not run, given that the line already
// carries a state glyph and an outcome glyph.
//
// A draft or deprecated record returns nothing at all: the state column has
// already said it, and repeating it in words two columns later is the
// duplication the state column exists to remove. `kind-not-executable` is the
// opposite case — it names a condition without naming its cause — so it is
// resolved against the binding's kind, which is where the cause actually is.
func notRunPhrase(r *Row) string {
	if r.Reason == Reason(ReasonRecordDraft) || r.Reason == Reason(ReasonRecordDeprecated) {
		return ""
	}
	if r.Reason == Reason(ReasonKindNotExecutable) {
		return unexecutableKind(r.Kind, r.Tool, r.Docs)
	}
	return reasonPhrase(r.Reason)
}

// plannedEvidence describes what a binding is about to do, in the bytes
// bindingEvidence will use once it has done it — that is the invariant, and
// breaking it makes the line rewrite itself the instant the binding settles.
//
// A binding Select would not attempt describes nothing: naming a command that
// is never going to run reads as a promise the run has already broken.
func plannedEvidence(b *hcr.ExecutableBinding) string {
	if b.Kind != "command" || Select(&VerdictReport{}, b, DefaultCadenceSet()) != nil {
		return ""
	}
	return strings.Join(strings.Fields(b.Run), " ")
}
