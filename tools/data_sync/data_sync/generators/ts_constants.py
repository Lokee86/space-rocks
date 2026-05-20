"""TypeScript constants generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any

from data_sync.generators.go_constants import ConstantsGenerationError


def generate_constants(section_name: str, values: Iterable[tuple[str, Any]]) -> str:
    lines = [
        f"export const {_to_upper_snake_case(name)} = {_format_ts_value(value)};"
        for name, value in values
    ]
    return "\n".join(lines)


def _to_upper_snake_case(name: str) -> str:
    _validate_snake_case(name)
    return name.upper()


def _format_ts_value(value: Any) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, int):
        return str(value)
    if isinstance(value, float):
        return repr(value)
    if isinstance(value, str):
        return _quote_string(value)
    raise ConstantsGenerationError(f"unsupported constant value type: {type(value).__name__}")


def _quote_string(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def _validate_snake_case(name: str) -> None:
    if not name or name.startswith("_") or name.endswith("_") or "__" in name:
        raise ConstantsGenerationError(f"invalid snake_case constant name: {name!r}")
    if not all(part.isidentifier() and part.islower() for part in name.split("_")):
        raise ConstantsGenerationError(f"invalid snake_case constant name: {name!r}")
