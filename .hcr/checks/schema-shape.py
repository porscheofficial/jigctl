#!/usr/bin/env -S uv run --script
# /// script
# dependencies = ["pyyaml"]
# ///
import sys
import json

if len(sys.argv) != 1:
    sys.exit(1)

with open("schema/hcr.schema.json") as f:
    schema = json.load(f)

with open("corpus/RULES.md") as f:
    rules_md = f.read()

matrix_section = rules_md.split("## Allowed properties by binding kind")[1]
lines = [l.strip() for l in matrix_section.split("\n") if l.strip().startswith("|")]
lines = lines[2:] # skip header and separator

matrix = {}
for line in lines:
    parts = [p.strip() for p in line.split("|") if p.strip()]
    if len(parts) == 3:
        kind = parts[0]
        owns = parts[1].split() if parts[1] != "—" else []
        matrix[kind] = (set(owns), int(parts[2]))

names = {"run", "ref", "select", "timeout_secs", "pattern", "file", "path", "op", 
         "value", "require", "forbid", "tool", "docs", "prompt", 
         "grounding", "model", "runs"}
base = {"kind", "severity", "cadence"}

item_props = schema["properties"]["enforced_by"]["items"]["properties"]
actual_props = set(item_props.keys())
if actual_props != (names | base):
    print(f"Global properties mismatch. Missing: {(names | base) - actual_props}, Extra: {actual_props - (names | base)}")
    sys.exit(1)

branches = {}
for item in schema["properties"]["enforced_by"]["items"]["allOf"]:
    if "if" in item and "properties" in item["if"] and "kind" in item["if"]["properties"]:
        branches[item["if"]["properties"]["kind"]["const"]] = item

total_prohibits = 0
for kind, (owns, prohibits) in matrix.items():
    if kind not in branches:
        print(f"{kind}: missing from schema")
        sys.exit(1)
    
    then_props = branches[kind].get("then", {}).get("properties", {})
    false_props = {k for k, v in then_props.items() if v is False}
    expected_false = names - owns
    
    if false_props != expected_false:
        print(f"{kind}: prohibition mismatch. Missing: {expected_false - false_props}, Extra: {false_props - expected_false}")
        sys.exit(1)
        
    if len(false_props) != prohibits:
        print(f"{kind}: count mismatch. Expected {prohibits}, got {len(false_props)}")
        sys.exit(1)
        
    total_prohibits += prohibits

if total_prohibits != 84:
    print(f"Total prohibits mismatch. Expected 84, got {total_prohibits}")
    sys.exit(1)

print(f"Shape check passed for {len(matrix)} kinds: {', '.join(matrix.keys())}")
sys.exit(0)
