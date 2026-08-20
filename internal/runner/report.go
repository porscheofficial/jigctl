package runner

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

type RenderOptions struct {
	Out               io.Writer
	Plan              *hcr.Plan
	Verdicts          []*Verdict
	NormalizeDuration bool
}

type bindingInfo struct {
	Executable *hcr.ExecutableBinding
	TargetKind string
	TargetPath string
	Locator    string
}

func relativize(plan *hcr.Plan, recordPath string) string {
	if plan == nil {
		return recordPath
	}
	rel, err := filepath.Rel(plan.Root, recordPath)
	if err != nil {
		return recordPath
	}
	return rel
}

// locate renders a record path relative to the tree root, as cmd/jigctl
// already renders a diagnostic. The binding index is suffixed only for a
// record declaring more than one, since an index that is always :0
// distinguishes nothing and reads as a line number that does not exist.
func locate(plan *hcr.Plan, recordPath string, bindingIndex, bindingsInRecord int) string {
	display := relativize(plan, recordPath)
	if bindingsInRecord > 1 {
		return fmt.Sprintf("%s:%d", display, bindingIndex)
	}
	return display
}

func buildLookup(plan *hcr.Plan) map[BindingIdentity]bindingInfo {
	lookup := make(map[BindingIdentity]bindingInfo)
	if plan == nil {
		return lookup
	}

	perRecord := make(map[string]int)
	for i := range plan.Targets {
		for j := range plan.Targets[i].Bindings {
			perRecord[plan.Targets[i].Bindings[j].RecordPath]++
		}
	}

	for i := range plan.Targets {
		target := &plan.Targets[i]
		for j := range target.Bindings {
			b := &target.Bindings[j]
			lookup[BindingIdentity{
				RecordPath:   b.RecordPath,
				BindingIndex: b.BindingIndex,
			}] = bindingInfo{
				Executable: b,
				TargetKind: target.Kind,
				TargetPath: target.Path,
				Locator:    locate(plan, b.RecordPath, b.BindingIndex, perRecord[b.RecordPath]),
			}
		}
	}
	return lookup
}

func Render(opts RenderOptions) error {
	lookup := buildLookup(opts.Plan)

	sorted := make([]*Verdict, len(opts.Verdicts))
	copy(sorted, opts.Verdicts)
	sort.Slice(sorted, func(i, j int) bool {
		ri := sorted[i].Report()
		rj := sorted[j].Report()
		if ri.Identity.RecordPath != rj.Identity.RecordPath {
			return ri.Identity.RecordPath < rj.Identity.RecordPath
		}
		return ri.Identity.BindingIndex < rj.Identity.BindingIndex
	})

	var counts renderCounts
	filesMap := make(map[string]struct{})

	var knownServicePaths []string
	if opts.Plan != nil {
		for _, t := range opts.Plan.Targets {
			if t.Kind == "service" {
				knownServicePaths = append(knownServicePaths, t.Path)
			}
		}
	}

	for _, v := range sorted {
		if v == nil {
			continue
		}
		renderVerdict(opts, v, lookup, &counts, filesMap, knownServicePaths)
	}

	printSummary(opts.Out, &counts, len(filesMap))
	return nil
}

type renderCounts struct {
	ExpectedUnchecked int
	BlockedUnchecked  int
	Operational       int
	Passes            int
	Violations        int
	Findings          int
}

func renderVerdict(
	opts RenderOptions,
	v *Verdict,
	lookup map[BindingIdentity]bindingInfo,
	counts *renderCounts,
	filesMap map[string]struct{},
	knownServicePaths []string,
) {
	rep := v.Report()
	info, ok := lookup[rep.Identity]

	var exceptions []string
	if ok {
		for _, exc := range info.Executable.Exceptions {
			exceptions = append(exceptions, exc.Scope)
		}
	}

	var mutatedFindings []Finding
	for _, f := range rep.Findings {
		mut, err := ApplyExceptions(f, rep.Kind, exceptions, rep.Identity.RecordPath, knownServicePaths)
		if err == nil {
			mutatedFindings = append(mutatedFindings, mut)
		} else {
			mutatedFindings = append(mutatedFindings, f)
		}
	}

	proj := computeProjection(v, mutatedFindings)
	updateCounts(proj, counts)

	unwaivedCount := 0
	for _, f := range mutatedFindings {
		if len(f.WaivedBy) == 0 {
			unwaivedCount++
			if f.Locus.File != "" {
				filesMap[f.Locus.File] = struct{}{}
			}
		}
	}
	counts.Findings += unwaivedCount

	printLine(opts, v, &rep, info, ok, proj, unwaivedCount)
}

