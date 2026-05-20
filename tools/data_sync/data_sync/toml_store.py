"""TOML source-of-truth access."""

from __future__ import annotations

import os
import tempfile
from pathlib import Path
from typing import Any, Mapping

from data_sync.model.constants import ConstantSection, ConstantValue
from data_sync.model.packets import PacketDefinition, PacketField


class TomlStoreError(Exception):
    """Raised when the source-of-truth TOML cannot be read or updated."""


class TomlStore:
    def __init__(self, path: Path, document: Any, tomlkit_module: Any) -> None:
        self.path = path
        self.document = document
        self._tomlkit = tomlkit_module

    @classmethod
    def load(cls, path: Path | str) -> "TomlStore":
        tomlkit = _load_tomlkit()
        resolved_path = Path(path)
        try:
            text = resolved_path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise TomlStoreError(f"SoT TOML file does not exist: {resolved_path}") from exc
        except OSError as exc:
            raise TomlStoreError(f"failed to read SoT TOML {resolved_path}: {exc}") from exc

        try:
            document = tomlkit.parse(text)
        except Exception as exc:
            raise TomlStoreError(f"failed to parse SoT TOML {resolved_path}: {exc}") from exc

        return cls(resolved_path, document, tomlkit)

    def constants(self, section_name: str) -> ConstantSection:
        table = self._get_table(section_name)
        values: list[tuple[str, ConstantValue]] = []
        for key, value in table.items():
            values.append((key, _unwrap_value(value)))
        return ConstantSection(section_name, tuple(values))

    def update_constants(self, section_name: str, values: Mapping[str, ConstantValue]) -> None:
        table = self._tomlkit.table()
        for key, value in values.items():
            table.add(key, value)
        self._set_table(section_name, table)

    def packets(self) -> tuple[PacketDefinition, ...]:
        packets_table = self.document.get("packets")
        if packets_table is None:
            return ()
        if not _is_mapping(packets_table):
            raise TomlStoreError("[packets] must be a table")

        packets: list[PacketDefinition] = []
        for packet_name in packets_table:
            packet_table = packets_table[packet_name]
            if not _is_mapping(packet_table):
                raise TomlStoreError(f"[packets.{packet_name}] must be a table")
            packets.append(self._packet_from_table(packet_name, packet_table))
        return tuple(packets)

    def packet(self, packet_name: str) -> PacketDefinition:
        table = self._get_table(f"packets.{packet_name}")
        return self._packet_from_table(packet_name, table)

    def update_packet(self, packet: PacketDefinition) -> None:
        table = self._tomlkit.table()
        table.add("id", packet.id)
        table.add("direction", packet.direction)

        fields_table = self._tomlkit.table()
        for field in packet.fields:
            fields_table.add(field.name, field.type)
        table.add("fields", fields_table)

        self._set_table(f"packets.{packet.name}", table)

    def update_packets(self, packets: tuple[PacketDefinition, ...] | list[PacketDefinition]) -> None:
        for packet in packets:
            self.update_packet(packet)

    def write(self, path: Path | str | None = None) -> None:
        output_path = Path(path) if path is not None else self.path
        output_path.parent.mkdir(parents=True, exist_ok=True)
        text = self._tomlkit.dumps(self.document)

        temp_name = ""
        try:
            with tempfile.NamedTemporaryFile(
                "w",
                encoding="utf-8",
                dir=output_path.parent,
                delete=False,
            ) as handle:
                temp_name = handle.name
                handle.write(text)
                handle.flush()
                os.fsync(handle.fileno())
            os.replace(temp_name, output_path)
        except OSError as exc:
            if temp_name:
                try:
                    os.unlink(temp_name)
                except OSError:
                    pass
            raise TomlStoreError(f"failed to write SoT TOML {output_path}: {exc}") from exc

    def _packet_from_table(self, packet_name: str, table: Mapping[str, Any]) -> PacketDefinition:
        if "id" not in table:
            raise TomlStoreError(f"[packets.{packet_name}] missing id")
        if "direction" not in table:
            raise TomlStoreError(f"[packets.{packet_name}] missing direction")

        packet_id = _unwrap_value(table["id"])
        direction = _unwrap_value(table["direction"])
        if not isinstance(packet_id, (int, str)) or isinstance(packet_id, bool):
            raise TomlStoreError(f"[packets.{packet_name}].id must be an int or string")
        if not isinstance(direction, str):
            raise TomlStoreError(f"[packets.{packet_name}].direction must be a string")

        fields_table = table.get("fields", {})
        if not _is_mapping(fields_table):
            raise TomlStoreError(f"[packets.{packet_name}.fields] must be a table")

        fields: list[PacketField] = []
        for field_name, field_type in fields_table.items():
            field_type = _unwrap_value(field_type)
            if not isinstance(field_type, str):
                raise TomlStoreError(f"[packets.{packet_name}.fields].{field_name} must be a string")
            fields.append(PacketField(field_name, field_type))

        return PacketDefinition(packet_name, packet_id, direction, tuple(fields))

    def _get_table(self, section_name: str) -> Mapping[str, Any]:
        current = self.document
        for part in _split_section_name(section_name):
            if not _is_mapping(current) or part not in current:
                raise TomlStoreError(f"missing TOML section [{section_name}]")
            current = current[part]
        if not _is_mapping(current):
            raise TomlStoreError(f"[{section_name}] must be a table")
        return current

    def _set_table(self, section_name: str, table: Any) -> None:
        parts = _split_section_name(section_name)
        current = self.document
        for part in parts[:-1]:
            if part not in current:
                current.add(part, self._tomlkit.table())
            current = current[part]
            if not _is_mapping(current):
                raise TomlStoreError(f"[{'.'.join(parts[:-1])}] must be a table")
        current[parts[-1]] = table


def _load_tomlkit() -> Any:
    try:
        import tomlkit

        return tomlkit
    except ModuleNotFoundError as exc:
        raise TomlStoreError("tomlkit is required for reading and writing the SoT TOML") from exc


def _split_section_name(section_name: str) -> list[str]:
    parts = section_name.split(".")
    if not parts or any(not part for part in parts):
        raise TomlStoreError(f"invalid TOML section name: {section_name!r}")
    return parts


def _is_mapping(value: Any) -> bool:
    return hasattr(value, "items") and hasattr(value, "__contains__")


def _unwrap_value(value: Any) -> Any:
    if hasattr(value, "unwrap"):
        return value.unwrap()
    return value
