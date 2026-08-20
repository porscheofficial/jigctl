package runner

import "unicode/utf8"

// Layout is the horizontal budget one run is rendered into. It is derived
// once, from the rows about to be printed and the width of the destination
// they are printed to, and every variable-width column reads its size from
// it rather than from a constant.
//
// A Width of zero means the destination has no measurable width — a pipe, a
// file, a test buffer — in which case nothing is truncated and the columns
// are sized purely by content. Rendering therefore stays a pure function of
// the rows whenever no terminal is involved, which is what the determinism
// tests hash.
type Layout struct {
	// Title is the rune width of the scan list's title column.
	Title int
	// Evidence is the rune width of the scan list's right-hand column, or
	// zero when it is unbounded.
	Evidence int
	// Prose is the rune width wrapped text in the detail block is broken at.
	Prose int
	// Heading is the rune width the detail block's title is broken at. It is
	// narrower than Prose because a detail heading carries the glyph and the
	// record id ahead of it, on the first line and by indent on the rest.
	Heading int
}

const (
	// scanFixedCells counts every cell of a scan line that is neither the
	// title nor the evidence: the two-cell indent, the glyph, the eight-cell
	// record id, the seven-cell duration column, and the four two-cell gaps
	// separating them.
	scanFixedCells = 2 + 1 + 8 + 7 + 2*4

	// detailIndent is the left margin the detail block wraps prose into.
	detailIndent = 7

	// headingIndent is the left margin a detail heading wraps into: the
	// two-cell indent, the glyph, the eight-cell record id and two gaps.
	headingIndent = 2 + 1 + 8 + 2*2

	// minTitleCells keeps a title readable when the terminal is too narrow to
	// honour both variable columns at their natural size.
	minTitleCells = 16

	// minEvidenceCells is what the title budget reserves for the right-hand
	// column, so a long title cannot push a command line or a reason code off
	// the row. It is a reservation, not a floor on the column itself: a
	// terminal too narrow to honour it gives the column whatever is left
	// rather than overflowing the width and wrapping the row.
	minEvidenceCells = 16

	// minProseCells is the narrowest the detail block will wrap to.
	minProseCells = 24

	// fallbackProseCells is the detail block's wrap width when the
	// destination has no measurable width.
	fallbackProseCells = 72
)

// ComputeLayout sizes the variable columns for one run. Width is the
// destination's rune width, or zero when it has none.
func ComputeLayout(width int, rows []Row) Layout {
	longest := 0
	for i := range rows {
		if n := utf8.RuneCountInString(rows[i].Title); n > longest {
			longest = n
		}
	}
	return layoutFor(width, longest)
}

func layoutFor(width, longestTitle int) Layout {
	layout := Layout{
		Title:   longestTitle,
		Prose:   fallbackProseCells,
		Heading: fallbackProseCells,
	}
	if width <= 0 {
		return layout
	}

	budget := width - scanFixedCells - minEvidenceCells
	if budget < minTitleCells {
		budget = minTitleCells
	}
	if layout.Title > budget {
		layout.Title = budget
	}

	layout.Evidence = atLeast(width-scanFixedCells-layout.Title, 1)

	layout.Prose = atLeast(width-detailIndent, minProseCells)
	layout.Heading = atLeast(width-headingIndent, minProseCells)

	return layout
}

func atLeast(n, floor int) int {
	if n < floor {
		return floor
	}
	return n
}
