package runner

import (
	"fmt"
	"io"
)

func renderLedger(out io.Writer, rows []Row) {
	passes := 0
	violations := 0
	unchecked := 0
	blocked := 0
	errors := 0

	for i := range rows {
		switch rows[i].Projection {
		case ProjectionPass:
			passes++
		case ProjectionViolation:
			violations++
		case ProjectionExpectedUnchecked:
			unchecked++
		case ProjectionBlockedUnchecked:
			blocked++
		case ProjectionOperational, ProjectionInvalid:
			errors++
		}
	}

	fmt.Fprintf(out, "  Done. PASS=%d VIOLATION=%d UNCHECKED=%d BLOCKED=%d ERROR=%d\n",
		passes, violations, unchecked, blocked, errors)
}
