package runner

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapText(t *testing.T) {
	text := "corpus/ is normative. When jigctl disagrees with a fixture's declared valid, at or covers, fix jigctl — never the expectation. A fixture diff shows that an expectation changed, never whether the change was justified."

	lines := wrapText(text)

	if len(lines) == 1 {
		t.Fatalf("expected text to be wrapped into multiple lines")
	}

	for i, line := range lines {
		if utf8.RuneCountInString(line) > 72 {
			t.Errorf("line %d exceeds 72 runes: %d", i, utf8.RuneCountInString(line))
		}
	}

	// Also test word that exceeds max width (should just be its own line)
	longWord := strings.Repeat("a", 80)
	lines = wrapText(longWord)
	if len(lines) != 1 || lines[0] != longWord {
		t.Errorf("long word should be on its own line")
	}
}
