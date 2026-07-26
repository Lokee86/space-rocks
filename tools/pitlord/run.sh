#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

PITLORD="${PITLORD:-pitlord}"
LEXICON="${LEXICON:-lexicon}"
ARCANA="${ARCANA:-arcana}"
POLICY="${PITLORD_POLICY:-tools/pitlord/policy.json}"

"$PITLORD" validate --policy "$POLICY"

if [[ -f .lexicon/CURRENT ]]; then
  "$LEXICON" scan --repo .
else
  init_args=(init --repo .)
  if [[ -n "${LEXICON_ADAPTERS:-}" ]]; then
    init_args+=(--adapters "$LEXICON_ADAPTERS")
  fi
  if [[ -n "${LEXICON_LANGUAGES:-}" ]]; then
    init_args+=(--languages "$LEXICON_LANGUAGES")
  fi
  "$LEXICON" "${init_args[@]}"
fi

"$ARCANA" sync --lexicon .lexicon --state .arcana
"$PITLORD" check \
  --repo . \
  --policy "$POLICY" \
  --arcana "$ARCANA" \
  --timeout "${PITLORD_TIMEOUT:-5m}"
