# /// script
# dependencies = [
#   "pyyaml",
# ]
# ///

import sys, glob, re, yaml, json, subprocess, tempfile, os

if len(sys.argv) != 1:
    sys.exit("Usage: corpus-runner.py takes no arguments")

globs = [
    "corpus/records/*.md",
    "corpus/fixtures/*/.hcr/*.md",
    "corpus/fixtures/*/services/*/.hcr/*.md"
]
files = [f for g in globs for f in glob.glob(g)]
if not files:
    sys.exit("No fixtures found")

try:
    with open("corpus/RULES.md") as f:
        rules_md = f.read()
    rules_in_md = set(re.findall(r'\|\s*(R-\d+)\s*\|', rules_md))
except Exception:
    sys.exit("Failed to read RULES.md")

results = {"validated": 0, "reported": 0, "missing": 0, "noisy": 0, "error": 0, "deferred": 0}
cited = set()
temp_dir = tempfile.mkdtemp()
fmap = {}

for fp in files:
    with open(fp) as f:
        c = f.read()
    
    m = re.match(r'^---\n(.*?)\n---\n', c, re.DOTALL)
    if not m:
        results["error"] += 1; continue
        
    fm = m.group(1)
    ex_blocks = re.findall(r'<!--\s*jig:expect\n(.*?)\n-->', c, re.DOTALL)
    if len(ex_blocks) != 1:
        results["error"] += 1; continue
        
    try:
        ex = yaml.safe_load(ex_blocks[0])
    except Exception:
        results["error"] += 1; continue
        
    if not isinstance(ex, dict) or "valid" not in ex or not isinstance(ex["valid"], bool):
        results["error"] += 1; continue
        
    if not set(ex.keys()).issubset({"valid", "covers", "diagnostics", "deferred"}):
        results["error"] += 1; continue
        
    cv = ex.get("covers", [])
    dg = ex.get("diagnostics", [])
    df = ex.get("deferred", [])
    
    if ex["valid"] and "diagnostics" in ex:
        results["error"] += 1; continue
    if not ex["valid"] and "diagnostics" not in ex:
        results["error"] += 1; continue
        
    if not ex["valid"]:
        if len(cv) != 1 or len(dg) != 1 or dg[0].get("rule") != cv[0]:
            results["error"] += 1; continue
            
    for r in cv: cited.add(r)
    for d in df: cited.add(d.get("rule"))
    results["deferred"] += len(df)
    
    tp = os.path.join(temp_dir, str(len(fmap)) + ".yaml")
    with open(tp, "w") as tf: tf.write(fm)
    fmap[tp] = (fp, ex, dg)

cmd = ["check-jsonschema", "--schemafile", "schema/hcr.schema.json", "-o", "json"] + list(fmap.keys())
res = subprocess.run(cmd, capture_output=True, text=True)
out = json.loads(res.stdout) if res.stdout else {}

errs_by_file = {k: [] for k in fmap}
for e in out.get("errors", []):
    fn = e.get("filename")
    if fn in errs_by_file:
        p = e.get("path", "")
        if p == "$": p = ""
        else:
            p = p.replace("$", "")
            p = re.sub(r'\.([^.\[]+)', r'/\1', p)
            p = re.sub(r'\[(\d+)\]', r'/\1', p)
        errs_by_file[fn].append(p)

for tp, (fp, ex, dg) in fmap.items():
    errs = errs_by_file[tp]
    if ex["valid"]:
        if len(errs) == 0: results["validated"] += 1
        else: results["noisy"] += 1
    else:
        if len(errs) == 0: results["missing"] += 1
        elif len(errs) == 1 and errs[0] == dg[0].get("at"): results["reported"] += 1
        else: results["noisy"] += 1

print(" ".join(f"{k}={v}" for k, v in results.items() if k != "deferred") + f" deferred={results['deferred']}")
if not cited.issubset(rules_in_md) or results["error"] > 0 or results["missing"] > 0 or results["noisy"] > 0:
    sys.exit(1)
if sum(results.values()) - results["deferred"] != len(files):
    sys.exit(1)
