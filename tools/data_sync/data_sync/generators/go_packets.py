"""Go packet generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any

from data_sync.model.packets import PacketDefinition


GO_TYPES = {
    "bool": "bool",
    "int": "int",
    "uint32": "uint32",
    "float32": "float32",
    "float64": "float64",
    "string": "string",
}


class PacketGenerationError(Exception):
    """Raised when packet code cannot be generated safely."""


def generate_packets(section_name: str, packets: Iterable[PacketDefinition]) -> str:
    chunks: list[str] = []
    for packet in packets:
        packet_name = _to_pascal_case(packet.name)
        chunks.append(f"const Packet{packet_name} = {_format_go_value(packet.id)}")
        chunks.append("")
        chunks.append(f"type {packet_name}Packet struct {{")
        chunks.extend(_go_fields(packet))
        chunks.append("}")
        chunks.append("")
    return "\n".join(chunks).rstrip()


def _go_fields(packet: PacketDefinition) -> list[str]:
    fields = [(_to_pascal_case(field.name), _go_type(field.type)) for field in packet.fields]
    if not fields:
        return []
    width = max(len(name) for name, _field_type in fields)
    return [f"    {name.ljust(width)} {field_type}" for name, field_type in fields]


def _go_type(field_type: str) -> str:
    try:
        return GO_TYPES[field_type]
    except KeyError as exc:
        raise PacketGenerationError(f"unsupported Go packet field type: {field_type}") from exc


def _format_go_value(value: Any) -> str:
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


def _validate_snake_case(name: str) -> None:
    if not name or name.startswith("_") or name.endswith("_") or "__" in name:
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
    if not all(part.isidentifier() and part.islower() for part in name.split("_")):
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
