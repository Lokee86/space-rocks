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

        phases = payload.get("phases")
        if not isinstance(phases, list) or not phases:
            raise ScenarioError("phases must be a non-empty array")
        for index, phase in enumerate(phases):
            if not isinstance(phase, dict):
                raise ScenarioError(f"phase {index} must be an object")
            _required_text(phase, "name")
            _positive_number(phase, "duration_seconds", 0.0, allow_zero=True)

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
