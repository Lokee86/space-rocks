from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from runtime_scenarios.model import Scenario


@dataclass(frozen=True)
class MatrixEntry:
    room_count: int
    scenario: Scenario


@dataclass(frozen=True)
class MatrixManifest:
    path: Path
    matrix_id: str
    entries: tuple[MatrixEntry, ...]
    workload_signature: str

    @classmethod
    def load(cls, path: Path) -> "MatrixManifest":
        resolved = path.expanduser().resolve()
        payload = _read_json(resolved)
        matrix_id = str(payload.get("id", "")).strip()
        if not matrix_id:
            raise ValueError("matrix id is required")
        raw_entries = payload.get("scenarios", [])
        if not isinstance(raw_entries, list) or not raw_entries:
            raise ValueError("matrix scenarios must be a non-empty array")

        entries: list[MatrixEntry] = []
        for index, raw_entry in enumerate(raw_entries, start=1):
            if not isinstance(raw_entry, dict):
                raise ValueError(f"matrix scenario {index} must be an object")
            room_count = raw_entry.get("room_count")
            if isinstance(room_count, bool) or not isinstance(room_count, int) or room_count <= 0:
                raise ValueError(f"matrix scenario {index} room_count must be positive")
            relative_path = str(raw_entry.get("path", "")).strip()
            if not relative_path:
                raise ValueError(f"matrix scenario {index} path is required")
            scenario = Scenario.load(resolved.parent / relative_path)
            if scenario.room_count != room_count:
                raise ValueError(
                    f"matrix scenario {scenario.scenario_id} declares {scenario.room_count} rooms, "
                    f"expected {room_count}"
                )
            entries.append(MatrixEntry(room_count=room_count, scenario=scenario))

        room_counts = [entry.room_count for entry in entries]
        if room_counts != [1, 2, 3, 4]:
            raise ValueError("multi-room matrix must contain ordered room counts 1, 2, 3, and 4")
        signatures = {_workload_signature(entry.scenario.raw) for entry in entries}
        if len(signatures) != 1:
            raise ValueError("multi-room matrix scenarios do not share the same workload")
        return cls(
            path=resolved,
            matrix_id=matrix_id,
            entries=tuple(entries),
            workload_signature=signatures.pop(),
        )


def _workload_signature(payload: dict[str, Any]) -> str:
    normalized = {
        key: value
        for key, value in payload.items()
        if key not in {"id", "room_count"}
    }
    encoded = json.dumps(normalized, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def _read_json(path: Path) -> dict[str, Any]:
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except OSError as exc:
        raise ValueError(f"cannot read matrix manifest {path}: {exc}") from exc
    except json.JSONDecodeError as exc:
        raise ValueError(f"invalid matrix manifest JSON {path}: {exc}") from exc
    if not isinstance(payload, dict):
        raise ValueError("matrix manifest root must be an object")
    return payload
