#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLIENT="$ROOT/client"
ARTIFACT_DIR="${GODOT_ARTIFACT_DIR:-$ROOT/.ci-artifacts/godot}"
BOOTSTRAP_TIMEOUT="${GODOT_BOOTSTRAP_TIMEOUT:-120}"
IMPORT_TIMEOUT="${GODOT_IMPORT_TIMEOUT:-120}"
GUT_TIMEOUT="${GODOT_GUT_TIMEOUT:-300}"
BOOTSTRAP_FRAMES="${GODOT_BOOTSTRAP_FRAMES:-1200}"
GODOT="${GODOT_BIN:-godot}"
mkdir -p "$ARTIFACT_DIR"

for required in "$GODOT" timeout tee grep sleep; do
  command -v "$required" >/dev/null 2>&1 || { printf 'Required command unavailable: %s\n' "$required" >&2; exit 127; }
done
GODOT_CLIENT="$CLIENT"
if command -v cygpath >/dev/null 2>&1; then
  GODOT_CLIENT="$(cygpath -w "$CLIENT")"
fi

run_logged() {
  local name="$1" timeout_seconds="$2" engine_log="$3"
  shift 3
  local engine_log_arg="$engine_log"
  if command -v cygpath >/dev/null 2>&1; then
    engine_log_arg="$(cygpath -w "$engine_log")"
  fi
  set +e
  timeout --signal=TERM --kill-after=10s "${timeout_seconds}s" "$@" --log-file "$engine_log_arg" \
    2>&1 | tee "$ARTIFACT_DIR/${name}.stdout.log"
  local status=${PIPESTATUS[0]}
  set -e
  if [[ $status -ne 0 ]]; then
    printf 'Godot step %s failed with status %s; see %s.stdout.log and %s\n' \
      "$name" "$status" "$name" "$engine_log" >&2
    return "$status"
  fi
}

run_gut_logged() {
  local name="$1" timeout_seconds="$2" engine_log="$3"
  shift 3
  local engine_log_arg="$engine_log"
  if command -v cygpath >/dev/null 2>&1; then
    engine_log_arg="$(cygpath -w "$engine_log")"
  fi
  local stdout_log="$ARTIFACT_DIR/${name}.stdout.log"
  local fatal_status=86
  local fatal_regex='GDScript[^[:cntrl:]]*(parse|compile|load)|parse error|compile error|failed to load (script|test)|cannot load script|could not load script|error loading script|ignoring invalid test script|invalid test script|Ignoring script .* because it does not extend GutTest|ignoring script .* because it does not extend GutTest'

  : > "$stdout_log"
  set +e
  timeout --signal=TERM --kill-after=10s "${timeout_seconds}s" "$@" --log-file "$engine_log_arg" \
    > >(tee "$stdout_log") 2>&1 &
  local runner_pid=$!
  local fatal_line=""
  while kill -0 "$runner_pid" 2>/dev/null; do
    if [[ -f "$stdout_log" ]] && fatal_line="$(grep -Eim1 "$fatal_regex" "$stdout_log")"; then
      printf 'Fatal GUT collection/load error detected: %s; see %s and %s\n' \
        "$fatal_line" "$stdout_log" "$engine_log" >&2
      kill -TERM "$runner_pid" 2>/dev/null || true
      break
    fi
    sleep 0.2
  done
  wait "$runner_pid"
  local status=$?
  set -e
  if [[ -z "$fatal_line" ]] && fatal_line="$(grep -Eim1 "$fatal_regex" "$stdout_log")"; then
    printf 'Fatal GUT collection/load error detected: %s; see %s and %s\n' \
      "$fatal_line" "$stdout_log" "$engine_log" >&2
  fi
  if [[ -n "$fatal_line" ]]; then
    return "$fatal_status"
  fi
  if [[ $status -ne 0 ]]; then
    printf 'Godot step %s failed with status %s; see %s.stdout.log and %s\n' \
      "$name" "$status" "$name" "$engine_log" >&2
    return "$status"
  fi
}

cd "$ROOT"
if command -v xvfb-run >/dev/null 2>&1; then
  DISPLAY_COMMAND=(xvfb-run -a)
elif [[ "${GODOT_REQUIRE_XVFB:-0}" == "1" ]]; then
  printf 'Required command unavailable: xvfb-run\n' >&2
  exit 127
else
  DISPLAY_COMMAND=()
fi

if ((${#DISPLAY_COMMAND[@]})); then
  run_logged bootstrap "$BOOTSTRAP_TIMEOUT" "$ARTIFACT_DIR/bootstrap.engine.log" \
    "${DISPLAY_COMMAND[@]}" "$GODOT" --path "$GODOT_CLIENT" --editor \
    --rendering-method gl_compatibility --rendering-driver opengl3 --quit-after "$BOOTSTRAP_FRAMES"
else
  run_logged bootstrap "$BOOTSTRAP_TIMEOUT" "$ARTIFACT_DIR/bootstrap.engine.log" \
    "$GODOT" --path "$GODOT_CLIENT" --editor --headless \
    --rendering-method gl_compatibility --rendering-driver opengl3 --quit-after "$BOOTSTRAP_FRAMES"
fi
run_logged import "$IMPORT_TIMEOUT" "$ARTIFACT_DIR/import.engine.log" \
  "$GODOT" --headless --path "$GODOT_CLIENT" --import --quit
run_gut_logged gut "$GUT_TIMEOUT" "$ARTIFACT_DIR/gut.engine.log" \
  "$GODOT" --headless --path "$GODOT_CLIENT" -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit

find "$CLIENT" -type f \( -name '*.log' -o -name 'godot.log' \) -exec cp --parents '{}' "$ARTIFACT_DIR" \; 2>/dev/null || true
