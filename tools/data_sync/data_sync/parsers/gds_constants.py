"""Strict GDScript constants parser for pull."""

from __future__ import annotations

import re
from typing import Any

from data_sync.parsers.go_constants import ConstantsParseError, _parse_value


LINE_RE = re.compile(r"^const (?P<name>[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)*) := (?P<value>.+)$")


def parse_constants(text: str) -> tuple[tuple[str, Any], ...]:
    values: list[tuple[str, Any]] = []
    for line_number, line in _canonical_lines(text):
        match = LINE_RE.fullmatch(line)
        if match is None:
            raise ConstantsParseError(f"line {line_number}: expected canonical GDScript const assignment")
        values.append((match.group("name").lower(), _parse_value(match.group("value"))))
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
