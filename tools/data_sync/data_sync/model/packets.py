"""Packet model definitions."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Mapping


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


@dataclass(frozen=True)
class PacketOutput:
    language: str
    path: str
    package: str | None = None
    imports: Mapping[str, str] | None = None
    packet_types: bool = False
    packet_type_ids: tuple[str, ...] = ()
    structs: tuple[str, ...] = ()
    base: str | None = None
    builders: tuple[str, ...] = ()
    extras: Mapping[str, Any] | None = None
    id: str | None = None


@dataclass(frozen=True)
class PacketSchemaField:
    name: str
    json: str
    type: str
    go_name: str | None = None
    go_type: str | None = None
    key_type: str | None = None
    value_type: str | None = None
    item_type: str | None = None
    go_item_type: str | None = None
    go_value_type: str | None = None
    extras: Mapping[str, Any] | None = None


@dataclass(frozen=True)
class PacketStruct:
    id: str
    fields: tuple[PacketSchemaField, ...]
    extras: Mapping[str, Any] | None = None


@dataclass(frozen=True)
class PacketType:
    id: str
    value: str
    extras: Mapping[str, Any] | None = None


@dataclass(frozen=True)
class PacketBuilder:
    id: str
    args: tuple[str, ...]
    body: Mapping[str, Any]
    extras: Mapping[str, Any] | None = None


@dataclass(frozen=True)
class PacketSchema:
    outputs: tuple[PacketOutput, ...]
    structs: tuple[PacketStruct, ...]
    packet_types: tuple[PacketType, ...]
    builders: tuple[PacketBuilder, ...]

    def output_for_id(self, output_id: str) -> PacketOutput:
        for output in self.outputs:
            if output.id == output_id:
                return output
        raise KeyError(output_id)

    def output_for_path(self, path: str) -> PacketOutput:
        for output in self.outputs:
            if output.path == path:
                return output
        raise KeyError(path)

    def struct(self, struct_id: str) -> PacketStruct:
        for struct in self.structs:
            if struct.id == struct_id:
                return struct
        raise KeyError(struct_id)

    def builder(self, builder_id: str) -> PacketBuilder:
        for builder in self.builders:
            if builder.id == builder_id:
                return builder
        raise KeyError(builder_id)

    def packet_type(self, packet_type_id: str) -> PacketType:
        for packet_type in self.packet_types:
            if packet_type.id == packet_type_id:
                return packet_type
        raise KeyError(packet_type_id)
