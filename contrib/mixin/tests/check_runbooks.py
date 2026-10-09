"""Check every mixin alert has a complete runbook at its default URL."""

import json
from pathlib import Path
import sys
from urllib.parse import urlparse


runbooks = Path(sys.argv[1])
alerts = json.load(sys.stdin)
errors = []
names = set()
for group in alerts["groups"]:
    for rule in group["rules"]:
        if "alert" not in rule:
            continue
        name = rule["alert"]
        names.add(name)
        slug = name.lower()
        expected_path = f"/windows_exporter/runbooks/{slug}/"
        if urlparse(rule["annotations"].get("runbook_url", "")).path != expected_path:
            errors.append(f"{name}: runbook URL must end in {expected_path}")
        path = runbooks / f"{slug}.md"
        if not path.is_file():
            errors.append(f"{name}: missing {path}")
            continue
        content = path.read_text(encoding="utf-8")
        for section in ("Meaning", "Impact", "Diagnosis", "Mitigation"):
            if f"## {section}\n" not in content:
                errors.append(f"{name}: missing {section} section in {path}")
        if "TODO" in content:
            errors.append(f"{name}: unfinished runbook {path}")

if errors:
    sys.exit("\n".join(errors))
print(f"Checked runbooks and URLs for {len(names)} alerts")
