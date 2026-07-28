from __future__ import annotations

import json
import sys
from pathlib import Path

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.churn_summary import summarize_churn


def _write(path: Path, payload: dict[str, object]) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload), encoding="utf-8")
    return path


def test_summarizes_per_round_drift(tmp_path: Path) -> None:
    run_directory = tmp_path / "run"
    server_reports: list[Path] = []
    combined_reports: list[Path] = []
    for round_number, rss_mib in [(1, 40), (2, 43)]:
        server_path = _write(
            run_directory / f"server-{round_number}.json",
            {
                "ticks": {"average": 200_000 + round_number * 1_000, "maximum": 2_000_000},
                "samples": [
                    {
                        "process": {
                            "peak_resident_set_bytes": rss_mib * 1_048_576,
                            "cpu_utilization_cores": 0.5,
                        },
                        "entities": {
                            "player_sessions": 8,
                            "asteroids": 50,
                            "projectiles": 100,
                        },
                    }
                ],
                "receiver": {
                    "candidate_build_time": {"average": 500_000, "maximum": 2_000_000},
                    "outbound_time": {"average": 700_000, "maximum": 3_000_000},
                    "skipped_send_ticks": 0,
                },
            },
        )
        server_reports.append(server_path)
        combined_reports.append(
            _write(
                run_directory / f"combined-{round_number}.json",
                {
                    "client": {
                        "metadata": {
                            "round": round_number,
                            "round_name": f"cycle-{round_number}",
                            "match_id": f"room-match-{round_number}",
                        },
                        "frame_timing": {"average": 7 + round_number, "p99": 16, "maximum": 30},
                        "resource_samples": {
                            "samples": [{"memory_bytes": (200 + round_number) * 1_048_576}]
                        },
                        "network_metrics": {"send_failures": 0},
                    },
                    "server": {"server_export": {"path": str(server_path)}},
                },
            )
        )

    _write(
        run_directory / "summary.json",
        {
            "scenario_id": "match_churn",
            "success": True,
            "clients": {
                "coordinator-1": {
                    "measurement_reports": [str(path) for path in combined_reports],
                    "match_ids": ["room-match-1", "room-match-2"],
                }
            },
        },
    )

    result = summarize_churn(run_directory)
    assert result["rounds_completed"] == 2
    assert result["unique_match_ids"] == 2
    assert result["drift"]["server_peak_rss_mib"] == 3.0
    assert result["drift"]["client_peak_memory_mib"] == 1.0
    assert result["rounds"][1]["server"]["receiver_skipped_send_ticks"] == 0
