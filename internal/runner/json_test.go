package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

type spyWriter struct {
	writes      int
	written     bytes.Buffer
	errToReturn error
	shortWrite  bool
}

func (s *spyWriter) Write(p []byte) (n int, err error) {
	s.writes++
	if s.errToReturn != nil {
		return 0, s.errToReturn
	}
	if s.shortWrite {
		n = len(p) - 1
		s.written.Write(p[:n])
		return n, nil
	}
	n, err = s.written.Write(p)
	//nolint:wrapcheck // mock writer
	return n, err
}

//nolint:gocyclo // tests for RenderJSON
func TestRenderJSON(t *testing.T) {
	t.Run("zero rows", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows:        []Row{},
			Diagnostics: []hcr.Diagnostic{},
			ExitCode:    77,
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var doc JSONDocument
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		if doc.Command != "run" {
			t.Errorf("expected command: run, got: %v", doc.Command)
		}
		if doc.Records == nil || len(doc.Records) != 0 {
			t.Errorf("expected records: [], got: %v", doc.Records)
		}
		if doc.Diagnostics == nil || len(doc.Diagnostics) != 0 {
			t.Errorf("expected diagnostics: [], got: %v", doc.Diagnostics)
		}
		if doc.ExitCode != 77 {
			t.Errorf("expected exit_code 77, got %d", doc.ExitCode)
		}
	})

	t.Run("exit code trust", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows:        []Row{},
			Diagnostics: []hcr.Diagnostic{},
			ExitCode:    1,
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var doc JSONDocument
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		if doc.ExitCode != 1 {
			t.Errorf("expected exit_code 1, got %d", doc.ExitCode)
		}
	})

	t.Run("nil slice normalization", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows: []Row{
				{
					Identity:   BindingIdentity{RecordPath: "a.md", BindingIndex: 0},
					TargetKind: "repo",
					TargetPath: "", // to test "." normalization
					Findings:   nil,
				},
			},
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := spy.written.String()
		if !bytes.Contains(spy.written.Bytes(), []byte(`"findings": []`)) {
			t.Errorf("expected findings to be [], output: %s", out)
		}

		var doc JSONDocument
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if doc.Records[0].Target.Path != "." {
			t.Errorf("expected Target.Path to be normalized to '.', got %q", doc.Records[0].Target.Path)
		}
	})

	t.Run("findings count assertion", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows: []Row{
				{
					Identity: BindingIdentity{RecordPath: "a.md", BindingIndex: 0},
					Findings: []Finding{
						{WaivedBy: nil},
						{WaivedBy: []ExceptionIdentity{{}}},
					},
					UnwaivedCount: 1,
					WaivedCount:   1,
				},
			},
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var doc JSONDocument
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		for _, rec := range doc.Records {
			for _, b := range rec.Bindings {
				total := b.WaivedCount + b.UnwaivedCount
				if total != len(b.Findings) {
					t.Errorf("expected waived + unwaived (%d) to equal findings len (%d)", total, len(b.Findings))
				}
			}
		}
	})

	t.Run("unresolved record metadata", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows: []Row{
				{
					Identity:  BindingIdentity{RecordPath: "a.md", BindingIndex: 0},
					IsUnknown: true,
				},
			},
		}
		err := RenderJSON(spy, opts)
		if err == nil {
			t.Fatal("expected error for unresolved record metadata, got nil")
		}
	})

	t.Run("one write call for multiple records", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			Rows: []Row{
				{Identity: BindingIdentity{RecordPath: "a.md", BindingIndex: 0}},
				{Identity: BindingIdentity{RecordPath: "b.md", BindingIndex: 0}},
			},
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spy.writes != 1 {
			t.Errorf("expected exactly 1 write call, got %d", spy.writes)
		}
	})

	t.Run("writer error propagates", func(t *testing.T) {
		spy := &spyWriter{errToReturn: errors.New("disk full")}
		opts := JSONOptions{Rows: []Row{}}
		err := RenderJSON(spy, opts)
		if err == nil || err.Error() != "write doc: disk full" {
			t.Errorf("expected 'write doc: disk full', got %v", err)
		}
	})

	t.Run("short write error", func(t *testing.T) {
		spy := &spyWriter{shortWrite: true}
		opts := JSONOptions{Rows: []Row{}}
		err := RenderJSON(spy, opts)
		if err == nil {
			t.Fatal("expected error for short write, got nil")
		}
	})

	t.Run("marshalAndWriteJSON failure", func(t *testing.T) {
		spy := &spyWriter{}
		unsupportedDoc := struct{ Ch chan int }{make(chan int)}
		err := marshalAndWriteJSON(spy, unsupportedDoc)
		if err == nil {
			t.Fatal("expected error when marshaling unsupported value")
		}
		if spy.writes != 0 {
			t.Errorf("expected 0 writes on marshal failure, got %d", spy.writes)
		}
	})

	t.Run("duration normalization and rounding", func(t *testing.T) {
		spy := &spyWriter{}
		opts := JSONOptions{
			NormalizeDuration: true,
			Rows: []Row{
				{
					Identity: BindingIdentity{RecordPath: "a.md", BindingIndex: 0},
					Execution: &Execution{
						Duration: 1500 * time.Microsecond, // 1.5ms
						Argv:     []string{"test"},        // Need this for execution to be rendered
					},
				},
			},
		}
		err := RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var doc JSONDocument
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if doc.Records[0].Bindings[0].Execution.DurationMs != 0 {
			t.Errorf("expected duration_ms to be 0 with NormalizeDuration, got %d", doc.Records[0].Bindings[0].Execution.DurationMs)
		}

		opts.NormalizeDuration = false
		spy = &spyWriter{}
		err = RenderJSON(spy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err = json.Unmarshal(spy.written.Bytes(), &doc)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if doc.Records[0].Bindings[0].Execution.DurationMs != 2 {
			t.Errorf("expected duration_ms to round to 2, got %d", doc.Records[0].Bindings[0].Execution.DurationMs)
		}
	})
}
