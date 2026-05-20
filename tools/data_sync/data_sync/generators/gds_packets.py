"""GDScript packet generator."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any

from data_sync.generators.go_packets import PacketGenerationError
from data_sync.model.packets import PacketDefinition


def generate_packets(section_name: str, packets: Iterable[PacketDefinition]) -> str:
    lines: list[str] = []
    for packet in packets:
        packet_name = _to_upper_snake_case(packet.name)
        field_names = ", ".join(_quote_string(field.name) for field in packet.fields)
        lines.append(f"const PACKET_{packet_name} := {_format_gds_value(packet.id)}")
        lines.append(f"const PACKET_{packet_name}_FIELDS := [{field_names}]")
    return "\n".join(lines)


def _format_gds_value(value: Any) -> str:
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


def _to_upper_snake_case(name: str) -> str:
    _validate_snake_case(name)
    return name.upper()


def _validate_snake_case(name: str) -> None:
    if not name or name.startswith("_") or name.endswith("_") or "__" in name:
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
    if not all(part.isidentifier() and part.islower() for part in name.split("_")):
        raise PacketGenerationError(f"invalid snake_case name: {name!r}")
