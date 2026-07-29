from __future__ import annotations

import argparse
import json
from pathlib import Path
from statistics import mean
from typing import Any


def summarize_multi_room(run_directory: Path) -> dict[str, Any]:
    run_directory = run_directory.resolve()
    run = _read_json(run_directory / "summary.json")
    rooms = run.get("rooms", [])
    clients = run.get("clients", {})
    if not isinstance(rooms, list) or not rooms:
        raise ValueError("run summary rooms must be a non-empty array")
    if not isinstance(clients, dict):
        raise ValueError("run summary clients must be an object")

    room_results = [
        _summarize_room(run_directory, room, clients)
        for room in rooms
        if isinstance(room, dict)
    ]
    if len(room_results) != len(rooms):
        raise ValueError("run summary contains an invalid room entry")

    room_codes = [str(room["room_code"]) for room in room_results]
    match_ids = [str(room["match_id"]) for room in room_results]
    return {
        "scenario_id": run.get("scenario_id", run_directory.name),
        "run_directory": str(run_directory),
        "success": bool(run.get("success", False)),
        "room_count": len(room_results),
        "unique_room_codes": len(set(room_codes)),
        "unique_match_ids": len(set(match_ids)),
        "rooms": room_results,
        "aggregate": _aggregate(room_results),
    }


def _summarize_room(
    run_directory: Path,
    room: dict[str, Any],
    clients: dict[str, Any],
) -> dict[str, Any]:
    coordinator_id = str(room.get("coordinator", "")).strip()
    room_code = str(room.get("room_code", "")).strip()
    if not coordinator_id or not room_code:
        raise ValueError("room summary is missing coordinator or room code")
    status = clients.get(coordinator_id, {})
    if not isinstance(status, dict) or status.get("state") != "completed":
        raise ValueError(f"{coordinator_id} did not complete")
    if str(status.get("room_code", "")) != room_code:
        raise ValueError(f"{coordinator_id} room code does not align")

    report_paths = status.get("measurement_reports", [])
    match_ids = status.get("match_ids", [])
    if not isinstance(report_paths, list) or len(report_paths) != 1:
        raise ValueError(f"{coordinator_id} must publish one measurement report")
    if not isinstance(match_ids, list) or len(match_ids) != 1:
        raise ValueError(f"{coordinator_id} must publish one match id")

    combined = _read_json(Path(str(report_paths[0])))
    client = combined.get("client", {})
    server = combined.get("server", {})
    if not isinstance(client, dict) or not isinstance(server, dict):
        raise ValueError(f"{coordinator_id} report is incomplete")
    server_export = server.get("server_export", {})
    if not isinstance(server_export, dict):
        raise ValueError(f"{coordinator_id} server export is missing")
    server_path = str(server_export.get("path", "")).strip()
    if not server_path:
        raise ValueError(f"{coordinator_id} server export path is missing")
    server_report = _read_json(_resolve_server_report_path(run_directory, server_path))

    return {
        "index": int(room.get("index", 0)),
        "room_code": room_code,
        "coordinator": coordinator_id,
        "participants": room.get("participants", []),
        "match_id": str(match_ids[0]),
        "server": _server_metrics(server_report),
        "client": _client_metrics(client),
    }


def _server_metrics(report: dict[str, Any]) -> dict[str, Any]:
    samples = report.get("samples", [])
    if not isinstance(samples, list):
        samples = []
    process_samples = [
        sample.get("process", {}) for sample in samples if isinstance(sample, dict)
    ]
    entity_samples = [
        sample.get("entities", {}) for sample in samples if isinstance(sample, dict)
    ]
    ticks = report.get("ticks", {})
    receiver = report.get("receiver", {})
    ticks = ticks if isinstance(ticks, dict) else {}
    receiver = receiver if isinstance(receiver, dict) else {}
    candidate = receiver.get("candidate_build_time", {})
    outbound = receiver.get("outbound_time", {})
    candidate = candidate if isinstance(candidate, dict) else {}
    outbound = outbound if isinstance(outbound, dict) else {}
    return {
        "tick_average_us": _round(float(ticks.get("average", 0)) / 1_000.0),
        "tick_maximum_ms": _round(float(ticks.get("maximum", 0)) / 1_000_000.0),
        "process_peak_rss_mib": _round(
            _max_number(process_samples, "peak_resident_set_bytes") / 1_048_576.0
        ),
        "process_peak_cpu_cores": _round(
            _max_number(process_samples, "cpu_utilization_cores")
        ),
        "max_player_sessions": int(_max_number(entity_samples, "player_sessions")),
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
        "receiver_skipped_send_ticks": int(receiver.get("skipped_send_ticks", 0)),
    }


def _client_metrics(report: dict[str, Any]) -> dict[str, Any]:
    frame = report.get("frame_timing", {})
    resources = report.get("resource_samples", {}).get("samples", [])
    frame = frame if isinstance(frame, dict) else {}
    resources = resources if isinstance(resources, list) else []
    network = report.get("network_metrics", {})
    network = network if isinstance(network, dict) else {}
    return {
        "frame_average_ms": _round(float(frame.get("average", 0))),
        "frame_p99_ms": _round(float(frame.get("p99", 0))),
        "peak_memory_mib": _round(
            _max_number(resources, "memory_bytes") / 1_048_576.0
        ),
        "send_failures": int(network.get("send_failures", 0)),
    }


def _aggregate(rooms: list[dict[str, Any]]) -> dict[str, Any]:
    servers = [room["server"] for room in rooms]
    clients = [room["client"] for room in rooms]
    return {
        "authoritative_participants": sum(
            int(server["max_player_sessions"]) for server in servers
        ),
        "summed_room_peak_asteroids": sum(
            int(server["max_asteroids"]) for server in servers
        ),
        "summed_room_peak_projectiles": sum(
            int(server["max_projectiles"]) for server in servers
        ),
        "process_peak_rss_mib": max(
            float(server["process_peak_rss_mib"]) for server in servers
        ),
        "process_peak_cpu_cores": max(
            float(server["process_peak_cpu_cores"]) for server in servers
        ),
        "mean_room_tick_average_us": _round(
            mean(float(server["tick_average_us"]) for server in servers)
        ),
        "maximum_room_tick_ms": max(
            float(server["tick_maximum_ms"]) for server in servers
        ),
        "mean_receiver_outbound_average_us": _round(
            mean(float(server["receiver_outbound_average_us"]) for server in servers)
        ),
        "maximum_receiver_outbound_ms": max(
            float(server["receiver_outbound_maximum_ms"]) for server in servers
        ),
        "receiver_skipped_send_ticks_total": sum(
            int(server["receiver_skipped_send_ticks"]) for server in servers
        ),
        "client_send_failures_total": sum(
            int(client["send_failures"]) for client in clients
        ),
        "worst_client_frame_p99_ms": max(
            float(client["frame_p99_ms"]) for client in clients
        ),
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
        description="Summarize one simultaneous multi-room runtime scenario."
    )
    parser.add_argument("run_directory", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    result = summarize_multi_room(args.run_directory)
    encoded = json.dumps(result, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(encoded, encoding="utf-8")
    print(encoded, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
