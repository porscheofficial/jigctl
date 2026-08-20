package hcr

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var planTestDate = time.Date(2026, time.August, 19, 15, 0, 0, 0, time.UTC)

func TestExecutionPlan_repo_only_tree_has_one_target_with_every_binding(t *testing.T) {
	// Given
	root := t.TempDir()
	writePlanFixture(t, root, "service_globs = []\n", map[string]string{
		".hcr/HCR-9001-command.md": commandRecord(commandFixture{
			id: "HCR-9001", scope: "repo", ref: "repo-check", run: "go test ./...",
		}),
		".hcr/HCR-9002-grep.md": `---
id: HCR-9002
title: Check release notes
scope: repo
regulates: maintainability
summary: The release notes retain their required heading.
state: enforced
enforced_by:
  - kind: grep
    severity: blocking
    cadence: [ci]
    file: CHANGELOG.md
    require: [Unreleased]
---
`,
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(plan.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(plan.Targets))
	}
	assertRepoTarget(t, plan.Targets[0])
	if plan.Root != filepath.Clean(root) {
		t.Fatalf("expected absolute root %q, got %q", filepath.Clean(root), plan.Root)
	}
	t.Logf("repo-only targets=%d bindings=%d", len(plan.Targets), bindingCount(plan.Targets))
}

func TestExecutionPlan_two_services_assigns_each_binding_once(t *testing.T) {
	// Given
	root := t.TempDir()
	writePlanFixture(t, root, "service_globs = [\"services/*\"]\n", map[string]string{
		".hcr/HCR-9001-repo.md": commandRecord(commandFixture{
			id: "HCR-9001", scope: "repo", ref: "repo-check", run: "go test ./...",
		}),
		"scripts/check-api-contracts.sh": "#!/bin/sh\nexit 0\n",
		"services/api/.hcr/HCR-9002-api.md": commandRecord(commandFixture{
			id: "HCR-9002", scope: "service", ref: "api-check", run: "scripts/check-api-contracts.sh",
		}),
		"services/billing/.hcr/HCR-9003-billing.md": commandRecord(commandFixture{
			id: "HCR-9003", scope: "service", ref: "billing-check", run: "go test ./services/billing/...",
		}),
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(plan.Targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(plan.Targets))
	}
	kindCounts := make(map[string]int)
	seen := make(map[string]int)
	for _, target := range plan.Targets {
		kindCounts[target.Kind]++
		for _, binding := range target.Bindings {
			seen[binding.RecordPath]++
			if target.Kind == "service" && binding.RecordID == "HCR-9001" {
				t.Fatalf("repo binding duplicated into service target %q", target.Path)
			}
		}
	}
	if len(seen) != 3 || bindingCount(plan.Targets) != 3 {
		t.Fatalf("expected 3 unique bindings, got unique=%d total=%d", len(seen), bindingCount(plan.Targets))
	}
	if kindCounts["repo"] != 1 || kindCounts["service"] != 2 {
		t.Fatalf("expected one repo and two service targets, got %#v", kindCounts)
	}
	if plan.Targets[1].Path != filepath.Join("services", "api") ||
		plan.Targets[2].Path != filepath.Join("services", "billing") {
		t.Fatalf("unexpected service identities: %#v", plan.Targets)
	}
	for path, count := range seen {
		if count != 1 {
			t.Fatalf("binding %q appeared %d times", path, count)
		}
	}
	t.Logf("two-service targets=%d bindings=%d", len(plan.Targets), bindingCount(plan.Targets))
}

func TestExecutionPlan_schema_invalid_tree_returns_zero_plan(t *testing.T) {
	// Given
	root := t.TempDir()
	writePlanFixture(t, root, "service_globs = []\n", map[string]string{
		".hcr/HCR-9001-invalid.md": commandRecord(commandFixture{
			id: "not-an-hcr-id", scope: "repo", ref: "bad", run: "go test ./...",
		}),
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) == 0 {
		t.Fatal("expected schema diagnostics")
	}
	if plan.Root != "" || len(plan.Targets) != 0 {
		t.Fatalf("expected zero plan, got %#v", plan)
	}
}

func TestExecutionPlan_unknown_frontmatter_key_cannot_yield_command(t *testing.T) {
	// Given
	root := t.TempDir()
	record := commandRecord(commandFixture{
		id: "HCR-9001", scope: "repo", ref: "blocked", run: "printf forbidden",
	})
	record = record[:len(record)-5] + "unknown_execution_switch: true\n---\n"
	writePlanFixture(t, root, "service_globs = []\n", map[string]string{
		".hcr/HCR-9001-unknown-key.md": record,
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) == 0 {
		t.Fatal("expected unknown-key diagnostic")
	}
	if len(plan.Targets) != 0 {
		t.Fatalf("malformed tree yielded %d executable targets", len(plan.Targets))
	}
	t.Logf("malformed-tree targets=%d diagnostics=%d", len(plan.Targets), len(diagnostics))
}

func TestExecutionPlan_binding_title_matches_frontmatter_verbatim(t *testing.T) {
	// Given
	root := t.TempDir()
	writePlanFixture(t, root, "service_globs = []\n", map[string]string{
		".hcr/HCR-9001-colon.md": `---
id: HCR-9001
title: "Fix: use go vet, not lint"
scope: repo
regulates: reliability
summary: The command is validated before it can execute.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: colon-check
    run: go vet ./...
---
`,
		".hcr/HCR-9002-nonascii.md": `---
id: HCR-9002
title: Vérifie que le café est bien testé
scope: repo
regulates: reliability
summary: The command is validated before it can execute.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: nonascii-check
    run: go vet ./...
---
`,
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	titles := make(map[string]string, 2)
	for _, binding := range plan.Targets[0].Bindings {
		titles[binding.RecordID] = binding.Title
	}
	if titles["HCR-9001"] != `Fix: use go vet, not lint` {
		t.Fatalf("colon title mismatch: %#v", titles["HCR-9001"])
	}
	if titles["HCR-9002"] != "Vérifie que le café est bien testé" {
		t.Fatalf("non-ASCII title mismatch: %#v", titles["HCR-9002"])
	}
}

func writePlanFixture(t *testing.T, root, config string, records map[string]string) {
	t.Helper()
	mustNoError(t, os.WriteFile(filepath.Join(root, "jig.toml"), []byte(config), 0o600))
	for relative, source := range records {
		path := filepath.Join(root, relative)
		mustNoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		mustNoError(t, os.WriteFile(path, []byte(source), 0o600))
	}
}

type commandFixture struct {
	id    string
	scope string
	ref   string
	run   string
}

func commandRecord(fixture commandFixture) string {
	return "---\n" +
		"id: " + fixture.id + "\n" +
		"title: Execute a validated command\n" +
		"scope: " + fixture.scope + "\n" +
		"regulates: reliability\n" +
		"summary: The command is validated before it can execute.\n" +
		"state: enforced\n" +
		"enforced_by:\n" +
		"  - kind: command\n" +
		"    severity: blocking\n" +
		"    cadence: [ci]\n" +
		"    ref: " + fixture.ref + "\n" +
		"    run: " + fixture.run + "\n" +
		"---\n"
}

func bindingCount(targets []Target) int {
	total := 0
	for _, target := range targets {
		total += len(target.Bindings)
	}
	return total
}

func assertRepoTarget(t *testing.T, target Target) {
	t.Helper()
	if target.Kind != "repo" || target.Path != "" || len(target.Bindings) != 2 {
		t.Fatalf("unexpected repo target: %#v", target)
	}
	command := target.Bindings[0]
	if command.RecordID != "HCR-9001" || command.BindingIndex != 0 || command.Kind != "command" ||
		command.Severity != "blocking" || len(command.Cadence) != 1 || command.Cadence[0] != "ci" ||
		command.Ref != "repo-check" || command.Run != "go test ./..." {
		t.Fatalf("command execution fields were not preserved: %#v", command)
	}
	grep := target.Bindings[1]
	if grep.File != "CHANGELOG.md" || len(grep.Require) != 1 || grep.Require[0] != "Unreleased" {
		t.Fatalf("grep execution fields were not preserved: %#v", grep)
	}
}
