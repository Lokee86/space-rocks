#!/usr/bin/env python3
"""One-time JSON to TOML migration for the data_sync source of truth."""

from __future__ import annotations

import argparse
import json
import sys
from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any


CONSTANT_TYPES = {"int", "float", "bool", "string", "vector2"}
CONSTANT_SECTION_MAP = {
    "server_tick_rate": "constants.server.runtime",
    "player_rotation_speed": "constants.server.player_movement",
    "player_thrust_force": "constants.server.player_movement",
    "player_max_speed": "constants.server.player_movement",
    "player_damping": "constants.server.player_movement",
    "player_starting_lives": "constants.shared.player_state",
    "player_respawn_delay": "constants.shared.player_state",
    "player_respawn_buffer": "constants.server.player_session",
    "game_over_sound_delay": "constants.client.presentation",
    "player_resume_invulnerability_seconds": "constants.server.player_session",
    "background_parallax": "constants.client.presentation",
    "foreground_background_parallax": "constants.client.presentation",
    "foreground_background_offset": "constants.client.presentation",
    "player_interpolation_speed": "constants.client.presentation",
    "world_size": "constants.server.world",
    "window_min_size": "constants.client.presentation",
    "window_max_size": "constants.client.presentation",
    "base_score": "constants.server.scoring",
    "asteroid_spawn_interval": "constants.server.asteroids",
    "asteroid_spawn_batch_size": "constants.server.asteroids",
    "asteroid_spawn_margin": "constants.server.asteroids",
    "asteroid_despawn_margin": "constants.server.asteroids",
    "asteroid_min_speed": "constants.server.asteroids",
    "asteroid_max_speed": "constants.server.asteroids",
    "asteroid_size_scale": "constants.shared.asteroid_visual",
    "asteroid_aim_randomness_degrees": "constants.server.asteroids",
    "bullet_speed": "constants.server.bullets",
    "bullet_lifetime": "constants.server.bullets",
    "bullet_cooldown": "constants.server.bullets",
    "bullet_spawn_offset": "constants.server.bullets",
    "collision_despawn_delay": "constants.server.bullets",
}
FIELD_TYPE_MAP = {
    "bool": "bool",
    "int": "int",
    "float": "float64",
    "string": "string",
}
CLIENT_TO_SERVER_TYPES = {
    "input",
    "client_config",
    "respawn",
    "pause_player",
    "resume_player",
    "toggle_debug_invincible",
    "toggle_debug_infinite_lives",
    "toggle_debug_freeze_world",
}
SERVER_TO_CLIENT_TYPES = {
    "state",
    "bullet_blast",
    "ship_death",
}


class MigrationError(Exception):
    """Raised when the disposable migration cannot safely convert input data."""


def main() -> None:
    raise SystemExit(run())


