"""Packet model definitions."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class PacketField:
    name: str
    type: str


@dataclass(frozen=True)
class PacketDefinition:
    name: str
    id: int | str
    direction: str
    fields: tuple[PacketField, ...]

    def field_types(self) -> dict[str, str]:
        return {field.name: field.type for field in self.fields}
