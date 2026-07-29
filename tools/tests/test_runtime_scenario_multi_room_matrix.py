from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.multi_room_matrix_manifest import MatrixManifest
from runtime_scenarios.multi_room_matrix_summary import summarize_matrix


def _scenario_payload(room_count: int) -> dict[str, object]:
    return {
        "id": f"matrix-{room_count}",
        "seed": 17,
        "timeout_seconds": 30,
        "setup_timeout_seconds": 5,
        "room_count": room_count,
        "clients": {"visible": 1, "headless": 0},
        "bots": 7,
        "rounds": [
            {
                "name": "cycle",
                "lives": 2,
                "setup": {"asteroid_spawns": 24, "settle_seconds": 1},
                "phases": [
                    {"name": "pressure", "duration_seconds": 1, "bullet_streams": 20}
                ],
            }
        ],
    }


def _write_manifest(root: Path, *, alter_fourth: bool = False) -> Path:
    entries = []
    for room_count in range(1, 5):
        payload = _scenario_payload(room_count)
        if alter_fourth and room_count == 4:
            payload["bots"] = 6
        scenario_path = root / f"scenario-{room_count}.json"
        scenario_path.write_text(json.dumps(payload), encoding="utf-8")
        entries.append({"room_count": room_count, "path": scenario_path.name})
    manifest_path = root / "matrix.json"
    manifest_path.write_text(
        json.dumps({"id": "matrix-test", "scenarios": entries}),
        encoding="utf-8",
    )
    return manifest_path


def _write_run(
    root: Path,
    scenario_path: Path,
    room_count: int,
    *,
    controlled: bool,
) -> Path:
    run_directory = root / f"run-{room_count}"
    run_directory.mkdir()
    rooms: list[dict[str, object]] = []
    clients: dict[str, object] = {}
    for index in range(1, room_count + 1):
        coordinator = "coordinator-1" if room_count == 1 else f"room-{index}-coordinator"
        room_code = f"R{room_count}{index}"
        match_id = f"{room_code}-match-1"
        server_path = run_directory / f"server-{index}.json"
        server_path.write_text(
            json.dumps(
                {
                    "ticks": {"average": room_count * 100_000, "maximum": 2_000_000},
                    "samples": [
                        {
                            "process": {
                                "peak_resident_set_bytes": (30 + room_count) * 1_048_576,
                                "cpu_utilization_cores": room_count / 2,
                            },
                            "entities": {
                                "player_sessions": 8,
                                "asteroids": 24,
                                "projectiles": 40,
                            },
                        }
                    ],
                    "receiver": {
                        "candidate_build_time": {"average": 200_000, "maximum": 1_000_000},
                        "outbound_time": {"average": 300_000, "maximum": 1_500_000},
                        "skipped_send_ticks": 0,
                    },
                }
            ),
            encoding="utf-8",
        )
        combined_path = run_directory / f"combined-{index}.json"
        combined_path.write_text(
            json.dumps(
                {
                    "client": {
                        "frame_timing": {"average": 7.0, "p99": 16.0},
                        "resource_samples": {
                            "samples": [{"memory_bytes": 256 * 1_048_576}]
                        },
                        "network_metrics": {"send_failures": 0},
                    },
                    "server": {"server_export": {"path": str(server_path)}},
                }
            ),
            encoding="utf-8",
        )
        rooms.append(
            {
                "index": index,
                "room_code": room_code,
                "coordinator": coordinator,
                "participants": [],
            }
        )
        clients[coordinator] = {
            "state": "completed",
            "room_code": room_code,
            "rounds_completed": 1,
            "match_ids": [match_id],
            "measurement_reports": [str(combined_path)],
        }
    scenario = json.loads(scenario_path.read_text(encoding="utf-8"))
    (run_directory / "summary.json").write_text(
        json.dumps(
            {
                "scenario_id": scenario["id"],
                "scenario_path": str(scenario_path.resolve()),
                "success": True,
                "host_control": {
                    "controlled": controlled,
                    "note": "isolated host" if controlled else "development host",
                },
                "rooms": rooms,
                "clients": clients,
            }
        ),
        encoding="utf-8",
    )
    return run_directory


def test_loads_repository_matrix_manifest() -> None:
    repo_root = Path(__file__).resolve().parents[2]
    manifest = MatrixManifest.load(
        repo_root
        / "tools"
        / "runtime_scenarios"
        / "scenarios"
        / "multi_room_matrix_v1.json"
    )
    assert [entry.room_count for entry in manifest.entries] == [1, 2, 3, 4]
    assert len({entry.scenario.seed for entry in manifest.entries}) == 1


def test_rejects_matrix_workload_mismatch(tmp_path: Path) -> None:
    with pytest.raises(ValueError, match="same workload"):
        MatrixManifest.load(_write_manifest(tmp_path, alter_fourth=True))


@pytest.mark.parametrize("controlled", [False, True])
def test_matrix_summary_gates_performance_eligibility(
    tmp_path: Path,
    controlled: bool,
) -> None:
    manifest = MatrixManifest.load(_write_manifest(tmp_path))
    runs = [
        _write_run(tmp_path, entry.scenario.path, entry.room_count, controlled=controlled)
        for entry in manifest.entries
    ]

    result = summarize_matrix(manifest, runs)

    assert result["functional_pass"] is True
    assert result["performance_eligible"] is controlled
    assert result["host_control"]["all_controlled"] is controlled
    assert [entry["room_count"] for entry in result["results"]] == [1, 2, 3, 4]