def run(argv: Sequence[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    try:
        import tomlkit
    except ModuleNotFoundError:
        print(
            "migration error: tomlkit is required for TOML writing. Install it, then rerun this script.",
            file=sys.stderr,
        )
        return 2

    try:
        constants_json = load_json(args.constants_input)
        packets_json = load_json(args.packets_input)
        constants_summary = convert_constants(tomlkit, constants_json)
        packets_summary = convert_packets(tomlkit, packets_json)

        document = tomlkit.document()
        for section_name, table in constants_summary.sections.items():
            add_nested_table(tomlkit, document, section_name, table)
        for packet_name, table in packets_summary.packets.items():
            add_nested_table(tomlkit, document, f"packets.{packet_name}", table)

        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(tomlkit.dumps(document), encoding="utf-8")
    except (OSError, json.JSONDecodeError, MigrationError) as exc:
        print(f"migration error: {exc}", file=sys.stderr)
        return 1

    print("JSON to TOML migration complete")
    print(f"constants sections converted: {', '.join(constants_summary.sections)}")
    print(f"constants count: {constants_summary.constant_count}")
    print(f"packets converted: {packets_summary.packet_count}")
    print("packet field counts:")
    for packet_name, count in packets_summary.field_counts.items():
        print(f"  {packet_name}: {count}")
    print(f"output: {args.output}")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Disposable migration from legacy JSON data sources to shared/game_data.toml.",
    )
    parser.add_argument("--constants-input", required=True, type=Path)
    parser.add_argument("--packets-input", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    return parser


def add_nested_table(tomlkit: Any, document: Any, section_name: str, table: Any) -> None:
    parts = section_name.split(".")
    current = document
    for part in parts[:-1]:
        if part not in current:
            current.add(part, tomlkit.table())
        current = current[part]
    current.add(parts[-1], table)


def load_json(path: Path) -> Mapping[str, Any]:
    try:
        with path.open("r", encoding="utf-8") as handle:
            data = json.load(handle)
    except FileNotFoundError as exc:
        raise MigrationError(f"input file does not exist: {path}") from exc
    except json.JSONDecodeError:
        raise
    except OSError as exc:
        raise MigrationError(f"failed to read {path}: {exc}") from exc

    if not isinstance(data, Mapping):
        raise MigrationError(f"top-level JSON must be an object: {path}")
    return data


class ConstantsSummary:
    def __init__(self, sections: Mapping[str, Any], constant_count: int) -> None:
        self.sections = sections
        self.constant_count = constant_count


class PacketsSummary:
    def __init__(self, packets: Mapping[str, Any], field_counts: Mapping[str, int]) -> None:
        self.packets = packets
        self.field_counts = field_counts

    @property
    def packet_count(self) -> int:
        return len(self.packets)


def convert_constants(tomlkit: Any, data: Mapping[str, Any]) -> ConstantsSummary:
    groups = require_list(data, "groups", "constants JSON")
    sections: dict[str, Any] = {}
    constant_count = 0

    for group in groups:
        if not isinstance(group, Mapping):
            raise MigrationError("constants groups must be objects")
        group_id = require_string(group, "id", "constants group")
        constants = require_list(group, "constants", f"constants group {group_id}")
        table = tomlkit.table()

        for constant in constants:
            if not isinstance(constant, Mapping):
                raise MigrationError(f"constants in group {group_id} must be objects")
            constant_id = require_string(constant, "id", f"constant in group {group_id}")
            constant_type = require_string(constant, "type", f"constant {constant_id}")
            if constant_type not in CONSTANT_TYPES:
                raise MigrationError(f"unsupported constant type for {constant_id}: {constant_type}")
            if "value" not in constant:
                raise MigrationError(f"constant {constant_id} missing value")
            section_name = CONSTANT_SECTION_MAP.get(constant_id)
            if section_name is None:
                raise MigrationError(f"constant {constant_id} has no domain-specific section mapping")
            section_table = sections.setdefault(section_name, tomlkit.table())
            for output_name, output_type, output_value in convert_constant_entries(
                constant_id,
                constant_type,
                constant["value"],
            ):
                section_table.add(output_name, convert_constant_value(output_name, output_type, output_value))
                constant_count += 1

    return ConstantsSummary(sections, constant_count)


def convert_constant_entries(
    constant_id: str,
    constant_type: str,
    value: Any,
) -> tuple[tuple[str, str, Any], ...]:
    if constant_id == "world_size":
        if not is_number_pair(value):
            raise MigrationError("world_size must be a two-number vector2")
        return (
            ("world_width", "float", value[0]),
            ("world_height", "float", value[1]),
        )
    return ((constant_id, constant_type, value),)


def convert_constant_value(constant_id: str, constant_type: str, value: Any) -> Any:
    if constant_type == "int" and isinstance(value, int) and not isinstance(value, bool):
        return value
    if constant_type == "float" and isinstance(value, (int, float)) and not isinstance(value, bool):
        return float(value)
    if constant_type == "bool" and isinstance(value, bool):
        return value
    if constant_type == "string" and isinstance(value, str):
        return value
    if constant_type == "vector2" and is_number_pair(value):
        return [float(value[0]), float(value[1])]
    raise MigrationError(f"unsupported value for {constant_id} ({constant_type}): {value!r}")


def convert_packets(tomlkit: Any, data: Mapping[str, Any]) -> PacketsSummary:
    packet_types = require_list(data, "packet_types", "packets JSON")
    structs = structs_by_id(require_list(data, "structs", "packets JSON"))
    builders = require_list(data, "builders", "packets JSON")
    builder_types = packet_types_from_builders(builders)

    packets: dict[str, Any] = {}
    field_counts: dict[str, int] = {}

    for packet_type in packet_types:
        if not isinstance(packet_type, Mapping):
            raise MigrationError("packet_types entries must be objects")
        packet_id = require_string(packet_type, "id", "packet type")
        packet_value = require_string(packet_type, "value", f"packet type {packet_id}")

        table = tomlkit.table()
        table.add("id", packet_value)
        table.add("direction", infer_direction(packet_value, builder_types))

        fields = fields_for_packet(packet_value, structs)
        if fields:
            fields_table = tomlkit.table()
            for field_name, field_type in fields.items():
                fields_table.add(field_name, field_type)
            table.add("fields", fields_table)

        packets[packet_id] = table
        field_counts[packet_id] = len(fields)

    return PacketsSummary(packets, field_counts)


def structs_by_id(structs: Sequence[Any]) -> dict[str, Mapping[str, Any]]:
    result: dict[str, Mapping[str, Any]] = {}
    for struct in structs:
        if not isinstance(struct, Mapping):
            raise MigrationError("structs entries must be objects")
        struct_id = require_string(struct, "id", "struct")
        result[struct_id] = struct
    return result


def packet_types_from_builders(builders: Sequence[Any]) -> set[str]:
    result: set[str] = set()
    for builder in builders:
        if not isinstance(builder, Mapping):
            raise MigrationError("builders entries must be objects")
        body = builder.get("body")
        if not isinstance(body, Mapping):
            raise MigrationError(f"builder {builder.get('id', '<unknown>')} missing body object")
        packet_type = body.get("type")
        if isinstance(packet_type, str):
            result.add(packet_type)
    return result


def infer_direction(packet_value: str, builder_types: set[str]) -> str:
    if packet_value in builder_types or packet_value in CLIENT_TO_SERVER_TYPES:
        return "client_to_server"
    if packet_value in SERVER_TO_CLIENT_TYPES:
        return "server_to_client"
    return "bidirectional"


def fields_for_packet(packet_value: str, structs: Mapping[str, Mapping[str, Any]]) -> dict[str, str]:
    if packet_value == "input":
        return struct_fields(structs, "InputState")
    if packet_value == "client_config":
        return struct_fields(structs, "ClientConfig")
    if packet_value == "state":
        return struct_fields(structs, "StatePacket", skip={"type"})
    if packet_value in {"bullet_blast", "ship_death"}:
        return struct_fields(structs, "EventState", skip={"type"})
    return {}


def struct_fields(
    structs: Mapping[str, Mapping[str, Any]],
    struct_id: str,
    skip: set[str] | None = None,
) -> dict[str, str]:
    struct = structs.get(struct_id)
    if struct is None:
        raise MigrationError(f"packet conversion needs missing struct: {struct_id}")
    fields = require_list(struct, "fields", f"struct {struct_id}")
    result: dict[str, str] = {}
    skip = skip or set()

    for field in fields:
        if not isinstance(field, Mapping):
            raise MigrationError(f"fields in struct {struct_id} must be objects")
        field_name = require_string(field, "json", f"field in struct {struct_id}")
        if field_name in skip:
            continue
        result[field_name] = convert_field_type(field, struct_id, field_name)

    return result


def convert_field_type(field: Mapping[str, Any], struct_id: str, field_name: str) -> str:
    field_type = require_string(field, "type", f"field {struct_id}.{field_name}")
    if field_type in FIELD_TYPE_MAP:
        return FIELD_TYPE_MAP[field_type]
    if field_type == "array":
        item_type = require_string(field, "item_type", f"field {struct_id}.{field_name}")
        return f"array<{item_type}>"
    if field_type == "map":
        key_type = require_string(field, "key_type", f"field {struct_id}.{field_name}")
        value_type = require_string(field, "value_type", f"field {struct_id}.{field_name}")
        return f"map<{key_type},{value_type}>"
    return field_type


def require_list(data: Mapping[str, Any], key: str, context: str) -> list[Any]:
    value = data.get(key)
    if not isinstance(value, list):
        raise MigrationError(f"{context} missing list: {key}")
    return value


def require_string(data: Mapping[str, Any], key: str, context: str) -> str:
    value = data.get(key)
    if not isinstance(value, str) or not value:
        raise MigrationError(f"{context} missing non-empty string: {key}")
    return value


def is_number_pair(value: Any) -> bool:
    return (
        isinstance(value, list)
        and len(value) == 2
        and all(isinstance(item, (int, float)) and not isinstance(item, bool) for item in value)
    )


if __name__ == "__main__":
    main()
