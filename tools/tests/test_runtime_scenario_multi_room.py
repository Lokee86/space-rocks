from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.model import Scenario, ScenarioError
from runtime_scenarios.multi_room_summary import summarize_multi_room
from runtime_scenarios.room_groups import coordinator_client_id, participant_client_id


def _write_scenario(path: Path, room_count: int) -> Path:
    path.write_text(
        json.dumps(
            {
                "id": "multi-room-test",
                "seed": 11,
                "room_count": room_count,
                "clients": {"visible": 1, "headless": 1},
                "bots": 6,
                "phases": [{"name": "pressure", "duration_seconds": 1}],
            }
        ),
        encoding="utf-8",
    )
    return path


def test_loads_room_count(tmp_path: Path) -> None:
    scenario = Scenario.load(_write_scenario(tmp_path / "scenario.json", 3))
    assert scenario.room_count == 3
    assert scenario.clients.total == 2


def test_rejects_nonpositive_room_count(tmp_path: Path) -> None:
    with pytest.raises(ScenarioError, match="room_count"):
        Scenario.load(_write_scenario(tmp_path / "scenario.json", 0))


def test_multi_room_client_ids_are_room_scoped() -> None:
    assert coordinator_client_id(1, 1) == "coordinator-1"
    assert participant_client_id(1, 1, 2) == "participant-2"
    assert coordinator_client_id(3, 2) == "room-2-coordinator"
    assert participant_client_id(3, 2, 1) == "room-2-participant-1"


def test_summarizes_independent_room_reports(tmp_path: Path) -> None:
    run_directory = tmp_path / "multi-room-run"
    run_directory.mkdir()
    rooms = []
    clients = {}
    for index in (1, 2):
        coordinator = f"room-{index}-coordinator"
        room_code = f"ROOM{index}"
        match_id = f"match-{index}"
        server_path = run_directory / f"server-{index}.json"
        server_path.write_text(
            json.dumps(
                {
                    "ticks": {"average": 100_000 + index, "maximum": 2_000_000},
                    "samples": [
                        {
                            "process": {
                                "peak_resident_set_bytes": (30 + index) * 1_048_576,
                                "cpu_utilization_cores": 0.5 + index,
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
                        "skipped_send_ticks": index - 1,
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
                        "network_metrics": {"send_failures": index - 1},
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
    (run_directory / "summary.json").write_text(
        json.dumps(
            {
                "scenario_id": "multi-room-test",
                "success": True,
                "rooms": rooms,
                "clients": clients,
            }
        ),
        encoding="utf-8",
    )

    result = summarize_multi_room(run_directory)

    assert result["room_count"] == 2
    assert result["unique_room_codes"] == 2
    assert result["unique_match_ids"] == 2
    assert result["aggregate"]["authoritative_participants"] == 16
    assert result["aggregate"]["process_peak_rss_mib"] == 32.0
    assert result["aggregate"]["receiver_skipped_send_ticks_total"] == 1
    assert result["aggregate"]["client_send_failures_total"] == 1
