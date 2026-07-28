from __future__ import annotations

from typing import Any

from runtime_scenarios.rounds import expand_rounds


def phase_markers_for_scenario(payload: dict[str, Any]) -> list[dict[str, Any]]:
    rounds = payload.get("rounds")
    if isinstance(rounds, list) and rounds:
        return _round_phase_markers(expand_rounds(rounds))
    phases = payload.get("phases", [])
    return _phase_markers(phases if isinstance(phases, list) else [])


def _round_phase_markers(rounds: list[dict[str, Any]]) -> list[dict[str, Any]]:
    elapsed = 0.0
    markers: list[dict[str, Any]] = []
    for round_index, round_payload in enumerate(rounds, start=1):
        phases = round_payload.get("phases", [])
        if not isinstance(phases, list):
            continue
        for marker in _phase_markers(phases, elapsed):
            marker["name"] = f"{round_payload.get('name', f'round-{round_index}')}/{marker['name']}"
            marker["round"] = round_index
            markers.append(marker)
        elapsed = markers[-1]["end_seconds"] if markers else elapsed
    return markers


def _phase_markers(phases: list[Any], start: float = 0.0) -> list[dict[str, Any]]:
    elapsed = start
    markers: list[dict[str, Any]] = []
    for phase in phases:
        if not isinstance(phase, dict):
            continue
        duration = float(phase.get("duration_seconds", 0.0))
        markers.append(
            {
                "name": str(phase.get("name", "")),
                "start_seconds": elapsed,
                "end_seconds": elapsed + duration,
                "duration_seconds": duration,
            }
        )
        elapsed += duration
    return markers
