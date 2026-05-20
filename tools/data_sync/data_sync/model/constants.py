"""Constants model definitions."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TypeAlias


ConstantValue: TypeAlias = int | float | bool | str | list[float]


@dataclass(frozen=True)
class ConstantSection:
    name: str
    values: tuple[tuple[str, ConstantValue], ...]

    def as_dict(self) -> dict[str, ConstantValue]:
        return dict(self.values)
