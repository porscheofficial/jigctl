package runner

import (
	"fmt"
	"io"
	"strings"
)

// RenderPlain renders a given slice of evaluated Rows in the exact one-line-per-binding
// machine-readable output format that jigctl emitted prior to ADR-0013.
func RenderPlain(out io.Writer, rows []Row) error {
	for i := range rows {
		printLinePlain(out, &rows[i])
	}
	printSummaryPlain(out, rows)
	return nil
}

func describePlain(row *Row) string {
	switch {
	case row.IsUnknown:
		return fmt.Sprintf("[unknown] %s", reasonMessage(row.Reason))
	case row.Projection == ProjectionViolation:
		return fmt.Sprintf("[%s] %s", row.RecordID, row.Summary)
	case row.Projection == ProjectionPass || row.Projection == ProjectionInvalid:
		return fmt.Sprintf("[%s]", row.RecordID)
	default:
		return fmt.Sprintf("[%s] %s", row.RecordID, reasonMessage(row.Reason))
	}
}

func printLinePlain(out io.Writer, row *Row) {
	var msgBuilder strings.Builder
	msgBuilder.WriteString(describePlain(row))

	if row.Kind == "command" && row.Execution != nil {
		fmt.Fprintf(&msgBuilder, " (argv: %s, exit: %d, duration: %s)",
			strings.Join(row.Execution.Argv, " "), row.Execution.ExitCode, formatDuration(row.Execution.Duration))
	}

	if row.Kind == "external" && !row.IsUnknown {
		fmt.Fprintf(&msgBuilder, " (tool: %s, docs: %s)",
			row.Tool, row.Docs)
	}

	if row.UnwaivedCount > 0 {
		fmt.Fprintf(&msgBuilder, " [%d finding(s)]", row.UnwaivedCount)
	}

	fmt.Fprintf(out, "%s: %s: %s\n",
		row.Locator, projectionCode(row.Projection, row.Reason), msgBuilder.String())
}

func printSummaryPlain(out io.Writer, rows []Row) {
	var expectedUnchecked, blockedUnchecked, findings int

	for i := range rows {
		switch rows[i].Projection {
		case ProjectionExpectedUnchecked:
			expectedUnchecked++
		case ProjectionBlockedUnchecked:
			blockedUnchecked++
		case ProjectionPass, ProjectionViolation, ProjectionOperational, ProjectionInvalid:
		}
		findings += rows[i].UnwaivedCount
	}

	numFiles := UnwaivedFileCount(rows)

	wordFindings := "findings"
	if findings == 1 {
		wordFindings = "finding"
	}
	wordFiles := "files"
	if numFiles == 1 {
		wordFiles = "file"
	}

	if findings == 0 {
		fmt.Fprintf(out, "no findings, %d expected-unchecked, %d blocked-unchecked\n",
			expectedUnchecked, blockedUnchecked)
	} else {
		fmt.Fprintf(out, "%d %s in %d %s, %d expected-unchecked, %d blocked-unchecked\n",
			findings, wordFindings, numFiles, wordFiles,
			expectedUnchecked, blockedUnchecked)
	}
}
