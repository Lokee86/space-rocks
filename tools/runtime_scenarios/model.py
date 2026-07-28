from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


class ScenarioError(ValueError):
    """Raised when a runtime scenario definition is invalid."""


@dataclass(frozen=True)
class ClientPlan:
    visible: int
    headless: int

    @property
    def total(self) -> int:
        return self.visible + self.headless


@dataclass(frozen=True)
class Scenario:
    path: Path
    scenario_id: str
    seed: int
    timeout_seconds: float
    setup_timeout_seconds: float
    bots: int
    clients: ClientPlan
    raw: dict[str, Any]

    @classmethod
    def load(cls, path: Path) -> "Scenario":
        resolved = path.expanduser().resolve()
        try:
            payload = json.loads(resolved.read_text(encoding="utf-8"))
        except OSError as exc:
            raise ScenarioError(f"cannot read scenario {resolved}: {exc}") from exc
        except json.JSONDecodeError as exc:
            raise ScenarioError(f"invalid scenario JSON {resolved}: {exc}") from exc
        if not isinstance(payload, dict):
            raise ScenarioError("scenario root must be an object")

        scenario_id = _required_text(payload, "id")
        seed = _required_int(payload, "seed")
        timeout_seconds = _positive_number(payload, "timeout_seconds", 180.0)
        setup_timeout_seconds = _positive_number(payload, "setup_timeout_seconds", 45.0)
        bots = _nonnegative_int(payload, "bots", 0)

        clients_payload = payload.get("clients", {})
        if not isinstance(clients_payload, dict):
            raise ScenarioError("clients must be an object")
        visible = _nonnegative_int(clients_payload, "visible", 1)
        headless = _nonnegative_int(clients_payload, "headless", 0)
        if visible != 1:
            raise ScenarioError("the first harness version requires exactly one visible coordinator")
        clients = ClientPlan(visible=visible, headless=headless)
        if clients.total < 1:
            raise ScenarioError("at least one client is required")

        setup = payload.get("setup", {})
        if not isinstance(setup, dict):
            raise ScenarioError("setup must be an object")
        _nonnegative_int(setup, "asteroid_spawns", 0)
        _positive_number(setup, "settle_seconds", 0.0, allow_zero=True)

        heap_profile_rounds = payload.get("heap_profile_rounds", [])
        if not isinstance(heap_profile_rounds, list):
            raise ScenarioError("heap_profile_rounds must be an array")
        for index, round_number in enumerate(heap_profile_rounds, start=1):
            valid_round = isinstance(round_number, int) and not isinstance(round_number, bool)
            if not valid_round or round_number <= 0:
                raise ScenarioError(
                    f"heap_profile_rounds entry {index} must be a positive integer"
                )

        phases = payload.get("phases")
        rounds = payload.get("rounds")
        if phases is not None and rounds is not None:
            raise ScenarioError("scenario must define phases or rounds, not both")
        if rounds is not None:
            if not isinstance(rounds, list) or not rounds:
                raise ScenarioError("rounds must be a non-empty array")
            for round_index, round_payload in enumerate(rounds):
                if not isinstance(round_payload, dict):
                    raise ScenarioError(f"round {round_index} must be an object")
                _required_text(round_payload, "name")
                _positive_int(round_payload, "repeat", 1)
                _positive_int(round_payload, "lives", 2)
                round_setup = round_payload.get("setup", {})
                if not isinstance(round_setup, dict):
                    raise ScenarioError(f"round {round_index} setup must be an object")
                _nonnegative_int(round_setup, "asteroid_spawns", 0)
                _positive_number(round_setup, "settle_seconds", 0.0, allow_zero=True)
                _validate_phases(round_payload.get("phases"), f"round {round_index}")
        else:
            _validate_phases(phases, "scenario")

        return cls(
            path=resolved,
            scenario_id=scenario_id,
            seed=seed,
            timeout_seconds=timeout_seconds,
            setup_timeout_seconds=setup_timeout_seconds,
            bots=bots,
            clients=clients,
            raw=payload,
        )


def _required_text(payload: dict[str, Any], key: str) -> str:
    value = str(payload.get(key, "")).strip()
    if not value:
        raise ScenarioError(f"{key} is required")
    return value


def _required_int(payload: dict[str, Any], key: str) -> int:
    value = payload.get(key)
    if isinstance(value, bool) or not isinstance(value, int):
        raise ScenarioError(f"{key} must be an integer")
    return value


def _validate_phases(value: object, owner: str) -> None:
    if not isinstance(value, list) or not value:
        raise ScenarioError(f"{owner} phases must be a non-empty array")
    for index, phase in enumerate(value):
        if not isinstance(phase, dict):
            raise ScenarioError(f"{owner} phase {index} must be an object")
        _required_text(phase, "name")
        _positive_number(phase, "duration_seconds", 0.0, allow_zero=True)
        _nonnegative_int(phase, "bullet_streams", 0)


def _positive_int(payload: dict[str, Any], key: str, default: int) -> int:
    value = payload.get(key, default)
    if isinstance(value, bool) or not isinstance(value, int) or value <= 0:
        raise ScenarioError(f"{key} must be a positive integer")
    return value


def _nonnegative_int(payload: dict[str, Any], key: str, default: int) -> int:
    value = payload.get(key, default)
    if isinstance(value, bool) or not isinstance(value, int) or value < 0:
        raise ScenarioError(f"{key} must be a non-negative integer")
    return value


def _positive_number(
    payload: dict[str, Any],
    key: str,
    default: float,
    *,
    allow_zero: bool = False,
) -> float:
    value = payload.get(key, default)
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ScenarioError(f"{key} must be a number")
    number = float(value)
    minimum = 0.0 if allow_zero else 0.000001
    if number < minimum:
        qualifier = "non-negative" if allow_zero else "positive"
        raise ScenarioError(f"{key} must be {qualifier}")
    return number
