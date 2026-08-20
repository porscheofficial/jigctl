package runner

import (
	"reflect"
	"testing"
)

func TestApplyExceptions(t *testing.T) {
	// A helper to simplify checks.
	check := func(t *testing.T, finding Finding, kind string, exceptions []string, recordPath string, knownServicePaths []string, expectedWaived []ExceptionIdentity, expectedErr bool) {
		t.Helper()
		res, err := ApplyExceptions(finding, kind, exceptions, recordPath, knownServicePaths)
		if (err != nil) != expectedErr {
			t.Errorf("ApplyExceptions() error = %v, expectedErr %v", err, expectedErr)
			return
		}
		if len(expectedWaived) == 0 {
			if len(res.WaivedBy) != 0 {
				t.Errorf("expected 0 waived, got %v", res.WaivedBy)
			}
		} else {
			if !reflect.DeepEqual(res.WaivedBy, expectedWaived) {
				t.Errorf("WaivedBy = %v, want %v", res.WaivedBy, expectedWaived)
			}
		}
	}

	t.Run("command kind never suppresses", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/main.go"}}
		check(t, f, "command", []string{"cmd/main.go"}, "rec.md", nil, nil, false)
	})

	t.Run("agent-review kind never suppresses", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/main.go"}}
		check(t, f, "agent-review", []string{"cmd/main.go"}, "rec.md", nil, nil, false)
	})

	t.Run("service shape exact match", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "services/api/main.go"}}
		// knownServicePaths includes services/api, so exc "services/api" matches it as a service.
		check(t, f, "grep", []string{"services/api"}, "rec.md", []string{"services/api"}, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("service shape directory match", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "services/api/handlers/foo.go"}}
		check(t, f, "grep", []string{"services/api"}, "rec.md", []string{"services/api"}, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("service shape outside directory", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "services/api-client/main.go"}}
		// Even though it starts with services/api, it's not under services/api/
		check(t, f, "grep", []string{"services/api"}, "rec.md", []string{"services/api"}, nil, false)
	})

	t.Run("path shape exact match", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/jigctl/main.go"}}
		// Contains / so it's path-shaped.
		check(t, f, "grep", []string{"cmd/jigctl/main.go"}, "rec.md", nil, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("path shape glob match", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/jigctl/main.go"}}
		check(t, f, "grep", []string{"cmd/**/*.go"}, "rec.md", nil, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("path shape no match", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/jigctl/main.go"}}
		check(t, f, "grep", []string{"pkg/**/*.go"}, "rec.md", nil, nil, false)
	})

	t.Run("invalid scope shape", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "cmd/jigctl/main.go"}}
		// Does not equal any knownServicePaths ("other"), does not contain glob chars or slash.
		// "repo" has no slash, no glob, not equal to "other". So it's invalid.
		check(t, f, "grep", []string{"repo"}, "rec.md", []string{"other"}, nil, true)
	})

	t.Run("service equality takes precedence over path shape", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "svc/foo/main.go"}}
		// "svc/foo" has a slash, so it could be path-shaped.
		// But it exactly equals servicePath, so it is evaluated as service-shaped.
		check(t, f, "grep", []string{"svc/foo"}, "rec.md", []string{"svc/foo"}, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("config-assert matches only file locus", func(t *testing.T) {
		f := Finding{Locus: Locus{File: "config.json", Pointer: "/foo"}}
		check(t, f, "config-assert", []string{"*.json"}, "rec.md", nil, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)
	})

	t.Run("repo-level binding service-shaped scope cross-service non-suppression", func(t *testing.T) {
		// This reproduces F3 scenario 6(b) where a repo-level record has an exception
		// for a single service (svc-a) but should not suppress findings in another service (svc-b).

		// 1. Check finding in svc-a (should be waived)
		fA := Finding{Locus: Locus{File: "svc-a/bad.go"}}
		check(t, fA, "grep", []string{"svc-a"}, "rec.md", []string{"svc-a", "svc-b"}, []ExceptionIdentity{
			{RecordPath: "rec.md", ExceptionIndex: 0},
		}, false)

		// 2. Check finding in svc-b (should NOT be waived)
		fB := Finding{Locus: Locus{File: "svc-b/bad.go"}}
		check(t, fB, "grep", []string{"svc-a"}, "rec.md", []string{"svc-a", "svc-b"}, nil, false)
	})
}
