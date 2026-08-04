#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"
PYTHON="${PYTHON:-python}"
DATA_SYNC=("$PYTHON" tools/data_sync/main.py)
export PYTHONDONTWRITEBYTECODE=1

"$PYTHON" -m pytest tests tools/tests tools/data_sync/tests
"$PYTHON" .standards/docs_policy/check.py --repo .
pitlord check --repo . --policy tools/pitlord/policy.json
"${DATA_SYNC[@]}" -validate
"${DATA_SYNC[@]}" -check -constants -go -gds
"${DATA_SYNC[@]}" -check -packets -go -gds
"${DATA_SYNC[@]}" -check -realtime-wire -go -gds -json -docs
"${DATA_SYNC[@]}" -check -drop-tables -go
