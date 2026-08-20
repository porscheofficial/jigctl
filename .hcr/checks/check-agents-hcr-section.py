#!/usr/bin/env -S uv run --script
# /// script
# dependencies = []
# ///
import re
import sys
from pathlib import Path

HEADING = "## Constraint Records"
TEMPLATE = Path(".agents/skills/hcr-author/references/agents-section.md")
TARGET = Path("AGENTS.md")

if len(sys.argv) != 1:
    sys.exit(1)


def section(path):
    """The lines under HEADING, up to the next second-level heading."""
    lines = path.read_text().splitlines()
    for index, line in enumerate(lines):
        if line.strip() != HEADING:
            continue
        rest = lines[index + 1 :]
        for offset, following in enumerate(rest):
            if following.startswith("## "):
                return rest[:offset]
        return rest
    print(f"{path}: no '{HEADING}' section")
    sys.exit(1)


def bullets(lines):
    """Whole bullets, wrapped continuations rejoined, whitespace collapsed."""
    items = []
    for line in lines:
        if line.startswith("- "):
            items.append(line[2:])
        elif items and line.startswith("  ") and line.strip():
            items[-1] += " " + line
    return [re.sub(r"\s+", " ", item).strip() for item in items]


required = bullets(section(TEMPLATE))
if not required:
    print(f"{TEMPLATE}: '{HEADING}' carries no bullets to check against")
    sys.exit(1)

present = re.sub(r"\s+", " ", " ".join(section(TARGET))).strip()
missing = [bullet for bullet in required if bullet not in present]

if missing:
    print(f"{TARGET}: '{HEADING}' has drifted from {TEMPLATE}")
    for bullet in missing:
        print(f"  missing: {bullet[:68]}...")
    sys.exit(1)

print(f"AGENTS.md carries all {len(required)} bullets of the hcr-author section")
sys.exit(0)
