#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

failed_stages=()

run_stage() {
  local name="$1"
  shift
  printf '\n==> %s\n' "$name"
  set +e
  "$@"
  local status=$?
  set -e
  if ((status != 0)); then
    failed_stages+=("$name")
    printf '<== %s failed (status %d)\n' "$name" "$status"
  else
    printf '<== %s passed\n' "$name"
  fi
}

pushd shared/go/servicelog >/dev/null
run_stage "shared/go/servicelog" go test -buildvcs=false ./...
popd >/dev/null

pushd services/player-data >/dev/null
run_stage "player-data" go test -buildvcs=false ./...
popd >/dev/null

pushd services/game-server >/dev/null
run_stage "game-server default" go test -buildvcs=false ./...
run_stage "game-server nodevtools" go test -tags nodevtools -buildvcs=false ./...
popd >/dev/null

if ((${#failed_stages[@]} > 0)); then
  printf '\nFailed stages: %s\n' "$(IFS=', '; printf '%s' "${failed_stages[*]}")"
  exit 1
fi

printf '\nAll Go test stages passed.\n'
