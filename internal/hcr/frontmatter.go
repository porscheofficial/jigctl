package hcr

import "bytes"

var (
	utf8BOM         = []byte{0xef, 0xbb, 0xbf}
	frontmatterOpen = []byte("---\n")
	frontmatterEnd  = []byte("\n---\n")
)

func extractFrontmatter(source []byte) ([]byte, bool) {
	withoutBOM := bytes.TrimPrefix(source, utf8BOM)
	// Normalize before splitting: the Python LF-only regex silently rejected
	// CRLF, making core.autocrlf=true worktrees diverge from CI.
	normalized := bytes.ReplaceAll(withoutBOM, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, frontmatterOpen) {
		return nil, false
	}
	remainder := normalized[len(frontmatterOpen):]
	end := bytes.Index(remainder, frontmatterEnd)
	if end < 0 {
		return nil, false
	}
	return remainder[:end], true
}
