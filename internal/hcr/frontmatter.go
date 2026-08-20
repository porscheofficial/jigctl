package hcr

import "bytes"

var (
	utf8BOM         = []byte{0xef, 0xbb, 0xbf}
	frontmatterOpen = []byte("---\n")
	frontmatterEnd  = []byte("\n---\n")
)

// extractFrontmatter splits source into its YAML frontmatter and the
// Markdown body that follows it. Both halves are read from the same
// normalized copy of source: a UTF-8 BOM is stripped and CRLF line endings
// are rewritten to LF before the delimiter search runs, so a
// core.autocrlf=true worktree parses identically to CI. body is everything
// after the closing "\n---\n" delimiter; that delimiter is searched for once,
// so a body that itself contains the sequence is returned whole rather than
// truncated at its first occurrence.
func extractFrontmatter(source []byte) (frontmatter, body []byte, present bool) {
	withoutBOM := bytes.TrimPrefix(source, utf8BOM)
	// Normalize before splitting: the Python LF-only regex silently rejected
	// CRLF, making core.autocrlf=true worktrees diverge from CI.
	normalized := bytes.ReplaceAll(withoutBOM, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, frontmatterOpen) {
		return nil, nil, false
	}
	remainder := normalized[len(frontmatterOpen):]
	end := bytes.Index(remainder, frontmatterEnd)
	if end < 0 {
		return nil, nil, false
	}
	return remainder[:end], remainder[end+len(frontmatterEnd):], true
}
