from __future__ import annotations

import argparse
import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.multi_room_matrix_manifest import MatrixManifest
from runtime_scenarios.multi_room_matrix_summary import summarize_matrix
from runtime_scenarios.runner import RunOptions, ScenarioRunner


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Run the one/two/three/four-room runtime matrix sequentially."
    )
    parser.add_argument("manifest", type=Path)
    parser.add_argument("--godot", help="Godot editor executable path")
    parser.add_argument("--output-root", type=Path)
    parser.add_argument("--validate-only", action="store_true")
    parser.add_argument(
        "--controlled-host",
        action="store_true",
        help="declare that unrelated host activity was controlled for performance evidence",
    )
    parser.add_argument(
        "--host-note",
        default="",
        help="describe host isolation, hardware, or known contention for every matrix point",
    )
    return parser


def main() -> int:
    args = build_parser().parse_args()
    try:
        manifest = MatrixManifest.load(args.manifest)
    except ValueError as exc:
        print(f"multi-room matrix invalid: {exc}", file=sys.stderr)
        return 2
    if args.validate_only:
        print(
            f"valid matrix: {manifest.matrix_id} "
            f"({', '.join(str(entry.room_count) for entry in manifest.entries)} rooms)"
        )
        return 0

    repo_root = Path(__file__).resolve().parents[2]
    base_output = (
        args.output_root.expanduser().resolve()
        if args.output_root
        else repo_root / ".ci-artifacts" / "runtime-scenarios"
    )
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    matrix_directory = base_output / f"{manifest.matrix_id}-{stamp}"
    run_directories: list[Path] = []
    status: dict[str, Any] = {
        "matrix_id": manifest.matrix_id,
        "manifest_path": str(manifest.path),
        "controlled_host": args.controlled_host,
        "host_note": args.host_note.strip(),
        "runs": [],
        "success": False,
    }

    try:
        for entry in manifest.entries:
            runner = ScenarioRunner(
                entry.scenario,
                RunOptions(
                    repo_root=repo_root,
                    output_root=matrix_directory,
                    godot=args.godot,
                    headless_coordinator=True,
                    controlled_host=args.controlled_host,
                    host_note=args.host_note,
                ),
            )
            exit_code = runner.run()
            run_directories.append(runner.run_directory)
            status["runs"].append(
                {
                    "room_count": entry.room_count,
                    "scenario_id": entry.scenario.scenario_id,
                    "run_directory": str(runner.run_directory),
                    "exit_code": exit_code,
                }
            )
            if exit_code != 0:
                raise RuntimeError(f"{entry.scenario.scenario_id} failed")

        result = summarize_matrix(manifest, run_directories)
        status["success"] = True
        status["summary"] = result
        matrix_directory.mkdir(parents=True, exist_ok=True)
        (matrix_directory / "matrix-summary.json").write_text(
            json.dumps(result, indent=2) + "\n",
            encoding="utf-8",
        )
        print(json.dumps(result, indent=2))
        return 0
    except Exception as exc:  # noqa: BLE001 - CLI records partial matrix state
        status["error"] = str(exc)
        print(f"multi-room matrix failed: {exc}", file=sys.stderr)
        return 1
    finally:
        matrix_directory.mkdir(parents=True, exist_ok=True)
        status["ended_at"] = datetime.now(timezone.utc).isoformat()
        (matrix_directory / "matrix-run.json").write_text(
            json.dumps(status, indent=2) + "\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    raise SystemExit(main())
