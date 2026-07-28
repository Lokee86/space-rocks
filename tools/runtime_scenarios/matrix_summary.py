from __future__ import annotations

import argparse
import json
from pathlib import Path
from statistics import mean
from typing import Any


def summarize_run(run_directory: Path) -> dict[str, Any]:
    summary = _read_json(run_directory / "summary.json")
    client_reports: list[dict[str, Any]] = []
    server_reports: list[dict[str, Any]] = []

    clients = summary.get("clients", {})
    if not isinstance(clients, dict):
        clients = {}
    for client_id in sorted(clients):
        status = clients[client_id]
        if not isinstance(status, dict):
            continue
        report_path = str(status.get("measurement_report", "")).strip()
        if not report_path:
            continue
        combined = _read_json(Path(report_path))
        client = combined.get("client", {})
        if isinstance(client, dict):
            client_reports.append(client)
        server = combined.get("server", {})
        if not isinstance(server, dict):
            continue
        export = server.get("server_export", {})
        if not isinstance(export, dict):
            continue
        server_path = str(export.get("path", "")).strip()
        if server_path:
            server_reports.append(
                _read_json(_resolve_server_report_path(run_directory, server_path))
            )

    if not client_reports or not server_reports:
        raise ValueError(f"run has incomplete measurement exports: {run_directory}")

    reference_server = server_reports[0]
    samples = reference_server.get("samples", [])
    if not isinstance(samples, list):
        samples = []
    process_samples = [sample.get("process", {}) for sample in samples if isinstance(sample, dict)]
    entity_samples = [sample.get("entities", {}) for sample in samples if isinstance(sample, dict)]

    frame_reports = [report.get("frame_timing", {}) for report in client_reports]
    presentation_reports = [report.get("presentation_timing", {}) for report in client_reports]
    networks = [report.get("network_metrics", {}) for report in client_reports]
    resource_samples = [
        sample
        for report in client_reports
        for sample in report.get("resource_samples", {}).get("samples", [])
        if isinstance(sample, dict)
    ]

    packet_summaries = [
        packet
        for report in server_reports
        for packet in report.get("packets", [])
        if isinstance(packet, dict)
    ]
    receiver_reports = [
        receiver
        for report in server_reports
        for receiver in [report.get("receiver", {})]
        if isinstance(receiver, dict)
    ]
    receiver_lanes = [
        lane
        for receiver in receiver_reports
        for lane in receiver.get("lanes", [])
        if isinstance(lane, dict)
    ]
    ticks = reference_server.get("ticks", {})
    tick_average_us = float(ticks.get("average", 0)) / 1_000.0
    receiver_outbound_average_us = (
        _mean_duration_field(receiver_reports, "outbound_time", "average") / 1_000.0
    )

    execution = summary.get("execution", {})
    coordinator_headless = (
        bool(execution.get("coordinator_headless"))
        if isinstance(execution, dict) and "coordinator_headless" in execution
        else None
    )

    return {
        "scenario_id": summary.get("scenario_id", run_directory.name),
        "run_directory": str(run_directory),
        "success": bool(summary.get("success", False)),
        "real_clients": len(client_reports),
        "seed": summary.get("seed"),
        "coordinator_headless": coordinator_headless,
        "phase_markers": summary.get("phase_markers", []),
        "server": {
            "tick_average_us": _round(tick_average_us),
            "tick_maximum_ms": _round(float(ticks.get("maximum", 0)) / 1_000_000.0),
            "peak_rss_mib": _round(
                _max_number(process_samples, "peak_resident_set_bytes") / 1_048_576.0
            ),
            "max_cpu_cores": _round(_max_number(process_samples, "cpu_utilization_cores")),
            "max_player_sessions": int(_max_number(entity_samples, "player_sessions")),
            "max_asteroids": int(_max_number(entity_samples, "asteroids")),
            "max_projectiles": int(_max_number(entity_samples, "projectiles")),
            "receiver_packet_count_total": int(
                sum(float(packet.get("packet_count", 0)) for packet in packet_summaries)
            ),
            "receiver_bytes_total": int(
                sum(float(packet.get("encoded_bytes_total", 0)) for packet in packet_summaries)
            ),
            "receiver_max_packet_bytes": int(
                max(
                    (
                        float(packet.get("maximum_encoded_bytes", 0))
                        for packet in packet_summaries
                    ),
                    default=0,
                )
            ),
            "receiver_tick_count_total": int(
                sum(float(receiver.get("tick_count", 0)) for receiver in receiver_reports)
            ),
            "receiver_skipped_send_ticks_total": int(
                sum(
                    float(receiver.get("skipped_send_ticks", 0))
                    for receiver in receiver_reports
                )
            ),
            "receiver_candidate_build_average_us_mean": _round(
                _mean_duration_field(receiver_reports, "candidate_build_time", "average")
                / 1_000.0
            ),
            "receiver_candidate_build_maximum_ms": _round(
                _max_duration_field(receiver_reports, "candidate_build_time", "maximum")
                / 1_000_000.0
            ),
            "receiver_candidate_build_phases": _candidate_build_phase_summaries(
                receiver_reports
            ),
            "receiver_lane_candidate_phases": _lane_candidate_phase_summaries(
                receiver_reports
            ),
            "receiver_candidate_build_peak": _candidate_build_peak(receiver_reports),
            "receiver_encoding_average_us_mean": _round(
                _mean_duration_field(receiver_reports, "encoding_time", "average") / 1_000.0
            ),
            "receiver_encoding_maximum_ms": _round(
                _max_duration_field(receiver_reports, "encoding_time", "maximum")
                / 1_000_000.0
            ),
            "receiver_outbound_average_us_mean": _round(receiver_outbound_average_us),
            "receiver_outbound_maximum_ms": _round(
                _max_duration_field(receiver_reports, "outbound_time", "maximum")
                / 1_000_000.0
            ),
            "receiver_outbound_to_simulation_average_ratio": (
                _round(receiver_outbound_average_us / tick_average_us)
                if tick_average_us > 0
                else 0.0
            ),
            "receiver_lane_peak_buffered_bytes": _lane_maximums(
                receiver_lanes, "peak_buffered_bytes"
            ),
            "receiver_lane_skipped_send_ticks": _lane_totals(
                receiver_lanes, "skipped_send_ticks"
            ),
        },
        "clients": {
            "frame_average_ms_mean": _round(_mean_field(frame_reports, "average")),
            "frame_p95_ms_max": _round(_max_number(frame_reports, "p95")),
            "frame_p99_ms_max": _round(_max_number(frame_reports, "p99")),
            "frame_maximum_ms": _round(_max_number(frame_reports, "maximum")),
            "presentation_average_ms_mean": _round(
                _mean_field(presentation_reports, "average")
            ),
            "presentation_p99_ms_max": _round(
                _max_number(presentation_reports, "p99")
            ),
            "bytes_in_total": int(
                sum(float(network.get("bytes_in", 0)) for network in networks)
            ),
            "packets_in_total": int(
                sum(float(network.get("packets_in", 0)) for network in networks)
            ),
            "send_failures_total": int(
                sum(float(network.get("send_failures", 0)) for network in networks)
            ),
            "peak_memory_mib": _round(
                _max_number(resource_samples, "memory_bytes") / 1_048_576.0
            ),
        },
    }


