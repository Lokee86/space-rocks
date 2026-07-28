from __future__ import annotations

import json
import sys
from pathlib import Path

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.matrix_summary import _resolve_server_report_path, summarize_run


def test_resolves_wsl_relative_server_report_from_game_server_root(tmp_path: Path) -> None:
    run = tmp_path / "repo" / ".ci-artifacts" / "runtime-scenarios" / "run"
    expected = (
        tmp_path
        / "repo"
        / ".ci-artifacts"
        / "runtime-scenarios"
        / "run"
        / "measurements"
        / "server.json"
    ).resolve()
    result = _resolve_server_report_path(
        run,
        "../../.ci-artifacts/runtime-scenarios/run/measurements/server.json",
    )
    assert result == expected


def test_summarize_run_aggregates_receiver_reports(tmp_path: Path) -> None:
    run = tmp_path / "run"
    run.mkdir()
    clients: dict[str, dict[str, str]] = {}
    for index, encoded_bytes in enumerate((1000, 2000), start=1):
        client_id = f"client-{index}"
        server_path = run / f"server-{index}.json"
        _write_json(server_path, _server_report(encoded_bytes))
        combined_path = run / f"combined-{index}.json"
        _write_json(combined_path, _combined_report(server_path, index))
        clients[client_id] = {"measurement_report": str(combined_path)}
    _write_json(
        run / "summary.json",
        {
            "scenario_id": "matrix-test",
            "seed": 7,
            "success": True,
            "execution": {"coordinator_headless": True},
            "phase_markers": [
                {
                    "name": "pressure",
                    "start_seconds": 0.0,
                    "end_seconds": 10.0,
                    "duration_seconds": 10.0,
                }
            ],
            "clients": clients,
        },
    )

    result = summarize_run(run)

    assert result["real_clients"] == 2
    assert result["server"]["receiver_bytes_total"] == 3000
    assert result["server"]["tick_average_us"] == 25.0
    assert result["clients"]["bytes_in_total"] == 300
    assert result["server"]["receiver_skipped_send_ticks_total"] == 3
    assert result["server"]["receiver_candidate_build_average_us_mean"] == 3.0
    phases = result["server"]["receiver_candidate_build_phases"]
    assert phases["snapshot_capture"] == {"average_us_mean": 0.3, "maximum_ms": 0.001}
    assert phases["lane_candidates"] == {"average_us_mean": 1.5, "maximum_ms": 0.003}
    lane_phases = result["server"]["receiver_lane_candidate_phases"]
    assert lane_phases["world_hot_lifecycle"] == {
        "average_us_mean": 0.9,
        "maximum_ms": 0.002,
    }
    assert lane_phases["overlay"] == {"average_us_mean": 0.15, "maximum_ms": 0.0}
    assert result["server"]["receiver_candidate_build_peak"] == {
        "total_ms": 0.009,
        "phases_ms": {
            "snapshot_capture": 0.001,
            "pending_event_copy": 0.0,
            "interest_filter": 0.001,
            "lane_candidates": 0.004,
            "chunk_planning": 0.002,
            "scheduling": 0.001,
        },
        "lane_candidate_phases_ms": {
            "state_advance": 0.0,
            "world_hot_lifecycle": 0.002,
            "player_locator": 0.0,
            "overlay": 0.0,
            "session": 0.001,
            "event": 0.0,
            "candidate_finalize": 0.0,
        },
    }
    assert result["server"]["receiver_encoding_maximum_ms"] == 0.006
    assert result["server"]["receiver_lane_peak_buffered_bytes"] == {"world": 200}
    assert result["server"]["receiver_lane_skipped_send_ticks"] == {"world": 3}
    assert result["phase_markers"][0]["name"] == "pressure"
    assert result["coordinator_headless"] is True


def _combined_report(server_path: Path, index: int) -> dict[str, object]:
    return {
        "client": {
            "frame_timing": {"average": 7 + index, "p95": 16, "p99": 20, "maximum": 30},
            "presentation_timing": {"average": 2, "p99": 5},
            "network_metrics": {"bytes_in": 100 * index, "packets_in": 10, "send_failures": 0},
            "resource_samples": {"samples": [{"memory_bytes": 104857600 * index}]},
        },
        "server": {"server_export": {"path": str(server_path)}},
    }


