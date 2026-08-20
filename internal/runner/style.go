package runner

// Style carries whether ANSI colour rendering is enabled. Terminal
// detection, NO_COLOR and its siblings are decided in cmd/jigctl (see
// ADR-0013's Decision section); this file must never learn what a terminal
// is. The zero value has colour disabled, so a caller who forgets to set it
// gets plain text, and the determinism test can never accidentally hash an
// escape code.
type Style struct {
	// Colour is true when Colorize should wrap text in ANSI SGR codes.
	Colour bool
}

// glyphEntry is one projection's rendering vocabulary: the glyph opening a
// scan-list line, the ANSI SGR escape applied only when a Style enables
// colour, and the legend line printed above the list. This is a vocabulary
// table kept in its own file, the same shape as reasonData in reason.go.
// The legend describes only the projection-level meaning; the finer-grained
// reason code (e.g. "kind-not-executable", "cadence-excluded") stays the
// right-hand column's job in reason.go and is deliberately not repeated
// here, since the glyph does not distinguish between reasons that share a
// projection either (ADR-0013).
type glyphEntry struct {
	Glyph  string
	Colour string
	Legend string
}

// ANSI SGR escape sequences, written as plain string constants. No library,
// no terminal query: Colorize below only ever concatenates these.
const (
	ansiReset   = "\x1b[0m"
	ansiDim     = "\x1b[2m"
	ansiGreen   = "\x1b[32m"
	ansiRed     = "\x1b[31m"
	ansiYellow  = "\x1b[33m"
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[35m"
)

// vocabulary maps every Projection to its glyph entry. The map is total over
// all six Projection constants (five real outcomes plus ProjectionInvalid),
// so golangci-lint's exhaustive linter (check: map) holds this file to the
// same completeness report.go and exit.go already keep for their own
// Projection switches.
//
// Glyphs are geometric shapes, not emoji (ADR-0013's own example uses
// exactly the check mark, cross and hollow circle below). Pass, Violation
// and ExpectedUnchecked reuse that example verbatim. BlockedUnchecked gets
// a triangle rather than a filled circle: ADR-0012 exists specifically so
// that a blocked binding (gates the run) and an expected-unchecked one
// (does not) cannot collapse into each other, and two circles differing
// only by fill are an easy pair to misread at a glance, not a distinct
// shape. Operational gets a diamond: it is jigctl's own failure, a
// different kind of anomaly than either a binding result or a declared
// non-execution.
//
// ProjectionInvalid is the completion-unset sentinel: Verdict.Projection()
// reports it only together with ok=false, meaning a well-behaved caller
// never has a real verdict to render with it. It is not one of the five
// outcomes ADR-0012's taxonomy enumerates. It still gets its own entry,
// distinct from all five real glyphs rather than sharing one, so that if a
// defect ever lets it reach a renderer anyway, it reads as a visibly
// unrecognised glyph instead of silently masquerading as a certified
// outcome - most dangerously, as a pass.
var vocabulary = map[Projection]glyphEntry{
	ProjectionInvalid:           {Glyph: "?", Colour: ansiRed, Legend: "internal defect: no projection was derived"},
	ProjectionPass:              {Glyph: "✓", Colour: ansiGreen, Legend: "passed"},
	ProjectionViolation:         {Glyph: "✗", Colour: ansiRed, Legend: "found a violation"},
	ProjectionExpectedUnchecked: {Glyph: "○", Colour: ansiCyan, Legend: "did not run by design"},
	ProjectionBlockedUnchecked:  {Glyph: "▲", Colour: ansiYellow, Legend: "could not run and blocks the result"},
	ProjectionOperational:       {Glyph: "◆", Colour: ansiMagenta, Legend: "jigctl itself failed"},
}

// legendOrder is the fixed, deterministic order legend lines print in,
// independent of the order callers supply projections in. ProjectionInvalid
// is intentionally absent: a legend explaining an outcome that
// Verdict.Projection() itself refuses to certify (ok=false) would confuse a
// reader rather than help one.
var legendOrder = []Projection{
	ProjectionPass,
	ProjectionViolation,
	ProjectionExpectedUnchecked,
	ProjectionBlockedUnchecked,
	ProjectionOperational,
}

// Glyph returns the single-rune glyph for a projection.
func Glyph(p Projection) string {
	return vocabulary[p].Glyph
}

// Legend returns the one-line legend entry for a projection: its glyph, two
// spaces, then a short description of what the projection means.
func Legend(p Projection) string {
	entry := vocabulary[p]
	return entry.Glyph + "  " + entry.Legend
}

// LegendLines returns one legend line per distinct real projection present
// in projections, in the fixed vocabulary order above, regardless of the
// input's order or any duplicate entries in it. Whether to print the result
// at all is not this function's decision: ADR-0013 says a legend prints
// only on a run that used a glyph other than the passing one, and that
// belongs to the renderer that already collected the run's projections.
// This function only turns the set the renderer decided to print into
// ordered, deduplicated lines.
func LegendLines(projections []Projection) []string {
	present := make(map[Projection]struct{}, len(projections))
	for _, p := range projections {
		present[p] = struct{}{}
	}

	lines := make([]string, 0, len(legendOrder))
	for _, p := range legendOrder {
		if _, ok := present[p]; ok {
			lines = append(lines, Legend(p))
		}
	}
	return lines
}

// Colorize returns text styled for a projection: unchanged when s.Colour is
// false (including the Style zero value), or wrapped in that projection's
// ANSI SGR colour and a reset sequence when true. Colour is applied only by
// this method, is never learned from the environment, and is never the sole
// carrier of meaning (ADR-0013): the glyph and the legend word already
// distinguish every projection without it.
func (s Style) Colorize(p Projection, text string) string {
	if !s.Colour {
		return text
	}
	return vocabulary[p].Colour + text + ansiReset
}

// dim and accent style the marks the live view uses for a binding that has
// not settled yet. They carry no projection, because none has been derived,
// and they never reach the settled output, so they stay out of the
// projection vocabulary above.
func (s Style) dim(text string) string {
	if !s.Colour {
		return text
	}
	return ansiDim + text + ansiReset
}

func (s Style) accent(text string) string {
	if !s.Colour {
		return text
	}
	return ansiCyan + text + ansiReset
}
