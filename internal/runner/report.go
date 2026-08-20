package runner

import (
	"fmt"
	"io"
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
}

func buildLookup(plan *hcr.Plan) map[BindingIdentity]bindingInfo {
	lookup := make(map[BindingIdentity]bindingInfo)
	if plan == nil {
		return lookup
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

func printLine(
	opts RenderOptions,
	v *Verdict,
	rep *VerdictReport,
	info bindingInfo,
	ok bool,
	proj Projection,
	unwaivedCount int,
) {
	var code string
	switch proj {
	case ProjectionPass:
		code = "pass"
	case ProjectionViolation:
		code = "violation"
	case ProjectionExpectedUnchecked, ProjectionBlockedUnchecked, ProjectionOperational:
		code = reasonCode(v.Reason())
	case ProjectionInvalid:
		code = "unknown"
	}

	var msgBuilder strings.Builder
	if ok {
		fmt.Fprintf(&msgBuilder, "[%s] %s", info.Executable.RecordID, info.Executable.Summary)
	} else {
		fmt.Fprintf(&msgBuilder, "[unknown] %s", reasonMessage(v.Reason()))
	}

	if proj == ProjectionExpectedUnchecked || proj == ProjectionBlockedUnchecked || proj == ProjectionOperational {
		fmt.Fprintf(&msgBuilder, " (%s)", reasonMessage(v.Reason()))
	}

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

	fmt.Fprintf(opts.Out, "%s:%d: %s: %s\n",
		rep.Identity.RecordPath, rep.Identity.BindingIndex, code, msgBuilder.String())
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

var reasonData = map[Reason]struct{ Code, Message string }{
	ReasonNone:                          {"none", "OK"},
	Reason(ReasonExecutableMissing):     {"executable-missing", "Executable is absent from PATH"},
	Reason(ReasonExecutableDenied):      {"executable-denied", "Executable permission is denied"},
	Reason(ReasonTimeout):               {"timeout", "Timeout expires"},
	Reason(ReasonAuthorizationDenied):   {"authorization-denied", "Execution authorization is absent"},
	Reason(ReasonGlobNoMatches):         {"glob-no-matches", "Grep glob matches no files"},
	Reason(ReasonInputMissing):          {"input-missing", "Configuration data file is missing"},
	Reason(ReasonInputUnreadable):       {"input-unreadable", "Data is unreadable"},
	Reason(ReasonInputMalformed):        {"input-malformed", "Data is malformed"},
	Reason(ReasonPointerMalformed):      {"pointer-malformed", "RFC 6901 pointer is malformed"},
	Reason(ReasonPatternInvalid):        {"pattern-invalid", "matches cannot compile"},
	Reason(ReasonScopeInvalid):          {"scope-invalid", "Scope has no valid shape"},
	Reason(ReasonGlobInvalid):           {"glob-invalid", "Grep glob syntax is invalid"},
	Reason(ReasonArgvInvalid):           {"argv-invalid", "Command argv cannot be split"},
	Reason(ReasonFormatUnsupported):     {"format-unsupported", "Format is unsupported"},
	Reason(ReasonModifierUnimplemented): {"modifier-unimplemented", "Binding has pattern or select"},
	Reason(ReasonKindNotExecutable):     {"kind-not-executable", "Kind cannot execute"},
	Reason(ReasonCadenceExcluded):       {"cadence-excluded", "Cadence excluded"},
	Reason(ReasonRecordDraft):           {"record-draft", "Record is draft"},
	Reason(ReasonRecordDeprecated):      {"record-deprecated", "Record is deprecated"},
	Reason(ReasonProcessStart):          {"process-start", "Other process-start failure"},
	Reason(ReasonPathEscapesRoot):       {"path-escapes-root", "Path escapes root"},
	Reason(ReasonLimitExceeded):         {"limit-exceeded", "Output or read limit exceeded"},
	Reason(ReasonInvocationCancelled):   {"invocation-cancelled", "Invocation is cancelled"},
}

func reasonCode(r Reason) string {
	if data, ok := reasonData[r]; ok {
		return data.Code
	}
	return "unknown"
}

func reasonMessage(r Reason) string {
	if data, ok := reasonData[r]; ok {
		return data.Message
	}
	return "Unknown reason"
}
