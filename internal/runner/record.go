package runner

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Record is every binding of one HCR record collapsed into the single unit a
// run reports on. A record is what an author writes, what a reader remembers
// and what a rule is identified by; that a record happens to be verified by
// two bindings rather than one is an implementation detail of the record, not
// a second rule, and printing it twice invites the reader to count it twice.
type Record struct {
	RecordID   string
	Path       string
	Title      string
	State      string
	Body       string
	TargetKind string
	TargetPath string
	Summary    string
	Projection Projection
	Rows       []Row
}

// projectionRank orders the outcomes by how much of the reader's attention
// they deserve, so that a record verified by several bindings reports the
// most demanding of them rather than whichever happened to be listed first.
//
// The order is not the exit-code order. ProjectionInvalid ranks above every
// real outcome despite never being reachable through a well-behaved verdict,
// precisely because reaching it means a defect: nothing may mask it, least of
// all the pass of a sibling binding. Violation outranks BlockedUnchecked
// because both gate the run but only one of them names something concrete to
// fix.
//
// The map is total over all six Projection constants, so golangci-lint's
// exhaustive linter (check: map) holds it to the same completeness as the
// vocabulary table in style.go.
var projectionRank = map[Projection]int{
	ProjectionPass:              0,
	ProjectionExpectedUnchecked: 1,
	ProjectionBlockedUnchecked:  2,
	ProjectionViolation:         3,
	ProjectionOperational:       4,
	ProjectionInvalid:           5,
}

// GroupRecords collapses rows into one Record per record file, in the order a
// run is always printed in: by record path, and by binding index within a
// record. Every row belongs to exactly one record, so no row is dropped and
// none is counted twice — exit aggregation keeps reading the ungrouped rows,
// because how loudly a run reports a record is a rendering question and what
// the process exits with is not.
func GroupRecords(rows []Row) []Record {
	byPath := make(map[string][]Row, len(rows))
	for i := range rows {
		path := rows[i].Identity.RecordPath
		byPath[path] = append(byPath[path], rows[i])
	}

	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	records := make([]Record, 0, len(paths))
	for _, path := range paths {
		grouped := byPath[path]
		if len(grouped) == 0 {
			continue
		}
		sort.Slice(grouped, func(i, j int) bool {
			return grouped[i].Identity.BindingIndex < grouped[j].Identity.BindingIndex
		})
		records = append(records, newRecord(path, grouped))
	}
	return records
}

// newRecord folds a record's bindings into the single line and single detail
// entry the record is reported as. rows must not be empty: a record exists in
// the plan only because a binding put it there.
func newRecord(path string, rows []Row) Record {
	rec := Record{Path: path, Rows: rows, Projection: rows[0].Projection}
	for i := range rows {
		r := &rows[i]
		if projectionRank[r.Projection] > projectionRank[rec.Projection] {
			rec.Projection = r.Projection
		}
		rec.RecordID = firstNonEmpty(rec.RecordID, r.RecordID)
		rec.Title = firstNonEmpty(rec.Title, r.Title)
		rec.State = firstNonEmpty(rec.State, r.State)
		rec.Body = firstNonEmpty(rec.Body, r.Body)
		rec.TargetKind = firstNonEmpty(rec.TargetKind, r.TargetKind)
		rec.TargetPath = firstNonEmpty(rec.TargetPath, r.TargetPath)
		rec.Summary = firstNonEmpty(rec.Summary, r.Summary)
	}
	return rec
}

func firstNonEmpty(current, candidate string) string {
	if current != "" {
		return current
	}
	return candidate
}

// dominant reports whether a row carries the outcome the record reports, and
// is therefore one of the rows the record's evidence and detail are read from.
func (rec *Record) dominant(r *Row) bool {
	return r.Projection == rec.Projection
}

// recordDuration totals what the record's executed bindings spent. It reports
// false when nothing in the record executed, which is what leaves the duration
// column blank rather than claiming a measured zero.
func recordDuration(rec *Record) (time.Duration, bool) {
	var total time.Duration
	executed := false
	for i := range rec.Rows {
		if exec := rec.Rows[i].Execution; exec != nil {
			total += exec.Duration
			executed = true
		}
	}
	return total, executed
}

// recordEvidence is the record's right-hand column: what its bindings did, or
// why they did not. Only the bindings carrying the reported outcome are read,
// since a sibling that passed explains nothing about the one that did not,
// and identical evidence from two bindings — the ordinary case for a record
// asserting two things about one file — collapses to a single phrase.
//
// A finding count is a fallback rather than a peer of the other phrases: it
// says how much is wrong where a command line says what to re-run, so it is
// totalled across the bindings that offered nothing better and appended once.
func recordEvidence(rec *Record) string {
	parts := make([]string, 0, len(rec.Rows)+1)
	seen := make(map[string]struct{}, len(rec.Rows))
	findings := 0

	for i := range rec.Rows {
		r := &rec.Rows[i]
		if !rec.dominant(r) {
			continue
		}
		text := bindingEvidence(r)
		if text == "" {
			findings += r.UnwaivedCount
			continue
		}
		if _, duplicate := seen[text]; duplicate {
			continue
		}
		seen[text] = struct{}{}
		parts = append(parts, text)
	}

	if findings > 0 {
		parts = append(parts, findingsPhrase(findings))
	}
	return strings.Join(parts, "; ")
}

func findingsPhrase(n int) string {
	if n == 1 {
		return "1 finding"
	}
	return fmt.Sprintf("%d findings", n)
}
