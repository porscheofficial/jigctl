package runner

import "time"

// Completion is the evaluation axis of a verdict.
type Completion uint8

const (
	CompletionUnset Completion = iota
	CompletionCompleted
	CompletionBlocked
	CompletionNotAttempted
	CompletionOperational
)

// Projection is the reported view derived from completion and findings.
type Projection uint8

const (
	ProjectionInvalid Projection = iota
	ProjectionPass
	ProjectionViolation
	ProjectionExpectedUnchecked
	ProjectionBlockedUnchecked
	ProjectionOperational
)

// Reason identifies why a binding did not complete normally.
type Reason uint8

const (
	ReasonNone Reason = iota
	reasonExecutableMissing
	reasonExecutableDenied
	reasonTimeout
	reasonAuthorizationDenied
	reasonGlobNoMatches
	reasonInputMissing
	reasonInputUnreadable
	reasonInputMalformed
	reasonPointerMalformed
	reasonPatternInvalid
	reasonScopeInvalid
	reasonGlobInvalid
	reasonArgvInvalid
	reasonFormatUnsupported
	reasonModifierUnimplemented
	reasonKindNotExecutable
	reasonCadenceExcluded
	reasonRecordDraft
	reasonRecordDeprecated
	reasonProcessStart
	reasonPathEscapesRoot
	reasonLimitExceeded
	reasonInvocationCancelled
)

// BlockedReason is a closed subset accepted only by blocked verdicts.
type BlockedReason Reason

const (
	ReasonExecutableMissing     BlockedReason = BlockedReason(reasonExecutableMissing)
	ReasonExecutableDenied      BlockedReason = BlockedReason(reasonExecutableDenied)
	ReasonTimeout               BlockedReason = BlockedReason(reasonTimeout)
	ReasonAuthorizationDenied   BlockedReason = BlockedReason(reasonAuthorizationDenied)
	ReasonGlobNoMatches         BlockedReason = BlockedReason(reasonGlobNoMatches)
	ReasonInputMissing          BlockedReason = BlockedReason(reasonInputMissing)
	ReasonInputUnreadable       BlockedReason = BlockedReason(reasonInputUnreadable)
	ReasonInputMalformed        BlockedReason = BlockedReason(reasonInputMalformed)
	ReasonPointerMalformed      BlockedReason = BlockedReason(reasonPointerMalformed)
	ReasonPatternInvalid        BlockedReason = BlockedReason(reasonPatternInvalid)
	ReasonScopeInvalid          BlockedReason = BlockedReason(reasonScopeInvalid)
	ReasonGlobInvalid           BlockedReason = BlockedReason(reasonGlobInvalid)
	ReasonArgvInvalid           BlockedReason = BlockedReason(reasonArgvInvalid)
	ReasonFormatUnsupported     BlockedReason = BlockedReason(reasonFormatUnsupported)
	ReasonModifierUnimplemented BlockedReason = BlockedReason(reasonModifierUnimplemented)
)

// ExpectedReason is a closed subset accepted only by not-attempted verdicts.
type ExpectedReason Reason

const (
	ReasonKindNotExecutable ExpectedReason = ExpectedReason(reasonKindNotExecutable)
	ReasonCadenceExcluded   ExpectedReason = ExpectedReason(reasonCadenceExcluded)
	ReasonRecordDraft       ExpectedReason = ExpectedReason(reasonRecordDraft)
	ReasonRecordDeprecated  ExpectedReason = ExpectedReason(reasonRecordDeprecated)
)

// OperationalReason is a closed subset accepted only by operational verdicts.
type OperationalReason Reason

const (
	ReasonProcessStart        OperationalReason = OperationalReason(reasonProcessStart)
	ReasonPathEscapesRoot     OperationalReason = OperationalReason(reasonPathEscapesRoot)
	ReasonLimitExceeded       OperationalReason = OperationalReason(reasonLimitExceeded)
	ReasonInvocationCancelled OperationalReason = OperationalReason(reasonInvocationCancelled)
)

