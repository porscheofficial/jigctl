package runner

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// wrapText breaks a string into lines of at most 72 runes.
// It breaks on whitespace boundaries. A word longer than 72 runes is placed on its own line.
func wrapText(s string) []string {
	var lines []string

	words := strings.FieldsFunc(s, unicode.IsSpace)
	if len(words) == 0 {
		return nil
	}

	var currentLine strings.Builder
	currentLen := 0
	maxWidth := 72

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)

		if currentLen == 0 {
			currentLine.WriteString(word)
			currentLen = wordLen
		} else {
			if currentLen+1+wordLen <= maxWidth {
				currentLine.WriteRune(' ')
				currentLine.WriteString(word)
				currentLen += 1 + wordLen
			} else {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(word)
				currentLen = wordLen
			}
		}
	}

	if currentLen > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
