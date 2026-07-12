"""GDScript constants generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any

from data_sync.generators.go_constants import ConstantsGenerationError


def generate_constants(section_name: str, values: Iterable[tuple[str, Any]]) -> str:
    lines = [f"const {_to_upper_snake_case(name)} := {_format_gds_value(value)}" for name, value in values]
    return "\n".join(lines)


def _to_upper_snake_case(name: str) -> str:
    _validate_snake_case(name)
    return name.upper()


def _format_gds_value(value: Any) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, int):
        return str(value)
    if isinstance(value, float):
        return repr(value)
    if isinstance(value, str):
        return _quote_string(value)
    if _is_vector2(value):
        return f"Vector2({repr(float(value[0]))}, {repr(float(value[1]))})"
    if isinstance(value, list):
        return "[" + ", ".join(_format_gds_value(item) for item in value) + "]"
    if isinstance(value, dict):
        if any(not isinstance(key, str) for key in value):
            invalid_key = next(key for key in value if not isinstance(key, str))
            raise ConstantsGenerationError(
                f"unsupported GDScript dictionary key type: {type(invalid_key).__name__}"
            )
        entries = [
            f"{_quote_string(key)}: {_format_gds_value(nested_value)}"
            for key, nested_value in value.items()
        ]
        return "{" + ", ".join(entries) + "}"
    raise ConstantsGenerationError(f"unsupported constant value type: {type(value).__name__}")


def _quote_string(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def _validate_snake_case(name: str) -> None:
    if not name or name.startswith("_") or name.endswith("_") or "__" in name:
        raise ConstantsGenerationError(f"invalid snake_case constant name: {name!r}")
    if not all(part.isidentifier() and part.islower() for part in name.split("_")):
        raise ConstantsGenerationError(f"invalid snake_case constant name: {name!r}")


def _is_vector2(value: Any) -> bool:
    return (
        isinstance(value, list)
        and len(value) == 2
        and all(isinstance(item, (int, float)) and not isinstance(item, bool) for item in value)
    )
