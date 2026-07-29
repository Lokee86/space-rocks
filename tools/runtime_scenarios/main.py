from __future__ import annotations

import argparse
import sys
from pathlib import Path

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.model import Scenario, ScenarioError
from runtime_scenarios.runner import RunOptions, ScenarioRunner


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Run a deterministic Space Rocks client/server runtime scenario."
    )
    parser.add_argument("scenario", type=Path, help="scenario JSON file")
    parser.add_argument("--godot", help="Godot editor executable path")
    parser.add_argument(
        "--output-root",
        type=Path,
        help="run artifact directory (default: .ci-artifacts/runtime-scenarios)",
    )
    parser.add_argument(
        "--validate-only",
        action="store_true",
        help="validate the scenario without starting processes",
    )
    parser.add_argument(
        "--headless-coordinator",
        action="store_true",
        help="run the coordinator headlessly for unattended orchestration verification",
    )
    parser.add_argument(
        "--controlled-host",
        action="store_true",
        help="declare that unrelated host activity was controlled for performance evidence",
    )
    parser.add_argument(
        "--host-note",
        default="",
        help="describe host isolation, hardware, or known contention for this run",
    )
    parser.add_argument(
        "--server-url",
        help=(
            "use an already deployed game server instead of launching one; "
            "accepts ws://, wss://, http://, or https://"
        ),
    )
    return parser


def main() -> int:
    args = build_parser().parse_args()
    repo_root = Path(__file__).resolve().parents[2]
    try:
        scenario = Scenario.load(args.scenario)
    except ScenarioError as exc:
        print(f"runtime scenario invalid: {exc}")
        return 2
    if args.validate_only:
        print(
            f"valid scenario: {scenario.scenario_id} "
            f"({scenario.clients.total} clients, {scenario.bots} bots, seed {scenario.seed})"
        )
        return 0

    output_root = (
        args.output_root.expanduser().resolve()
        if args.output_root
        else repo_root / ".ci-artifacts" / "runtime-scenarios"
    )
    runner = ScenarioRunner(
        scenario,
        RunOptions(
            repo_root=repo_root,
            output_root=output_root,
            godot=args.godot,
            headless_coordinator=args.headless_coordinator,
            controlled_host=args.controlled_host,
            host_note=args.host_note,
            server_url=args.server_url,
        ),
    )
    return runner.run()


if __name__ == "__main__":
    raise SystemExit(main())
