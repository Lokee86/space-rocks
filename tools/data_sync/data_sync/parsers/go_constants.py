"""Strict Go constants parser for pull."""

from __future__ import annotations

import re
from typing import Any


LINE_RE = re.compile(r"^const (?P<name>[A-Z][A-Za-z0-9]*) = (?P<value>.+)$")


class ConstantsParseError(Exception):
    """Raised when a managed constants block is not canonical parseable syntax."""


def parse_constants(text: str) -> tuple[tuple[str, Any], ...]:
    values: list[tuple[str, Any]] = []
    for line_number, line in _canonical_lines(text):
        match = LINE_RE.fullmatch(line)
        if match is None:
            raise ConstantsParseError(f"line {line_number}: expected canonical Go const assignment")
        values.append((_pascal_to_snake(match.group("name")), _parse_value(match.group("value"))))
    return tuple(values)


def _canonical_lines(text: str) -> tuple[tuple[int, str], ...]:
    stripped = text.rstrip("\n")
    if not stripped:
        return ()
    lines = stripped.split("\n")
    result: list[tuple[int, str]] = []
    for index, line in enumerate(lines, start=1):
        if not line or line != line.strip():
            raise ConstantsParseError(f"line {index}: non-canonical whitespace")
        result.append((index, line))
    return tuple(result)


def _parse_value(value: str) -> Any:
    if value == "true":
        return True
    if value == "false":
        return False
    if value.startswith('"') and value.endswith('"'):
        return _parse_string(value)
    vector_match = re.fullmatch(
        r"Vector2\((-?(?:0|[1-9][0-9]*)\.[0-9]+), (-?(?:0|[1-9][0-9]*)\.[0-9]+)\)",
        value,
    )
    if vector_match is not None:
        return [float(vector_match.group(1)), float(vector_match.group(2))]
    if re.fullmatch(r"-?(0|[1-9][0-9]*)", value):
        return int(value)
    if re.fullmatch(r"-?(0|[1-9][0-9]*)\.[0-9]+", value):
        return float(value)
    raise ConstantsParseError(f"unsupported constant value syntax: {value}")


def _parse_string(value: str) -> str:
    content = value[1:-1]
    result = ""
    index = 0
    while index < len(content):
        char = content[index]
        if char != "\\":
            result += char
            index += 1
            continue
        index += 1
        if index >= len(content) or content[index] not in {'"', "\\"}:
            raise ConstantsParseError(f"unsupported string escape in value: {value}")
        result += content[index]
        index += 1
    return result


def _pascal_to_snake(name: str) -> str:
    chars: list[str] = []
    for index, char in enumerate(name):
        if char.isupper() and index > 0:
            chars.append("_")
        chars.append(char.lower())
    return "".join(chars)
