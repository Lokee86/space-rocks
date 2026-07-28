from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


def summarize_churn(run_directory: Path) -> dict[str, Any]:
    run_directory = run_directory.resolve()
    summary = _read_json(run_directory / "summary.json")
    clients = summary.get("clients", {})
    if not isinstance(clients, dict):
        raise ValueError("run summary clients must be an object")
    coordinator = clients.get("coordinator-1", {})
    if not isinstance(coordinator, dict):
        raise ValueError("run summary is missing coordinator status")
    report_paths = coordinator.get("measurement_reports", [])
    match_ids = coordinator.get("match_ids", [])
    if not isinstance(report_paths, list) or not report_paths:
        raise ValueError("coordinator did not publish per-round measurement reports")
    if not isinstance(match_ids, list) or len(match_ids) != len(report_paths):
        raise ValueError("coordinator match ids do not align with reports")

    rounds: list[dict[str, Any]] = []
    for expected_round, report_value in enumerate(report_paths, start=1):
        combined = _read_json(Path(str(report_value)))
        client = combined.get("client", {})
        server = combined.get("server", {})
        if not isinstance(client, dict) or not isinstance(server, dict):
            raise ValueError(f"round {expected_round} report is incomplete")
        metadata = client.get("metadata", {})
        if not isinstance(metadata, dict):
            metadata = {}
        server_export = server.get("server_export", {})
        if not isinstance(server_export, dict):
            raise ValueError(f"round {expected_round} server export is missing")
        server_path = str(server_export.get("path", "")).strip()
        if not server_path:
            raise ValueError(f"round {expected_round} server export path is missing")
        server_report = _read_json(_resolve_server_report_path(run_directory, server_path))
        rounds.append(
            _summarize_round(
                expected_round,
                str(match_ids[expected_round - 1]),
                metadata,
                client,
                server_report,
            )
        )

    first = rounds[0]
    last = rounds[-1]
    return {
        "scenario_id": summary.get("scenario_id", run_directory.name),
        "run_directory": str(run_directory),
        "success": bool(summary.get("success", False)),
        "rounds_completed": len(rounds),
        "unique_match_ids": len(set(str(value) for value in match_ids)),
        "rounds": rounds,
        "drift": {
            "server_peak_rss_mib": _round(
                float(last["server"]["peak_rss_mib"])
                - float(first["server"]["peak_rss_mib"])
            ),
            "server_tick_average_us": _round(
                float(last["server"]["tick_average_us"])
                - float(first["server"]["tick_average_us"])
            ),
            "client_peak_memory_mib": _round(
                float(last["client"]["peak_memory_mib"])
                - float(first["client"]["peak_memory_mib"])
            ),
            "client_frame_average_ms": _round(
                float(last["client"]["frame_average_ms"])
                - float(first["client"]["frame_average_ms"])
            ),
        },
    }


def _summarize_round(
    expected_round: int,
    expected_match_id: str,
    metadata: dict[str, Any],
    client: dict[str, Any],
    server: dict[str, Any],
) -> dict[str, Any]:
    round_number = int(metadata.get("round", expected_round))
    match_id = str(metadata.get("match_id", expected_match_id))
    if round_number != expected_round:
        raise ValueError(f"expected round {expected_round}, got {round_number}")
    if match_id != expected_match_id:
        raise ValueError(f"round {expected_round} match id does not align")

    samples = server.get("samples", [])
    if not isinstance(samples, list):
        samples = []
    process_samples = [
        sample.get("process", {}) for sample in samples if isinstance(sample, dict)
    ]
    final_process = process_samples[-1] if process_samples else {}
    entity_samples = [
        sample.get("entities", {}) for sample in samples if isinstance(sample, dict)
    ]
    resource_samples = client.get("resource_samples", {}).get("samples", [])
    if not isinstance(resource_samples, list):
        resource_samples = []
    ticks = server.get("ticks", {})
    if not isinstance(ticks, dict):
        ticks = {}
    receiver = server.get("receiver", {})
    if not isinstance(receiver, dict):
        receiver = {}
    candidate = receiver.get("candidate_build_time", {})
    outbound = receiver.get("outbound_time", {})
    if not isinstance(candidate, dict):
        candidate = {}
    if not isinstance(outbound, dict):
        outbound = {}
    frame = client.get("frame_timing", {})
    if not isinstance(frame, dict):
        frame = {}

    return {
        "round": round_number,
        "name": str(metadata.get("round_name", "")),
        "match_id": match_id,
        "server": {
            "tick_average_us": _round(float(ticks.get("average", 0)) / 1_000.0),
            "tick_maximum_ms": _round(float(ticks.get("maximum", 0)) / 1_000_000.0),
            "peak_rss_mib": _round(
                _max_number(process_samples, "peak_resident_set_bytes") / 1_048_576.0
            ),
            "resident_set_mib": _round(
                float(final_process.get("resident_set_bytes", 0)) / 1_048_576.0
            ),
            "heap_allocated_mib": _round(
                float(final_process.get("heap_allocated_bytes", 0)) / 1_048_576.0
            ),
            "heap_in_use_mib": _round(
                float(final_process.get("heap_in_use_bytes", 0)) / 1_048_576.0
            ),
            "goroutines": int(float(final_process.get("goroutines", 0))),
            "gc_cycles": int(float(final_process.get("gc_cycles", 0))),
            "max_cpu_cores": _round(
                _max_number(process_samples, "cpu_utilization_cores")
            ),
            "max_player_sessions": int(
                _max_number(entity_samples, "player_sessions")
            ),
            "max_asteroids": int(_max_number(entity_samples, "asteroids")),
            "max_projectiles": int(_max_number(entity_samples, "projectiles")),
            "receiver_candidate_average_us": _round(
                float(candidate.get("average", 0)) / 1_000.0
            ),
            "receiver_candidate_maximum_ms": _round(
                float(candidate.get("maximum", 0)) / 1_000_000.0
            ),
            "receiver_outbound_average_us": _round(
                float(outbound.get("average", 0)) / 1_000.0
            ),
            "receiver_outbound_maximum_ms": _round(
                float(outbound.get("maximum", 0)) / 1_000_000.0
            ),
            "receiver_skipped_send_ticks": int(
                float(receiver.get("skipped_send_ticks", 0))
            ),
        },
        "client": {
            "frame_average_ms": _round(float(frame.get("average", 0))),
            "frame_p99_ms": _round(float(frame.get("p99", 0))),
            "frame_maximum_ms": _round(float(frame.get("maximum", 0))),
            "peak_memory_mib": _round(
                _max_number(resource_samples, "memory_bytes") / 1_048_576.0
            ),
            "send_failures": int(
                float(client.get("network_metrics", {}).get("send_failures", 0))
            ),
        },
    }


def _resolve_server_report_path(run_directory: Path, value: str) -> Path:
    path = Path(value)
    if path.is_absolute():
        return path
    repo_root = run_directory.parents[2]
    return (repo_root / "services" / "game-server" / path).resolve()


def _read_json(path: Path) -> dict[str, Any]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(payload, dict):
        raise ValueError(f"expected JSON object: {path}")
    return payload


def _max_number(items: list[dict[str, Any]], key: str) -> float:
    return max(
        (float(item.get(key, 0)) for item in items if isinstance(item, dict)),
        default=0.0,
    )


def _round(value: float) -> float:
    return round(value, 3)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Summarize per-round match-churn runtime measurements."
    )
    parser.add_argument("run_directory", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    result = summarize_churn(args.run_directory)
    encoded = json.dumps(result, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(encoded, encoding="utf-8")
    print(encoded, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