// BindingIdentity is the invocation-unique verdict key.
type BindingIdentity struct {
	RecordPath   string
	BindingIndex int
}

// TargetProvenance records applicability without becoming part of BindingIdentity.
type TargetProvenance struct {
	Name string
	Path string
}

// ExceptionIdentity identifies one authored exception.
type ExceptionIdentity struct {
	RecordPath     string
	ExceptionIndex int
}

// Locus identifies where a finding occurred. Pointer is empty outside config assertions.
type Locus struct {
	File    string
	Pointer string
}

// Finding is evaluation data, including every exception that suppressed it.
type Finding struct {
	Locus    Locus
	Severity string
	WaivedBy []ExceptionIdentity
	Partial  bool
}

// TimeoutRecord preserves the declared, resolved, and shared effective timeout.
// Nil Declared means the default was used; nil Effective means no execution was eligible.
type TimeoutRecord struct {
	Declared  *time.Duration
	Resolved  time.Duration
	Effective *time.Duration
}

// Execution records the process observation for an executed command binding.
type Execution struct {
	Argv     []string
	ExitCode int
	Duration time.Duration
}

// VerdictReport contains the actionable data shared by every outcome.
type VerdictReport struct {
	Identity  BindingIdentity
	Target    TargetProvenance
	Kind      string
	Severity  string
	Timeouts  TimeoutRecord
	Execution *Execution
	Findings  []Finding
}

// Verdict is outcome data, not a Go error. Its state and reason are private so
// callers must use a state-specific constructor. Projection structurally gives
// pass only to completed verdicts; callers remain responsible for supplying the
// truthful reason and marking findings partial when evaluation stopped early.
type Verdict struct {
	report     VerdictReport
	completion Completion
	reason     Reason
}

// NewCompletedVerdict records evaluation of the whole declared subject.
func NewCompletedVerdict(report *VerdictReport) *Verdict {
	return &Verdict{report: *report, completion: CompletionCompleted}
}

// NewBlockedVerdict records attempted evaluation that could not finish.
func NewBlockedVerdict(report *VerdictReport, reason BlockedReason) *Verdict {
	return &Verdict{report: *report, completion: CompletionBlocked, reason: Reason(reason)}
}

// NewNotAttemptedVerdict records deliberate non-execution.
func NewNotAttemptedVerdict(report *VerdictReport, reason ExpectedReason) *Verdict {
	return &Verdict{report: *report, completion: CompletionNotAttempted, reason: Reason(reason)}
}

// NewOperationalVerdict records failure of jigctl or its invocation.
func NewOperationalVerdict(report *VerdictReport, reason OperationalReason) *Verdict {
	return &Verdict{report: *report, completion: CompletionOperational, reason: Reason(reason)}
}

// Completion returns the evaluation axis.
func (verdict *Verdict) Completion() Completion {
	if verdict == nil {
		return CompletionUnset
	}
	return verdict.completion
}

// Reason returns why evaluation did not complete normally.
func (verdict *Verdict) Reason() Reason {
	if verdict == nil {
		return ReasonNone
	}
	return verdict.reason
}

// Report returns the actionable outcome data.
func (verdict *Verdict) Report() VerdictReport {
	if verdict == nil {
		return VerdictReport{}
	}
	return verdict.report
}

// Projection derives the report state. False marks an unset or unknown completion.
func (verdict *Verdict) Projection() (Projection, bool) {
	if verdict == nil {
		return ProjectionInvalid, false
	}
	switch verdict.completion {
	case CompletionUnset:
		return ProjectionInvalid, false
	case CompletionCompleted:
		for _, finding := range verdict.report.Findings {
			if len(finding.WaivedBy) == 0 {
				return ProjectionViolation, true
			}
		}
		return ProjectionPass, true
	case CompletionBlocked:
		return ProjectionBlockedUnchecked, true
	case CompletionNotAttempted:
		return ProjectionExpectedUnchecked, true
	case CompletionOperational:
		return ProjectionOperational, true
	}
	return ProjectionInvalid, false
}
