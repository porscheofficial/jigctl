package runner

import (
	"fmt"
	"path/filepath"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

// Row represents the fully evaluated outcome of a single binding, ready for rendering or exit aggregation.
type Row struct {
	Identity BindingIdentity
	Locator  string
	RecordID string
	// Not-in-lookup case yields empty title as there is no executable binding to read from.
	Title string
	Kind  string
	// State is the record's authored lifecycle position. Every binding of
	// one record carries the same value, since state is a record-level
	// field, which is what lets a grouped record render one state glyph.
	State         string
	Body          string
	TargetKind    string
	TargetPath    string
	Severity      string
	Summary       string
	Tool          string
	Docs          string
	Projection    Projection
	Reason        Reason
	Execution     *Execution
	Findings      []Finding
	WaivedCount   int
	UnwaivedCount int
	IsUnknown     bool
}

func relativizePath(plan *hcr.Plan, recordPath string) string {
	if plan == nil {
		return recordPath
	}
	rel, err := filepath.Rel(plan.Root, recordPath)
	if err != nil {
		return recordPath
	}
	return rel
}

func locateBinding(plan *hcr.Plan, recordPath string, bindingIndex, bindingsInRecord int) string {
	display := relativizePath(plan, recordPath)
	if bindingsInRecord > 1 {
		return fmt.Sprintf("%s:%d", display, bindingIndex)
	}
	return display
}

func deriveProjection(v *Verdict, mutatedFindings []Finding) Projection {
	if v.Completion() == CompletionCompleted {
		return projectionFromFindings(mutatedFindings)
	}
	proj, _ := v.Projection()
	return proj
}

func getExceptions(b *hcr.ExecutableBinding) []string {
	var exceptions []string
	if b != nil {
		for _, exc := range b.Exceptions {
			exceptions = append(exceptions, exc.Scope)
		}
	}
	return exceptions
}

func applyAllExceptions(
	findings []Finding,
	kind string,
	exceptions []string,
	recordPath string,
	knownServicePaths []string,
) []Finding {
	var mutated []Finding
	for _, f := range findings {
		mut, err := ApplyExceptions(f, kind, exceptions, recordPath, knownServicePaths)
		if err == nil {
			mutated = append(mutated, mut)
		} else {
			mutated = append(mutated, f)
		}
	}
	return mutated
}

func buildLookupAndServicePaths(plan *hcr.Plan) (
	lookup map[BindingIdentity]*hcr.ExecutableBinding,
	perRecord map[string]int,
	knownServicePaths []string,
) {
	perRecord = make(map[string]int)
	lookup = make(map[BindingIdentity]*hcr.ExecutableBinding)

	if plan != nil {
		for i := range plan.Targets {
			t := &plan.Targets[i]
			if t.Kind == "service" {
				knownServicePaths = append(knownServicePaths, t.Path)
			}
			for j := range t.Bindings {
				b := &t.Bindings[j]
				perRecord[b.RecordPath]++
				id := BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex}
				lookup[id] = b
			}
		}
	}
	return lookup, perRecord, knownServicePaths
}

// RowBuilder converts verdicts into rows one at a time, holding the
// per-plan lookups that conversion needs. A run builds it once and then
// either drains a whole verdict slice through BuildRows or feeds it single
// verdicts as they settle, which is what the live view does.
type RowBuilder struct {
	plan              *hcr.Plan
	lookup            map[BindingIdentity]*hcr.ExecutableBinding
	perRecord         map[string]int
	knownServicePaths []string
}

// NewRowBuilder indexes a plan so verdicts can be converted to rows.
func NewRowBuilder(plan *hcr.Plan) *RowBuilder {
	lookup, perRecord, knownServicePaths := buildLookupAndServicePaths(plan)
	return &RowBuilder{
		plan:              plan,
		lookup:            lookup,
		perRecord:         perRecord,
		knownServicePaths: knownServicePaths,
	}
}

// Row computes the post-exception findings and the final Projection for one
// verdict.
func (b *RowBuilder) Row(v *Verdict) Row {
	return processVerdict(b.plan, v, b.lookup, b.perRecord, b.knownServicePaths)
}

// BuildRows computes post-exception findings and the final Projection ONCE per verdict.
func BuildRows(plan *hcr.Plan, verdicts []*Verdict) []Row {
	builder := NewRowBuilder(plan)

	rows := make([]Row, 0, len(verdicts))
	for _, v := range verdicts {
		if v == nil {
			continue
		}

		rows = append(rows, builder.Row(v))
	}

	return rows
}

func getLocator(
	plan *hcr.Plan,
	rep *VerdictReport,
	ok bool,
	perRecord map[string]int,
) string {
	if ok {
		return locateBinding(
			plan,
			rep.Identity.RecordPath,
			rep.Identity.BindingIndex,
			perRecord[rep.Identity.RecordPath],
		)
	}
	return fmt.Sprintf("%s:%d", relativizePath(plan, rep.Identity.RecordPath), rep.Identity.BindingIndex)
}

func processVerdict(
	plan *hcr.Plan,
	v *Verdict,
	lookup map[BindingIdentity]*hcr.ExecutableBinding,
	perRecord map[string]int,
	knownServicePaths []string,
) Row {
	rep := v.Report()
	b, ok := lookup[rep.Identity]
	exceptions := getExceptions(b)

	mutatedFindings := applyAllExceptions(
		rep.Findings,
		rep.Kind,
		exceptions,
		rep.Identity.RecordPath,
		knownServicePaths,
	)

	row := Row{
		Identity:   rep.Identity,
		Locator:    getLocator(plan, &rep, ok, perRecord),
		Kind:       rep.Kind,
		TargetKind: rep.Target.Name,
		TargetPath: rep.Target.Path,
		Severity:   rep.Severity,
		Projection: deriveProjection(v, mutatedFindings),
		Reason:     v.Reason(),
		Execution:  rep.Execution,
		Findings:   mutatedFindings,
		IsUnknown:  !ok,
	}

	if ok {
		row.RecordID = b.RecordID
		row.Summary = b.Summary
		row.Tool = b.Tool
		row.Docs = b.Docs
		row.Title = b.Title
		row.State = b.State
		row.Body = b.Body
	}

	for i := range mutatedFindings {
		if len(mutatedFindings[i].WaivedBy) == 0 {
			row.UnwaivedCount++
		} else {
			row.WaivedCount++
		}
	}

	return row
}

// ExitSummaries converts rows to the format AggregateExitCode consumes.
func ExitSummaries(rows []Row) []ExitSummary {
	summaries := make([]ExitSummary, 0, len(rows))
	for i := range rows {
		summaries = append(summaries, ExitSummary{
			Projection: rows[i].Projection,
			IsBlocking: rows[i].Severity != "advisory",
		})
	}
	return summaries
}

// UnwaivedFileCount returns the number of unique files involved in unwaived findings.
func UnwaivedFileCount(rows []Row) int {
	filesMap := make(map[string]struct{})
	for i := range rows {
		for _, f := range rows[i].Findings {
			if len(f.WaivedBy) == 0 && f.Locus.File != "" {
				filesMap[f.Locus.File] = struct{}{}
			}
		}
	}
	return len(filesMap)
}
