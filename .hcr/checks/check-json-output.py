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

    print("JSON output matches schema successfully.")
    sys.exit(0)

except Exception as e:
    print(f"Unexpected error: {e}")
    sys.exit(1)