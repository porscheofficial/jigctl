#!/usr/bin/env -S uv run --script
# /// script
# dependencies = []
# ///
import os
import sys
import json
import subprocess
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent.parent
fixture_tree = repo_root / "corpus" / "fixtures" / "multi-service"
schema_file = repo_root / "schema" / "run-output-v1.schema.json"

if not fixture_tree.is_dir():
    print(f"Fixture tree not found: {fixture_tree}")
    sys.exit(1)

env = os.environ.copy()
env.pop("JIGCTL_ALLOW_EXEC", None)

try:
    result = subprocess.run(
        ["go", "run", "./cmd/jigctl", "run", str(fixture_tree), "--format=json"],
        cwd=repo_root,
        env=env,
        capture_output=True,
        text=True,
    )
    
    if result.returncode not in (0, 1, 77):
        print(f"jigctl run failed with unexpected exit code {result.returncode}:\n{result.stderr}\n{result.stdout}")
        sys.exit(1)

    output = result.stdout
    
    try:
        json_data = json.loads(output)
    except json.JSONDecodeError as e:
        print(f"Failed to parse JSON output:\n{e}\nOutput was:\n{output}")
        sys.exit(1)

    schema_check = subprocess.run(
        ["mise", "exec", "--", "check-jsonschema", "--schemafile", str(schema_file), "-"],
        input=output,
        text=True,
        capture_output=True,
    )

    if schema_check.returncode != 0:
        print(f"Schema validation failed:\n{schema_check.stdout}\n{schema_check.stderr}")
        sys.exit(1)

    # --only-failures: the filtered document must independently satisfy the same
    # v1 schema, and must agree with the unfiltered document on everything the
    # filter is not supposed to touch.
    result_filtered = subprocess.run(
        ["go", "run", "./cmd/jigctl", "run", str(fixture_tree), "--format=json", "--only-failures"],
        cwd=repo_root,
        env=env,
        capture_output=True,
        text=True,
    )

    if result_filtered.returncode not in (0, 1, 77):
        print(f"jigctl run --only-failures failed with unexpected exit code {result_filtered.returncode}:\n{result_filtered.stderr}\n{result_filtered.stdout}")
        sys.exit(1)

    filtered_output = result_filtered.stdout

    try:
        filtered_json_data = json.loads(filtered_output)
    except json.JSONDecodeError as e:
        print(f"Failed to parse JSON output (--only-failures):\n{e}\nOutput was:\n{filtered_output}")
        sys.exit(1)

    filtered_schema_check = subprocess.run(
        ["mise", "exec", "--", "check-jsonschema", "--schemafile", str(schema_file), "-"],
        input=filtered_output,
        text=True,
        capture_output=True,
    )

    if filtered_schema_check.returncode != 0:
        print(f"Schema validation failed (--only-failures):\n{filtered_schema_check.stdout}\n{filtered_schema_check.stderr}")
        sys.exit(1)

    if filtered_json_data["summary"] != json_data["summary"]:
        print(
            "Invariant violated: --only-failures changed `summary`, but summary must be "
            "computed over the entire run regardless of filtering.\n"
            f"unfiltered summary: {json.dumps(json_data['summary'])}\n"
            f"filtered summary:   {json.dumps(filtered_json_data['summary'])}"
        )
        sys.exit(1)

    if filtered_json_data["exit_code"] != json_data["exit_code"]:
        print(
            "Invariant violated: --only-failures changed `exit_code`, but exit_code must be "
            "computed over the entire run regardless of filtering.\n"
            f"unfiltered exit_code: {json_data['exit_code']}\n"
            f"filtered exit_code:   {filtered_json_data['exit_code']}"
        )
        sys.exit(1)

    actionable_projections = {"violation", "operational", "invalid", "blocked-unchecked"}
    leaked = [
        rec.get("record_id") for rec in filtered_json_data["records"]
        if rec.get("projection") not in actionable_projections
    ]
    if leaked:
        print(
            "Invariant violated: --only-failures kept non-actionable record(s) "
            f"{leaked} whose projection is not in {sorted(actionable_projections)} "
            "(only pass/expected-unchecked records should ever be dropped, and none of "
            "the survivors may carry them)."
        )
        sys.exit(1)

    print("JSON output matches schema successfully.")
    print("--only-failures output matches schema and preserves summary/exit_code while dropping non-actionable records.")
    sys.exit(0)

except Exception as e:
    print(f"Unexpected error: {e}")
    sys.exit(1)