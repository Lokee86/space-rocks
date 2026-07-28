from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from statistics import mean
from typing import Any

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.churn_summary import summarize_churn


METRICS = {
    "server_peak_rss_mib": ("server", "peak_rss_mib"),
    "server_resident_set_mib": ("server", "resident_set_mib"),
    "server_heap_allocated_mib": ("server", "heap_allocated_mib"),
    "server_heap_in_use_mib": ("server", "heap_in_use_mib"),
    "server_goroutines": ("server", "goroutines"),
    "server_tick_average_us": ("server", "tick_average_us"),
    "server_outbound_average_us": ("server", "receiver_outbound_average_us"),
    "client_peak_memory_mib": ("client", "peak_memory_mib"),
    "client_frame_average_ms": ("client", "frame_average_ms"),
}


def summarize_soak(run_directory: Path, window_size: int = 10) -> dict[str, Any]:
    result = summarize_churn(run_directory)
    rounds = result.get("rounds", [])
    if not isinstance(rounds, list) or not rounds:
        raise ValueError("soak summary requires at least one completed round")
    window = min(max(window_size, 1), len(rounds))
    result["soak"] = {
        "window_rounds": window,
        "metrics": summarize_round_windows(rounds, window),
        "receiver_skipped_send_ticks_total": sum(
            int(round_result.get("server", {}).get("receiver_skipped_send_ticks", 0))
            for round_result in rounds
        ),
        "client_send_failures_total": sum(
            int(round_result.get("client", {}).get("send_failures", 0))
            for round_result in rounds
        ),
    }
    return result


def summarize_round_windows(
    rounds: list[dict[str, Any]], window_size: int
) -> dict[str, dict[str, float]]:
    window = min(max(window_size, 1), len(rounds))
    summaries: dict[str, dict[str, float]] = {}
    for name, path in METRICS.items():
        values = [_metric(round_result, path) for round_result in rounds]
        head = mean(values[:window])
        tail = mean(values[-window:])
        summaries[name] = {
            "first": _round(values[0]),
            "last": _round(values[-1]),
            "minimum": _round(min(values)),
            "maximum": _round(max(values)),
            "head_average": _round(head),
            "tail_average": _round(tail),
            "tail_minus_head": _round(tail - head),
        }
    return summaries


def _metric(round_result: dict[str, Any], path: tuple[str, str]) -> float:
    section = round_result.get(path[0], {})
    if not isinstance(section, dict):
        return 0.0
    return float(section.get(path[1], 0.0))


def _round(value: float) -> float:
    return round(value, 3)


def main() -> int:
    parser = argparse.ArgumentParser(description="Summarize long match-churn soak drift.")
    parser.add_argument("run_directory", type=Path)
    parser.add_argument("--window-rounds", type=int, default=10)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--quiet", action="store_true")
    args = parser.parse_args()
    result = summarize_soak(args.run_directory, args.window_rounds)
    encoded = json.dumps(result, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(encoded, encoding="utf-8")
    if not args.quiet:
        print(encoded, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
