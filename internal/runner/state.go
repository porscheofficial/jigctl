package runner

// stateEntry is one lifecycle position's rendering vocabulary: the glyph
// standing in the scan list's state column, the ANSI SGR escape applied only
// when a Style enables colour, and the legend line printed above the list.
// It is the same shape as glyphEntry in style.go and reasonData in reason.go,
// and is kept in its own file for the same reason.
type stateEntry struct {
	Glyph  string
	Colour string
	Legend string
}

// stateVocabulary maps every value of the schema's `state` enum to its glyph.
//
// The three active positions form a ramp rather than three unrelated marks,
// because they are not three unrelated things: draft, warn and enforced are
// increasing degrees of the same property, how much force the record has over
// the result. A shade ramp reads as an amount without being read, which is
// the same property the outcome glyph is chosen for, and it costs one cell
// where the word `deprecated` costs ten.
//
// `deprecated` deliberately leaves the ramp. It is not less force than a
// draft, it is a record withdrawn from the ramp altogether, so it takes a
// mark from a different family instead of a fainter shade that would invite
// the reader to order it against the other three.
//
// Colour never carries the distinction alone (ADR-0013), and here it barely
// carries it at all: the ramp is made of shading, so the only styling that
// belongs on it is intensity. Dim makes a light shade lighter, which is the
// axis the glyph is already on. A hue is a second axis competing with it, and
// `warn` briefly had one — yellow on a half-filled block reads as an alarm
// rather than as a position on the ramp, and drowns out the fill the reader
// is actually meant to see. `enforced`, the overwhelmingly common case, stays
// unstyled so the state column never competes with the outcome glyph that
// opens the line.
var stateVocabulary = map[string]stateEntry{
	"draft":      {Glyph: "░", Colour: ansiDim, Legend: "draft: not enforced yet"},
	"warn":       {Glyph: "▒", Colour: "", Legend: "warn: reports without blocking"},
	"enforced":   {Glyph: "█", Colour: "", Legend: "enforced: gates the result"},
	"deprecated": {Glyph: "╳", Colour: ansiDim, Legend: "deprecated: kept for history"},
}

// stateOrder is the fixed order state legend lines print in, independent of
// the order records happen to be read in. It follows the lifecycle rather
// than the alphabet.
var stateOrder = []string{"draft", "warn", "enforced", "deprecated"}

// stateGlyphCell is the width of the state column: one glyph, or one space
// when a row has no state to report.
const stateGlyphCell = " "

// StateGlyph returns the single-rune glyph for a record's lifecycle state, or
// a space when the state is absent or unrecognised. A blank cell is the right
// answer there rather than an error mark: the only way to reach it is a
// verdict whose binding is missing from the plan, which the outcome glyph
// already flags as a defect, and a second unrecognised mark on the same line
// would say the same thing twice.
func StateGlyph(state string) string {
	if entry, ok := stateVocabulary[state]; ok {
		return entry.Glyph
	}
	return stateGlyphCell
}

// StateLegend returns the one-line legend entry for a state: its glyph, two
// spaces, then what that lifecycle position means. The glyph is styled
// exactly as the list styles it, for the reason given on Legend.
func StateLegend(style Style, state string) string {
	entry := stateVocabulary[state]
	return style.ColorizeState(state, entry.Glyph) + "  " + entry.Legend
}

// StateLegendLines returns one legend line per distinct recognised state
// present in states, in lifecycle order, regardless of the input's order or
// its duplicates. Whether to print the result at all is the renderer's
// decision, not this function's.
func StateLegendLines(style Style, states []string) []string {
	present := make(map[string]struct{}, len(states))
	for _, s := range states {
		present[s] = struct{}{}
	}

	lines := make([]string, 0, len(stateOrder))
	for _, s := range stateOrder {
		if _, ok := present[s]; ok {
			lines = append(lines, StateLegend(style, s))
		}
	}
	return lines
}

// ColorizeState returns a state glyph styled for its lifecycle position:
// unchanged when s.Colour is false, when the state is unrecognised, or when
// that state has no colour of its own.
func (s Style) ColorizeState(state, text string) string {
	entry, ok := stateVocabulary[state]
	if !s.Colour || !ok || entry.Colour == "" {
		return text
	}
	return entry.Colour + text + ansiReset
}
