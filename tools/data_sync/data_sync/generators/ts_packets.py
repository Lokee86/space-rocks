"""TypeScript packet generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any

from data_sync.generators.go_packets import PacketGenerationError
from data_sync.model.packets import PacketDefinition


TS_TYPES = {
    "bool": "boolean",
    "int": "number",
    "uint32": "number",
    "float32": "number",
    "float64": "number",
    "string": "string",
}


def generate_packets(section_name: str, packets: Iterable[PacketDefinition]) -> str:
    chunks: list[str] = []
    for packet in packets:
        packet_name = _to_pascal_case(packet.name)
        chunks.append(f"export const PACKET_{_to_upper_snake_case(packet.name)} = {_format_ts_value(packet.id)};")
        chunks.append("")
        chunks.append(f"export interface {packet_name}Packet {{")
        for field in packet.fields:
            chunks.append(f"  {field.name}: {_ts_type(field.type)};")
        chunks.append("}")
        chunks.append("")
    return "\n".join(chunks).rstrip()


def _ts_type(field_type: str) -> str:
    try:
        return TS_TYPES[field_type]
    except KeyError as exc:
        raise PacketGenerationError(f"unsupported TypeScript packet field type: {field_type}") from exc


def _format_ts_value(value: Any) -> str:
    if isinstance(value, bool):
        raise PacketGenerationError("packet IDs cannot be bool")
    if isinstance(value, int):
        return str(value)
    if isinstance(value, str):
        return _quote_string(value)
    raise PacketGenerationError(f"unsupported packet ID type: {type(value).__name__}")


def _quote_string(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def _to_pascal_case(name: str) -> str:
    _validate_snake_case(name)
    return "".join(part.capitalize() for part in name.split("_"))


def _to_upper_snake_case(name: str) -> str:
    _validate_snake_case(name)
    return name.upper()


def _validate_snake_case(name: str) -> None:
    if not name or name.startswith("_") or name.endswith("_") or "__" in name:
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
    if not all(part.isidentifier() and part.islower() for part in name.split("_")):
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
