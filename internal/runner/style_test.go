package runner

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// operationalProjections are the five outcomes ADR-0012's taxonomy assigns to
// a runner outcome, excluding ProjectionInvalid, which is not one of them:
// Verdict.Projection() reports it only together with ok=false.
var operationalProjections = []Projection{
	ProjectionPass,
	ProjectionViolation,
	ProjectionExpectedUnchecked,
	ProjectionBlockedUnchecked,
	ProjectionOperational,
}

// allProjections additionally covers ProjectionInvalid, so tests that must
// hold for every value the vocabulary answers do not silently skip it.
var allProjections = append([]Projection{ProjectionInvalid}, operationalProjections...)

func TestGlyph_five_operational_projections_are_pairwise_distinct(t *testing.T) {
	// Given the five projections that matter operationally (ADR-0012)

	// When each is rendered to its glyph and collected into a set
	seen := make(map[string]struct{}, len(operationalProjections))
	for _, p := range operationalProjections {
		seen[Glyph(p)] = struct{}{}
	}

	// Then the set has exactly five members: no two projections share a glyph.
	if len(seen) != len(operationalProjections) {
		t.Fatalf("distinct glyphs = %d, want %d (glyphs seen: %v)", len(seen), len(operationalProjections), seen)
	}
}

func TestGlyph_invalid_projection_is_distinct_from_every_operational_glyph(t *testing.T) {
	// Given the sentinel Projection that Verdict.Projection() reports only
	// with ok=false (see the ProjectionInvalid vocabulary comment in
	// style.go for why it still renders something)
	invalid := Glyph(ProjectionInvalid)

	// When compared against each of the five real projections' glyphs
	for _, p := range operationalProjections {
		// Then none of them match: a defect must never be misread as one of
		// the five certified outcomes, least of all a pass.
		if Glyph(p) == invalid {
			t.Fatalf("Glyph(%v) = %q, collides with Glyph(ProjectionInvalid), want distinct", p, invalid)
		}
	}
}

func TestGlyph_is_exactly_one_rune_for_every_projection(t *testing.T) {
	// Given every projection the vocabulary answers for, including the
	// invalid sentinel.
	for _, p := range allProjections {
		glyph := Glyph(p)

		// Then it is exactly one rune: downstream column alignment assumes
		// one rune per glyph, and a multi-rune glyph would silently
		// misalign every line that follows it.
		if got := utf8.RuneCountInString(glyph); got != 1 {
			t.Fatalf("Glyph(%v) = %q, RuneCountInString = %d, want 1", p, glyph, got)
		}
	}
}

func TestGlyph_blocked_and_expected_unchecked_differ(t *testing.T) {
	// Given the two unchecked projections ADR-0012 requires to stay apart:
	// a blocked binding gates the run and an expected-unchecked one does
	// not, so collapsing their glyphs would undo that decision in the
	// presentation layer.
	blocked := Glyph(ProjectionBlockedUnchecked)
	expected := Glyph(ProjectionExpectedUnchecked)

	// Then they are not the same glyph.
	if blocked == expected {
		t.Fatalf("BlockedUnchecked and ExpectedUnchecked share glyph %q, want distinct (ADR-0012)", blocked)
	}
}

func TestStyle_zero_value_never_emits_escape_bytes(t *testing.T) {
	// Given the zero value of Style, which is what a forgetful caller gets
	// and what the determinism test in report_test.go will render with.
	var style Style

	// When every projection's glyph is coloured through it
	for _, p := range allProjections {
		out := style.Colorize(p, Glyph(p))

		// Then no ANSI escape byte appears anywhere in the output.
		if strings.ContainsRune(out, '\x1b') {
			t.Fatalf("Style{}.Colorize(%v, ...) = %q, contains an escape byte, want none", p, out)
		}
	}
}

func TestStyle_colour_disabled_returns_input_unchanged(t *testing.T) {
	// Given colour explicitly disabled
	style := Style{Colour: false}

	for _, p := range allProjections {
		text := Glyph(p)

		// When it colorizes a projection's glyph
		got := style.Colorize(p, text)

		// Then the output equals the input exactly, byte for byte.
		if got != text {
			t.Fatalf("Colorize(%v, %q) = %q, want unchanged input", p, text, got)
		}
	}
}

