package runner

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
