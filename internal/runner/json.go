package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

// JSONOptions defines the parameters for JSON output rendering.
type JSONOptions struct {
	Root              string
	Rows              []Row
	Diagnostics       []hcr.Diagnostic
	ExitCode          int
	NormalizeDuration bool
	OnlyFailures      bool
}

// marshalAndWriteJSON formats the document as indented JSON and writes it to out
// in exactly one Write call to fulfill the channel contract.
func marshalAndWriteJSON(out io.Writer, doc any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	n, err := out.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("write doc: %w", err)
	}
	if n < buf.Len() {
		return fmt.Errorf("short write: wrote %d of %d bytes", n, buf.Len())
	}
	return nil
}

// RenderJSON renders a given slice of evaluated Rows as a JSON document.
//
//nolint:gocritic // matching expected API signature
func RenderJSON(out io.Writer, opts JSONOptions) error {
	doc, err := buildJSONDocument(opts)
	if err != nil {
		return err
	}
	return marshalAndWriteJSON(out, doc)
}

//nolint:funlen,gocritic // building JSON document is naturally long, and passes opts by value
func buildJSONDocument(opts JSONOptions) (JSONDocument, error) {
	doc := JSONDocument{
		SchemaVersion: 1,
		Command:       "run",
		Root:          opts.Root,
		Diagnostics:   make([]JSONDiagnostic, 0, len(opts.Diagnostics)),
		Records:       make([]JSONRecord, 0),
		ExitCode:      opts.ExitCode,
	}

	for i := range opts.Diagnostics {
		d := &opts.Diagnostics[i]
		doc.Diagnostics = append(doc.Diagnostics, JSONDiagnostic{
			File:    d.File,
			Pointer: d.Pointer,
			Code:    d.Code,
			Message: d.Message,
		})
	}

	doc.Summary.BindingsByProjection = make(map[string]int)
	for _, k := range jsonProjectionCode {
		doc.Summary.BindingsByProjection[k] = 0
	}

	grouped := GroupRecords(opts.Rows)
	for i := range grouped {
		rec := &grouped[i]
		jRec := JSONRecord{
			RecordID: rec.RecordID,
			Path:     rec.Path,
			Title:    rec.Title,
			State:    rec.State,
			Summary:  rec.Summary,
			Body:     rec.Body,
			Target: JSONTarget{
				Kind: rec.TargetKind,
				Path: rec.TargetPath,
			},
			Projection: jsonProjectionCode[rec.Projection],
			Bindings:   make([]JSONBinding, 0, len(rec.Rows)),
		}

		if jRec.Target.Path == "" && jRec.Target.Kind == "repo" {
			jRec.Target.Path = "."
		}

		for j := range rec.Rows {
			row := &rec.Rows[j]
			if row.IsUnknown {
				return doc, fmt.Errorf("row for record %s binding %d has no resolved metadata",
					row.Identity.RecordPath, row.Identity.BindingIndex)
			}

			doc.Summary.Bindings++
			doc.Summary.BindingsByProjection[jsonProjectionCode[row.Projection]]++

			jb := JSONBinding{
				Index:         row.Identity.BindingIndex,
				Kind:          row.Kind,
				Severity:      row.Severity,
				Projection:    jsonProjectionCode[row.Projection],
				Tool:          row.Tool,
				Docs:          row.Docs,
				WaivedCount:   row.WaivedCount,
				UnwaivedCount: row.UnwaivedCount,
				Findings:      make([]JSONFinding, 0, len(row.Findings)),
			}

			if row.Reason != 0 {
				rCode := reasonCode(row.Reason)
				jb.Reason = &rCode
			}

			if row.Execution != nil {
				jb.Execution = buildJSONExecution(row, opts.NormalizeDuration)
			}

			for k := range row.Findings {
				f := &row.Findings[k]
				jf := JSONFinding{
					File:     f.Locus.File,
					Pointer:  f.Locus.Pointer,
					Severity: f.Severity,
					Partial:  f.Partial,
					WaivedBy: make([]JSONWaivedBy, 0, len(f.WaivedBy)),
				}

				for l := range f.WaivedBy {
					jf.WaivedBy = append(jf.WaivedBy, JSONWaivedBy(f.WaivedBy[l]))
				}
				jb.Findings = append(jb.Findings, jf)
			}
			jRec.Bindings = append(jRec.Bindings, jb)
		}

		doc.Records = append(doc.Records, jRec)
	}

	doc.Summary.Records = len(doc.Records)

	unwaivedSum := 0
	for i := range opts.Rows {
		unwaivedSum += opts.Rows[i].UnwaivedCount
	}
	doc.Summary.UnwaivedFindings = unwaivedSum
	doc.Summary.FilesWithUnwaivedFindings = UnwaivedFileCount(opts.Rows)

	if opts.OnlyFailures {
		doc.Records = filterFailureRecords(doc.Records)
	}

	return doc, nil
}

// filterFailureRecords returns only the records whose projection is actionable,
// per jsonFailureProjection. The returned slice is never nil.
func filterFailureRecords(records []JSONRecord) []JSONRecord {
	filtered := make([]JSONRecord, 0)
	for i := range records {
		if jsonFailureProjectionString[records[i].Projection] {
			filtered = append(filtered, records[i])
		}
	}
	return filtered
}

func buildJSONExecution(row *Row, normalize bool) *JSONExecution {
	exec := JSONExecution{
		Argv:     row.Execution.Argv,
		ExitCode: row.Execution.ExitCode,
	}
	if exec.Argv == nil {
		exec.Argv = []string{}
	}
	if normalize {
		exec.DurationMs = 0
	} else {
		exec.DurationMs = int(row.Execution.Duration.Round(time.Millisecond) / time.Millisecond)
	}
	return &exec
}
