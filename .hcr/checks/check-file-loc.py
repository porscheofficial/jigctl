#!/usr/bin/env -S uv run --script
# /// script
# dependencies = []
# ///
import sys
from pathlib import Path

LIMIT = 250

if len(sys.argv) != 1:
    sys.exit(1)


def pure_lines(text):
    lines = (line.strip() for line in text.splitlines())
    return sum(1 for line in lines if line and not line.startswith("//"))


files = sorted(Path(".").rglob("*.go"))
oversized = False
for source in files:
    count = pure_lines(source.read_text())
    if count > LIMIT:
        print(f"{source}: {count} pure lines exceeds the {LIMIT}-line ceiling")
        oversized = True

if oversized:
    sys.exit(1)

print(f"File length check passed for {len(files)} Go files within {LIMIT} pure lines")
sys.exit(0)
