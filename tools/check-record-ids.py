#!/usr/bin/env -S uv run --script
# /// script
# dependencies = []
# ///
import re
import sys
from pathlib import Path

if len(sys.argv) != 1:
    sys.exit(1)

records = sorted(Path(".hcr").glob("*.md"))
for record in records:
    match = re.search(r"^id:\s*(\S+)", record.read_text(), re.MULTILINE)
    record_id = match.group(1) if match else "<missing>"
    if not re.fullmatch(r"HCR-04\d{2}", record_id):
        print(f"{record}: id {record_id} is outside the HCR-04xx band")
        sys.exit(1)

print(f"Record id check passed for {len(records)} records in the HCR-04xx band")
sys.exit(0)