func TestStyle_colour_enabled_wraps_in_escape_and_reset(t *testing.T) {
	// Given colour explicitly enabled
	style := Style{Colour: true}

	for _, p := range operationalProjections {
		// When it colorizes a projection's glyph
		got := style.Colorize(p, Glyph(p))

		// Then the result starts with an ANSI escape and ends with a reset.
		if !strings.HasPrefix(got, "\x1b[") {
			t.Fatalf("Colorize(%v, ...) = %q, want prefix \\x1b[", p, got)
		}
		if !strings.HasSuffix(got, "\x1b[0m") {
			t.Fatalf("Colorize(%v, ...) = %q, want suffix reset \\x1b[0m", p, got)
		}
	}
}

func TestStyle_stripping_ansi_yields_colour_disabled_output(t *testing.T) {
	// Given the same glyph rendered with colour on and with colour off
	for _, p := range operationalProjections {
		plain := (Style{}).Colorize(p, Glyph(p))
		coloured := (Style{Colour: true}).Colorize(p, Glyph(p))

		// When ANSI SGR sequences are stripped from the coloured rendering
		stripped := stripANSI(coloured)

		// Then it is byte-identical to the colour-disabled rendering: colour
		// never carries meaning alone (ADR-0013).
		if stripped != plain {
			t.Fatalf("stripANSI(%q) = %q, want %q", coloured, stripped, plain)
		}
	}
}

// stripANSI removes ANSI SGR sequences (\x1b[...m) from s. Test-only: style.go
// itself never needs to parse its own escape codes back out.
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func TestLegend_contains_projection_glyph(t *testing.T) {
	// Given each of the five real projections
	for _, p := range operationalProjections {
		got := Legend(Style{}, p)
		glyph := Glyph(p)

		// When its legend line is read

		// Then it opens with that projection's glyph and carries descriptive
		// text after it.
		if !strings.HasPrefix(got, glyph) {
			t.Fatalf("Legend(%v) = %q, want prefix %q", p, got, glyph)
		}
		if strings.TrimSpace(strings.TrimPrefix(got, glyph)) == "" {
			t.Fatalf("Legend(%v) = %q, want descriptive text after the glyph", p, got)
		}
	}
}

func TestLegendLines_orders_deterministically_and_dedupes(t *testing.T) {
	// Given projections supplied out of vocabulary order and with a repeat
	given := []Projection{ProjectionOperational, ProjectionPass, ProjectionPass, ProjectionViolation}

	// When the legend lines are requested
	got := LegendLines(Style{}, given)

	// Then there are exactly three lines, in fixed vocabulary order
	// (Pass, Violation, Operational) regardless of the caller's order.
	want := []string{Legend(Style{}, ProjectionPass), Legend(Style{}, ProjectionViolation), Legend(Style{}, ProjectionOperational)}
	if len(got) != len(want) {
		t.Fatalf("LegendLines(%v) = %v (%d lines), want %d lines", given, got, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("LegendLines(%v)[%d] = %q, want %q", given, i, got[i], want[i])
		}
	}
}

func TestLegendLines_excludes_projection_invalid(t *testing.T) {
	// Given a caller that (incorrectly) includes the invalid sentinel
	got := LegendLines(Style{}, []Projection{ProjectionInvalid, ProjectionPass})

	// Then only the one legitimate projection's line is present: Invalid is
	// not one of the five reportable outcomes.
	want := Legend(Style{}, ProjectionPass)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("LegendLines with Invalid included = %v, want [%q]", got, want)
	}
}

func TestLegendGlyphIsStyledLikeTheListGlyph(t *testing.T) {
	// Given colour is enabled, as it is on the terminal the legend is for
	style := Style{Colour: true}

	// When each legend line is compared against the mark the list will draw
	for _, p := range operationalProjections {
		want := style.Colorize(p, Glyph(p))

		// Then the legend opens with those exact bytes. A key that renders
		// its mark differently from the mark it is keying is not a key.
		if !strings.HasPrefix(Legend(style, p), want) {
			t.Errorf("Legend(%v) = %q, want it to open with the list's glyph %q", p, Legend(style, p), want)
		}
	}

	for _, s := range stateOrder {
		want := style.ColorizeState(s, StateGlyph(s))
		if !strings.HasPrefix(StateLegend(style, s), want) {
			t.Errorf("StateLegend(%q) = %q, want it to open with the list's glyph %q", s, StateLegend(style, s), want)
		}
	}
}