def summarize_matrix(run_directories: list[Path]) -> dict[str, Any]:
    return {"runs": [summarize_run(path.resolve()) for path in run_directories]}


def _resolve_server_report_path(run_directory: Path, value: str) -> Path:
    path = Path(value)
    if path.is_absolute():
        return path
    repo_root = run_directory.resolve().parents[2]
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


def _mean_field(items: list[dict[str, Any]], key: str) -> float:
    values = [float(item.get(key, 0)) for item in items if isinstance(item, dict)]
    return mean(values) if values else 0.0


def _mean_duration_field(
    items: list[dict[str, Any]], summary_key: str, field: str
) -> float:
    values = [
        float(summary.get(field, 0))
        for item in items
        for summary in [item.get(summary_key, {})]
        if isinstance(summary, dict)
    ]
    return mean(values) if values else 0.0


def _max_duration_field(
    items: list[dict[str, Any]], summary_key: str, field: str
) -> float:
    return max(
        (
            float(summary.get(field, 0))
            for item in items
            for summary in [item.get(summary_key, {})]
            if isinstance(summary, dict)
        ),
        default=0.0,
    )


def _candidate_build_peak(receivers: list[dict[str, Any]]) -> dict[str, Any]:
    peak: dict[str, Any] = {}
    peak_total = 0.0
    for receiver in receivers:
        candidate = receiver.get("candidate_build_peak", {})
        if not isinstance(candidate, dict):
            continue
        total = float(candidate.get("total", 0))
        if total <= peak_total:
            continue
        phases = candidate.get("phases", {})
        peak_total = total
        lane_phases = phases.get("lane_candidate_phases", {})
        if not isinstance(lane_phases, dict):
            lane_phases = {}
        peak = {
            "total_ms": _round(total / 1_000_000.0),
            "phases_ms": {
                "snapshot_capture": _round(
                    float(phases.get("snapshot_capture_duration", 0)) / 1_000_000.0
                ),
                "pending_event_copy": _round(
                    float(phases.get("pending_event_copy_duration", 0)) / 1_000_000.0
                ),
                "interest_filter": _round(
                    float(phases.get("interest_filter_duration", 0)) / 1_000_000.0
                ),
                "lane_candidates": _round(
                    float(phases.get("lane_candidates_duration", 0)) / 1_000_000.0
                ),
                "chunk_planning": _round(
                    float(phases.get("chunk_planning_duration", 0)) / 1_000_000.0
                ),
                "scheduling": _round(
                    float(phases.get("scheduling_duration", 0)) / 1_000_000.0
                ),
            },
            "lane_candidate_phases_ms": {
                "state_advance": _round(
                    float(lane_phases.get("state_advance_duration", 0)) / 1_000_000.0
                ),
                "world_hot_lifecycle": _round(
                    float(lane_phases.get("world_hot_lifecycle_duration", 0))
                    / 1_000_000.0
                ),
                "player_locator": _round(
                    float(lane_phases.get("player_locator_duration", 0)) / 1_000_000.0
                ),
                "overlay": _round(
                    float(lane_phases.get("overlay_duration", 0)) / 1_000_000.0
                ),
                "session": _round(
                    float(lane_phases.get("session_duration", 0)) / 1_000_000.0
                ),
                "event": _round(
                    float(lane_phases.get("event_duration", 0)) / 1_000_000.0
                ),
                "candidate_finalize": _round(
                    float(lane_phases.get("candidate_finalize_duration", 0))
                    / 1_000_000.0
                ),
            },
        }
    return peak


