package runner

// JSONDiagnostic represents a single operational diagnostic emitted during a run.
type JSONDiagnostic struct {
	File    string `json:"file"`
	Pointer string `json:"pointer"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSONSummary gives an overview of the run's outcomes.
type JSONSummary struct {
	Records                   int            `json:"records"`
	Bindings                  int            `json:"bindings"`
	BindingsByProjection      map[string]int `json:"bindings_by_projection"`
	UnwaivedFindings          int            `json:"unwaived_findings"`
	FilesWithUnwaivedFindings int            `json:"files_with_unwaived_findings"`
}

// JSONTarget describes what a record applies to.
type JSONTarget struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

// JSONExecution describes what happened when a child process was invoked.
type JSONExecution struct {
	Argv       []string `json:"argv"`
	ExitCode   int      `json:"exit_code"`
	DurationMs int      `json:"duration_ms"`
}

// JSONWaivedBy describes which exception waived a finding.
type JSONWaivedBy struct {
	RecordPath     string `json:"record_path"`
	ExceptionIndex int    `json:"exception_index"`
}

// JSONFinding is a single violation reported by a binding.
type JSONFinding struct {
	File     string         `json:"file"`
	Pointer  string         `json:"pointer"`
	Severity string         `json:"severity"`
	Partial  bool           `json:"partial"`
	WaivedBy []JSONWaivedBy `json:"waived_by"`
}

// JSONBinding represents one binding within a record.
type JSONBinding struct {
	Index         int            `json:"index"`
	Kind          string         `json:"kind"`
	Severity      string         `json:"severity"`
	Projection    string         `json:"projection"`
	Reason        *string        `json:"reason"`
	Tool          string         `json:"tool"`
	Docs          string         `json:"docs"`
	Execution     *JSONExecution `json:"execution"`
	Findings      []JSONFinding  `json:"findings"`
	WaivedCount   int            `json:"waived_count"`
	UnwaivedCount int            `json:"unwaived_count"`
}

// JSONRecord represents a single evaluated record.
type JSONRecord struct {
	RecordID   string        `json:"record_id"`
	Path       string        `json:"path"`
	Title      string        `json:"title"`
	State      string        `json:"state"`
	Summary    string        `json:"summary"`
	Body       string        `json:"body"`
	Projection string        `json:"projection"`
	Target     JSONTarget    `json:"target"`
	Bindings   []JSONBinding `json:"bindings"`
}

// JSONDocument is the root of the emitted JSON output.
type JSONDocument struct {
	SchemaVersion int              `json:"schema_version"`
	Command       string           `json:"command"`
	Root          string           `json:"root"`
	ExitCode      int              `json:"exit_code"`
	Diagnostics   []JSONDiagnostic `json:"diagnostics"`
	Summary       JSONSummary      `json:"summary"`
	Records       []JSONRecord     `json:"records"`
}

// jsonProjectionCode maps the internal Projection outcome back to the stable string
// identifiers used in JSON output. It must be exhaustive over all Projection constants.
var jsonProjectionCode = map[Projection]string{
	ProjectionPass:              "pass",
	ProjectionViolation:         "violation",
	ProjectionExpectedUnchecked: "expected-unchecked",
	ProjectionBlockedUnchecked:  "blocked-unchecked",
	ProjectionOperational:       "operational",
	ProjectionInvalid:           "invalid",
}

// jsonFailureProjection specifies whether a record with this projection should be
// kept when --only-failures is requested. It must be exhaustive.
var jsonFailureProjection = map[Projection]bool{
	ProjectionPass:              false,
	ProjectionExpectedUnchecked: false,
	ProjectionBlockedUnchecked:  true,
	ProjectionOperational:       true,
	ProjectionInvalid:           true,
	ProjectionViolation:         true,
}

var jsonFailureProjectionString map[string]bool

func init() {
	jsonFailureProjectionString = make(map[string]bool, len(jsonFailureProjection))
	for p, keep := range jsonFailureProjection {
		jsonFailureProjectionString[jsonProjectionCode[p]] = keep
	}
}
