from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.multi_room_matrix_manifest import MatrixManifest
from runtime_scenarios.multi_room_summary import summarize_multi_room


def summarize_matrix(
    manifest: MatrixManifest,
    run_directories: list[Path],
) -> dict[str, Any]:
    if len(run_directories) != len(manifest.entries):
        raise ValueError(
            f"matrix requires {len(manifest.entries)} run directories, got {len(run_directories)}"
        )

    results_by_count: dict[int, dict[str, Any]] = {}
    host_entries: list[dict[str, Any]] = []
    for run_directory in run_directories:
        resolved = run_directory.expanduser().resolve()
        run = _read_json(resolved / "summary.json")
        scenario_id = str(run.get("scenario_id", ""))
        scenario_path = Path(str(run.get("scenario_path", ""))).expanduser().resolve()
        expected = next(
            (entry for entry in manifest.entries if entry.scenario.scenario_id == scenario_id),
            None,
        )
        if expected is None:
            raise ValueError(f"run {resolved} does not belong to matrix {manifest.matrix_id}")
        if scenario_path != expected.scenario.path:
            raise ValueError(f"run {resolved} scenario path does not match the matrix manifest")
        if expected.room_count in results_by_count:
            raise ValueError(f"duplicate matrix result for {expected.room_count} rooms")

        result = summarize_multi_room(resolved)
        if int(result.get("room_count", 0)) != expected.room_count:
            raise ValueError(f"run {resolved} room count does not match the matrix manifest")
        host_control = run.get("host_control", {})
        host_control = host_control if isinstance(host_control, dict) else {}
        controlled = bool(host_control.get("controlled", False))
        host_entry = {
            "room_count": expected.room_count,
            "controlled": controlled,
            "note": str(host_control.get("note", "")).strip(),
        }
        host_entries.append(host_entry)
        results_by_count[expected.room_count] = {
            "room_count": expected.room_count,
            "scenario_id": scenario_id,
            "run_directory": str(resolved),
            "success": bool(result.get("success", False)),
            "unique_room_codes": int(result.get("unique_room_codes", 0)),
            "unique_match_ids": int(result.get("unique_match_ids", 0)),
            "aggregate": result.get("aggregate", {}),
        }

    ordered = [results_by_count[entry.room_count] for entry in manifest.entries]
    functional_pass = all(
        bool(result["success"])
        and int(result["unique_room_codes"]) == int(result["room_count"])
        and int(result["unique_match_ids"]) == int(result["room_count"])
        for result in ordered
    )
    all_controlled = all(entry["controlled"] for entry in host_entries)
    return {
        "matrix_id": manifest.matrix_id,
        "manifest_path": str(manifest.path),
        "workload_signature": manifest.workload_signature,
        "functional_pass": functional_pass,
        "performance_eligible": functional_pass and all_controlled,
        "host_control": {
            "all_controlled": all_controlled,
            "runs": sorted(host_entries, key=lambda entry: int(entry["room_count"])),
        },
        "results": ordered,
    }


def _read_json(path: Path) -> dict[str, Any]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(payload, dict):
        raise ValueError(f"expected JSON object: {path}")
    return payload


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Aggregate the controlled one/two/three/four-room runtime matrix."
    )
    parser.add_argument("manifest", type=Path)
    parser.add_argument("run_directories", nargs="+", type=Path)
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--require-controlled-host",
        action="store_true",
        help="fail unless every run was declared controlled",
    )
    args = parser.parse_args()
    try:
        manifest = MatrixManifest.load(args.manifest)
        result = summarize_matrix(manifest, args.run_directories)
    except ValueError as exc:
        print(f"multi-room matrix summary failed: {exc}", file=sys.stderr)
        return 2
    if args.require_controlled_host and not result["performance_eligible"]:
        raise SystemExit("matrix is not eligible for performance comparison")
    encoded = json.dumps(result, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(encoded, encoding="utf-8")
    print(encoded, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