def _server_report(encoded_bytes: int) -> dict[str, object]:
    return {
        "ticks": {"average": 25000, "maximum": 2000000},
        "samples": [
            {
                "process": {
                    "peak_resident_set_bytes": 83886080,
                    "cpu_utilization_cores": 0.5,
                },
                "entities": {"player_sessions": 8, "asteroids": 100, "projectiles": 200},
            }
        ],
        "packets": [
            {
                "packet_count": 5,
                "encoded_bytes_total": encoded_bytes,
                "maximum_encoded_bytes": 500,
            }
        ],
        "receiver": {
            "tick_count": 10,
            "skipped_send_ticks": 1 if encoded_bytes == 1000 else 2,
            "candidate_build_time": {
                "count": 10,
                "average": 2000 if encoded_bytes == 1000 else 4000,
                "maximum": 7000,
            },
            "candidate_build_phases": {
                "snapshot_capture_time": {
                    "count": 10,
                    "average": 200 if encoded_bytes == 1000 else 400,
                    "maximum": 800 if encoded_bytes == 1000 else 1000,
                },
                "pending_event_copy_time": {"count": 10, "average": 20, "maximum": 40},
                "interest_filter_time": {"count": 10, "average": 300, "maximum": 600},
                "lane_candidates_time": {
                    "count": 10,
                    "average": 1000 if encoded_bytes == 1000 else 2000,
                    "maximum": 2000 if encoded_bytes == 1000 else 3000,
                },
                "lane_candidate_phases": {
                    "state_advance_time": {"count": 10, "average": 20, "maximum": 40},
                    "world_hot_lifecycle_time": {
                        "count": 10,
                        "average": 600 if encoded_bytes == 1000 else 1200,
                        "maximum": 1200 if encoded_bytes == 1000 else 1800,
                    },
                    "player_locator_time": {"count": 10, "average": 40, "maximum": 60},
                    "overlay_time": {
                        "count": 10,
                        "average": 100 if encoded_bytes == 1000 else 200,
                        "maximum": 200 if encoded_bytes == 1000 else 300,
                    },
                    "session_time": {"count": 10, "average": 300, "maximum": 600},
                    "event_time": {"count": 10, "average": 50, "maximum": 100},
                    "candidate_finalize_time": {"count": 10, "average": 30, "maximum": 50},
                },
                "chunk_planning_time": {"count": 10, "average": 100, "maximum": 200},
                "scheduling_time": {"count": 10, "average": 200, "maximum": 400},
            },
            "candidate_build_peak": {
                "total": 8000 if encoded_bytes == 1000 else 9000,
                "phases": {
                    "snapshot_capture_duration": 500,
                    "pending_event_copy_duration": 20,
                    "interest_filter_duration": 600,
                    "lane_candidates_duration": 4000,
                    "lane_candidate_phases": {
                        "state_advance_duration": 20,
                        "world_hot_lifecycle_duration": 2400,
                        "player_locator_duration": 40,
                        "overlay_duration": 200,
                        "session_duration": 1000,
                        "event_duration": 100,
                        "candidate_finalize_duration": 40,
                    },
                    "chunk_planning_duration": 2000,
                    "scheduling_duration": 1000,
                },
            },
            "encoding_time": {
                "count": 10,
                "average": 1000,
                "maximum": 5000 if encoded_bytes == 1000 else 6000,
            },
            "outbound_time": {
                "count": 10,
                "average": 8000,
                "maximum": 12000,
            },
            "lanes": [
                {
                    "lane": "world",
                    "sample_count": 10,
                    "current_buffered_bytes": 50,
                    "peak_buffered_bytes": 100 if encoded_bytes == 1000 else 200,
                    "skipped_send_ticks": 1 if encoded_bytes == 1000 else 2,
                }
            ],
        },
    }


def _write_json(path: Path, payload: dict[str, object]) -> None:
    path.write_text(json.dumps(payload), encoding="utf-8")
