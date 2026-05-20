"""Reusable packet rendering helpers."""

from __future__ import annotations

import json
import re

from data_sync.model.packets import PacketSchemaField


class PacketRenderingError(Exception):
    """Raised when a packet type or value cannot be rendered."""


GO_PRIMITIVES = {
    "string": "string",
    "float": "float64",
    "int": "int",
    "bool": "bool",
    "uint32": "uint32",
    "float32": "float32",
    "float64": "float64",
}
GDSCRIPT_PACKET_TYPE_CONSTANTS = {
    "input": "TYPE_INPUT",
    "client_config": "TYPE_CLIENT_CONFIG",
    "state": "TYPE_STATE",
    "bullet_blast": "TYPE_BULLET_BLAST",
    "ship_death": "TYPE_SHIP_DEATH",
    "respawn": "TYPE_RESPAWN",
}


def go_field_name(field: PacketSchemaField) -> str:
    return field.go_name or snake_to_go_name(field.name)


def go_json_tag(field: PacketSchemaField) -> str:
    return f'json:"{field.json}"'


def go_type_for_field(field: PacketSchemaField) -> str:
    if field.go_type:
        return field.go_type

    field_type = field.type
    if field_type in GO_PRIMITIVES:
        return GO_PRIMITIVES[field_type]
    if field_type in {"map", "dictionary"}:
        if not field.key_type or not field.value_type:
            raise PacketRenderingError(f"{field.name} map field requires key_type and value_type")
        key_type = scalar_go_type(field.key_type)
        value_type = field.go_value_type or field.value_type
        return f"map[{key_type}]{value_type}"
    if field_type in {"array", "list"}:
        if not field.item_type:
            raise PacketRenderingError(f"{field.name} array field requires item_type")
        item_type = field.go_item_type or field.item_type
        return f"[]{item_type}"

    rich_type = parse_rich_type(field_type)
    if rich_type is not None:
        kind, args = rich_type
        if kind in {"map", "dictionary"}:
            if len(args) != 2:
                raise PacketRenderingError(f"{field.name} map type requires key and value types")
            return f"map[{scalar_go_type(args[0])}]{args[1]}"
        if kind in {"array", "list"}:
            if len(args) != 1:
                raise PacketRenderingError(f"{field.name} array type requires one item type")
            return f"[]{args[0]}"

    return field_type


def scalar_go_type(value_type: str) -> str:
    try:
        return GO_PRIMITIVES[value_type]
    except KeyError as exc:
        raise PacketRenderingError(f"unsupported scalar Go type: {value_type}") from exc


def parse_rich_type(value: str) -> tuple[str, tuple[str, ...]] | None:
    match = re.fullmatch(r"([A-Za-z_][A-Za-z0-9_]*)<(.+)>", value.strip())
    if match is None:
        return None
    kind = match.group(1)
    args = tuple(part.strip() for part in split_type_args(match.group(2)))
    if not all(args):
        raise PacketRenderingError(f"invalid rich packet type: {value}")
    return kind, args


def split_type_args(value: str) -> tuple[str, ...]:
    args: list[str] = []
    depth = 0
    start = 0
    for index, char in enumerate(value):
        if char == "<":
            depth += 1
        elif char == ">":
            depth -= 1
            if depth < 0:
                raise PacketRenderingError(f"invalid rich packet type arguments: {value}")
        elif char == "," and depth == 0:
            args.append(value[start:index])
            start = index + 1
    if depth != 0:
        raise PacketRenderingError(f"invalid rich packet type arguments: {value}")
    args.append(value[start:])
    return tuple(args)


def gdscript_field_constant(field_name: str) -> str:
    return f"FIELD_{constant_name(field_name)}"


def gdscript_leaf(value: object) -> str:
    if isinstance(value, str):
        if value.startswith("$"):
            return value[1:]
        type_name = GDSCRIPT_PACKET_TYPE_CONSTANTS.get(value)
        if type_name is not None:
            return type_name
        return json.dumps(value)
    if isinstance(value, bool):
        return "true" if value else "false"
    if value is None:
        return "null"
    return str(value)


def snake_to_go_name(value: str) -> str:
    return "".join(go_word(part) for part in value.split("_"))


def go_word(value: str) -> str:
    if value.lower() == "id":
        return "ID"
    return value[:1].upper() + value[1:]


def constant_name(value: str) -> str:
    return re.sub(r"[^A-Za-z0-9]+", "_", value).upper()