def _candidate_build_phase_summaries(
    receivers: list[dict[str, Any]],
) -> dict[str, dict[str, float]]:
    phase_keys = {
        "snapshot_capture": "snapshot_capture_time",
        "pending_event_copy": "pending_event_copy_time",
        "interest_filter": "interest_filter_time",
        "lane_candidates": "lane_candidates_time",
        "chunk_planning": "chunk_planning_time",
        "scheduling": "scheduling_time",
    }
    result: dict[str, dict[str, float]] = {}
    for output_name, phase_key in phase_keys.items():
        summaries = [
            summary
            for receiver in receivers
            for phases in [receiver.get("candidate_build_phases", {})]
            if isinstance(phases, dict)
            for summary in [phases.get(phase_key, {})]
            if isinstance(summary, dict)
        ]
        result[output_name] = {
            "average_us_mean": _round(_mean_field(summaries, "average") / 1_000.0),
            "maximum_ms": _round(_max_number(summaries, "maximum") / 1_000_000.0),
        }
    return result


def _lane_candidate_phase_summaries(
    receivers: list[dict[str, Any]],
) -> dict[str, dict[str, float]]:
    phase_keys = {
        "state_advance": "state_advance_time",
        "world_hot_lifecycle": "world_hot_lifecycle_time",
        "player_locator": "player_locator_time",
        "overlay": "overlay_time",
        "session": "session_time",
        "event": "event_time",
        "candidate_finalize": "candidate_finalize_time",
    }
    result: dict[str, dict[str, float]] = {}
    for output_name, phase_key in phase_keys.items():
        summaries = [
            summary
            for receiver in receivers
            for candidate_phases in [receiver.get("candidate_build_phases", {})]
            if isinstance(candidate_phases, dict)
            for lane_phases in [candidate_phases.get("lane_candidate_phases", {})]
            if isinstance(lane_phases, dict)
            for summary in [lane_phases.get(phase_key, {})]
            if isinstance(summary, dict)
        ]
        result[output_name] = {
            "average_us_mean": _round(_mean_field(summaries, "average") / 1_000.0),
            "maximum_ms": _round(_max_number(summaries, "maximum") / 1_000_000.0),
        }
    return result


def _lane_maximums(lanes: list[dict[str, Any]], key: str) -> dict[str, int]:
    result: dict[str, int] = {}
    for lane in lanes:
        name = str(lane.get("lane", "")).strip()
        if not name:
            continue
        result[name] = max(result.get(name, 0), int(float(lane.get(key, 0))))
    return dict(sorted(result.items()))


def _lane_totals(lanes: list[dict[str, Any]], key: str) -> dict[str, int]:
    result: dict[str, int] = {}
    for lane in lanes:
        name = str(lane.get("lane", "")).strip()
        if not name:
            continue
        result[name] = result.get(name, 0) + int(float(lane.get(key, 0)))
    return dict(sorted(result.items()))


def _round(value: float) -> float:
    return round(value, 3)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Summarize receiver-scaling runtime scenario runs."
    )
    parser.add_argument("run_directories", nargs="+", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    result = summarize_matrix(args.run_directories)
    encoded = json.dumps(result, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(encoded, encoding="utf-8")
    print(encoded, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
