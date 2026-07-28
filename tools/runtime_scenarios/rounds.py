from __future__ import annotations

from typing import Any


def expand_rounds(rounds: list[Any]) -> list[dict[str, Any]]:
    expanded: list[dict[str, Any]] = []
    for raw_round in rounds:
        if not isinstance(raw_round, dict):
            continue
        repeat_count = int(raw_round.get("repeat", 1))
        base_name = str(raw_round.get("name", "round"))
        for repeat_index in range(repeat_count):
            expanded_round = dict(raw_round)
            expanded_round.pop("repeat", None)
            if repeat_count > 1:
                expanded_round["name"] = f"{base_name}-{repeat_index + 1:03d}"
            expanded.append(expanded_round)
    return expanded