func computeProjection(v *Verdict, mutated []Finding) Projection {
	if v.Completion() == CompletionCompleted {
		for _, f := range mutated {
			if len(f.WaivedBy) == 0 {
				return ProjectionViolation
			}
		}
		return ProjectionPass
	}
	proj, _ := v.Projection()
	return proj
}

func updateCounts(proj Projection, counts *renderCounts) {
	switch proj {
	case ProjectionExpectedUnchecked:
		counts.ExpectedUnchecked++
	case ProjectionBlockedUnchecked:
		counts.BlockedUnchecked++
	case ProjectionOperational:
		counts.Operational++
	case ProjectionPass:
		counts.Passes++
	case ProjectionViolation:
		counts.Violations++
	case ProjectionInvalid:
	}
}

func projectionCode(proj Projection, r Reason) string {
	switch proj {
	case ProjectionPass:
		return "pass"
	case ProjectionViolation:
		return "violation"
	case ProjectionExpectedUnchecked, ProjectionBlockedUnchecked, ProjectionOperational:
		return reasonCode(r)
	case ProjectionInvalid:
		return "unknown"
	}
	return "unknown"
}

// The schema defines summary as the text an agent is shown when the rule
// fires, so it appears on a violation and nowhere else: a pass is carried by
// the evidence appended after this, an unchecked outcome by its reason.
func describe(v *Verdict, info bindingInfo, ok bool, proj Projection) string {
	switch {
	case !ok:
		return fmt.Sprintf("[unknown] %s", reasonMessage(v.Reason()))
	case proj == ProjectionViolation:
		return fmt.Sprintf("[%s] %s", info.Executable.RecordID, info.Executable.Summary)
	case proj == ProjectionPass || proj == ProjectionInvalid:
		return fmt.Sprintf("[%s]", info.Executable.RecordID)
	default:
		return fmt.Sprintf("[%s] %s", info.Executable.RecordID, reasonMessage(v.Reason()))
	}
}

func printLine(
	opts RenderOptions,
	v *Verdict,
	rep *VerdictReport,
	info bindingInfo,
	ok bool,
	proj Projection,
	unwaivedCount int,
) {
	var msgBuilder strings.Builder
	msgBuilder.WriteString(describe(v, info, ok, proj))

	if rep.Kind == "command" && rep.Execution != nil {
		dur := rep.Execution.Duration.String()
		if opts.NormalizeDuration {
			dur = "<duration>"
		}
		fmt.Fprintf(&msgBuilder, " (argv: %s, exit: %d, duration: %s)",
			strings.Join(rep.Execution.Argv, " "), rep.Execution.ExitCode, dur)
	}

	if rep.Kind == "external" && ok {
		fmt.Fprintf(&msgBuilder, " (tool: %s, docs: %s)",
			info.Executable.Tool, info.Executable.Docs)
	}

	if unwaivedCount > 0 {
		fmt.Fprintf(&msgBuilder, " [%d finding(s)]", unwaivedCount)
	}

	locator := info.Locator
	if locator == "" {
		locator = fmt.Sprintf("%s:%d",
			relativize(opts.Plan, rep.Identity.RecordPath), rep.Identity.BindingIndex)
	}

	fmt.Fprintf(opts.Out, "%s: %s: %s\n",
		locator, projectionCode(proj, v.Reason()), msgBuilder.String())
}

func printSummary(out io.Writer, counts *renderCounts, numFiles int) {
	wordFindings := "findings"
	if counts.Findings == 1 {
		wordFindings = "finding"
	}
	wordFiles := "files"
	if numFiles == 1 {
		wordFiles = "file"
	}

	if counts.Findings == 0 {
		fmt.Fprintf(out, "no findings, %d expected-unchecked, %d blocked-unchecked\n",
			counts.ExpectedUnchecked, counts.BlockedUnchecked)
	} else {
		fmt.Fprintf(out, "%d %s in %d %s, %d expected-unchecked, %d blocked-unchecked\n",
			counts.Findings, wordFindings, numFiles, wordFiles,
			counts.ExpectedUnchecked, counts.BlockedUnchecked)
	}
}
