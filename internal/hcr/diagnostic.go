package hcr

// Diagnostic is validation data, not an operational error. The errors.Join
// and %w rules apply only to the error return used for operational failures.
type Diagnostic struct {
	File    string
	Pointer string
	Code    string
	Message string
}
