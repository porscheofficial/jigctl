package runner

// reasonData is the vocabulary of every reason a binding did not complete
// normally, in the wording `--plain` emits. Code is the stable identifier for
// grep and awk; Message is the sentence beside it.
//
// Neither moves. A script pinned to `--plain` is matching these bytes and must
// not have them reworded underneath it (ADR-0013). The default output speaks
// to a human and has its own vocabulary in reasonPhrases.
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

// reasonPhrases is the same vocabulary written for the person watching the
// run. Where there is something to do about a reason it says that instead of
// naming the condition, so `authorization-denied` reads as `needs
// --allow-exec`, and where the state column has already said it the phrase is
// empty rather than redundant.
var reasonPhrases = map[Reason]string{
	ReasonNone:                          "",
	Reason(ReasonExecutableMissing):     "executable is not on PATH",
	Reason(ReasonExecutableDenied):      "executable is not permitted to run",
	Reason(ReasonTimeout):               "timed out",
	Reason(ReasonAuthorizationDenied):   "needs --allow-exec",
	Reason(ReasonGlobNoMatches):         "glob matched no files",
	Reason(ReasonInputMissing):          "file is missing",
	Reason(ReasonInputUnreadable):       "file could not be read",
	Reason(ReasonInputMalformed):        "file could not be parsed",
	Reason(ReasonPointerMalformed):      "pointer is malformed",
	Reason(ReasonPatternInvalid):        "pattern does not compile",
	Reason(ReasonScopeInvalid):          "exception scope is malformed",
	Reason(ReasonGlobInvalid):           "glob syntax is invalid",
	Reason(ReasonArgvInvalid):           "command line could not be split",
	Reason(ReasonFormatUnsupported):     "file format is unsupported",
	Reason(ReasonModifierUnimplemented): "pattern and select are not implemented yet",
	Reason(ReasonKindNotExecutable):     "nothing here for jigctl to run",
	Reason(ReasonCadenceExcluded):       "not due on this invocation",
	Reason(ReasonRecordDraft):           "not enforced yet",
	Reason(ReasonRecordDeprecated):      "kept for history",
	Reason(ReasonProcessStart):          "process could not be started",
	Reason(ReasonPathEscapesRoot):       "path escapes the repository root",
	Reason(ReasonLimitExceeded):         "output or read limit exceeded",
	Reason(ReasonInvocationCancelled):   "run was cancelled",
}

// unexecutableKinds answers the question `kind-not-executable` raises and
// does not settle: not that jigctl declined to run something, but that there
// was never anything for it to run, and which of three different reasons that
// is. A reader who knows the six kinds can infer it; the output is written for
// the reader who does not.
var unexecutableKinds = map[string]string{
	"external":     "checked outside jigctl",
	"agent-review": "reviewed by an agent, not by jigctl",
	"inferential":  "a human judgement, nothing to run",
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
	return "unknown reason"
}

func reasonPhrase(r Reason) string {
	if phrase, ok := reasonPhrases[r]; ok {
		return phrase
	}
	return "unknown reason"
}

// unexecutableKind names why a binding of this kind had nothing to execute.
// An external binding also names the tool that does the checking and where it
// is documented, since that is the whole of what jigctl knows about a check
// happening somewhere else.
func unexecutableKind(kind, tool, docs string) string {
	phrase, ok := unexecutableKinds[kind]
	if !ok {
		return reasonPhrase(Reason(ReasonKindNotExecutable))
	}
	if kind != "external" {
		return phrase
	}
	if detail := externalEvidence(tool, docs); detail != "" {
		return phrase + " — " + detail
	}
	return phrase
}

func externalEvidence(tool, docs string) string {
	switch {
	case tool != "" && docs != "":
		return "tool: " + tool + ", docs: " + docs
	case tool != "":
		return "tool: " + tool
	case docs != "":
		return "docs: " + docs
	}
	return ""
}
