"""Go constants generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any


class ConstantsGenerationError(Exception):
    """Raised when constants cannot be generated safely."""


def generate_constants(section_name: str, values: Iterable[tuple[str, Any]]) -> str:
    lines = [f"const {_to_pascal_case(name)} = {_format_go_value(value)}" for name, value in values]
    return "\n".join(lines)


def _to_pascal_case(name: str) -> str:
    _validate_snake_case(name)
    return "".join(part.capitalize() for part in name.split("_"))


def _format_go_value(value: Any) -> str:
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
