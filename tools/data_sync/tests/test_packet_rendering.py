from __future__ import annotations

import pytest

from data_sync.generators.rich_gds_packets import RichGdsPacketGenerationError, render_gdscript_output
from data_sync.model.packets import PacketSchemaField
from data_sync.model.packets import PacketBuilder, PacketOutput, PacketSchema, PacketStruct, PacketType
from data_sync.packet_rendering import (
    PacketRenderingError,
    gdscript_field_constant,
    gdscript_leaf,
    go_field_name,
    go_json_tag,
    go_type_for_field,
    parse_rich_type,
)


def field(name: str, field_type: str, **kwargs: object) -> PacketSchemaField:
    return PacketSchemaField(name=name, json=kwargs.pop("json", name), type=field_type, **kwargs)


@pytest.mark.parametrize(
    ("field_type", "go_type"),
    [
        ("string", "string"),
        ("float", "float64"),
        ("int", "int"),
        ("bool", "bool"),
        ("uint32", "uint32"),
        ("float32", "float32"),
        ("float64", "float64"),
    ],
)
def test_go_primitive_types(field_type: str, go_type: str) -> None:
    assert go_type_for_field(field("value", field_type)) == go_type


def test_go_array_and_map_fields_from_expanded_schema() -> None:
    assert go_type_for_field(
        field("events", "array", item_type="EventState"),
    ) == "[]EventState"
    assert go_type_for_field(
        field(
            "players",
            "map",
            key_type="string",
            value_type="ShipState",
            go_value_type="runtime.ShipState",
        ),
    ) == "map[string]runtime.ShipState"


def test_go_rich_type_strings() -> None:
    assert go_type_for_field(field("players", "map<string,ShipState>")) == "map[string]ShipState"
    assert go_type_for_field(field("events", "array<EventState>")) == "[]EventState"
    assert go_type_for_field(field("events", "list<EventState>")) == "[]EventState"
    assert go_type_for_field(field("players", "dictionary<string,ShipState>")) == "map[string]ShipState"


def test_go_overrides_and_custom_struct_refs() -> None:
    assert go_type_for_field(
        field("input", "InputState", go_type="entities.InputState"),
    ) == "entities.InputState"
    assert go_type_for_field(field("ship", "ShipState")) == "ShipState"
    assert go_type_for_field(
        field("children", "array", item_type="ChildState", go_item_type="entities.ChildState"),
    ) == "[]entities.ChildState"


def test_go_field_names_and_json_tags() -> None:
    assert go_field_name(field("owner_id", "string")) == "OwnerID"
    assert go_field_name(field("player_id", "string", go_name="PlayerID")) == "PlayerID"
    assert go_json_tag(field("self_id", "string")) == 'json:"self_id"'


def test_invalid_map_shape_fails_loudly() -> None:
    with pytest.raises(PacketRenderingError, match="requires key_type and value_type"):
        go_type_for_field(field("players", "map", key_type="string"))

    with pytest.raises(PacketRenderingError, match="unsupported scalar Go type"):
        go_type_for_field(field("players", "map<boolish,ShipState>"))


def test_parse_nested_rich_type_arguments() -> None:
    assert parse_rich_type("map<string,array<EventState>>") == (
        "map",
        ("string", "array<EventState>"),
    )


def test_gdscript_field_constants_and_leaf_values_match_current_generator() -> None:
    assert gdscript_field_constant("player_id") == "FIELD_PLAYER_ID"
    assert gdscript_leaf("$forward") == "forward"
    assert gdscript_leaf("input") == "TYPE_INPUT"
    assert gdscript_leaf("respawn") == "TYPE_RESPAWN"
    assert gdscript_leaf("pause_player") == '"pause_player"'
    assert gdscript_leaf(True) == "true"
    assert gdscript_leaf(False) == "false"
    assert gdscript_leaf(None) == "null"


def test_rich_gds_packet_type_ids_scopes_packet_type_constants() -> None:
    schema = PacketSchema(
        outputs=(),
        structs=(PacketStruct(id="ExamplePacket", fields=(field("type", "string"),)),),
        packet_types=(
            PacketType(id="input", value="input"),
            PacketType(id="respawn", value="respawn"),
            PacketType(id="client_config", value="client_config"),
        ),
        builders=(PacketBuilder(id="input_packet", args=(), body={"type": "input"}),),
    )
    output = PacketOutput(
        id="client_packets",
        language="gdscript",
        path="client/scripts/generated/networking/packets/packets.gd",
        packet_type_ids=("input", "client_config"),
        builders=("input_packet",),
    )

    rendered = render_gdscript_output(schema, output)

    assert 'const TYPE_INPUT := "input"' in rendered
    assert 'const TYPE_CLIENT_CONFIG := "client_config"' in rendered
    assert "TYPE_RESPAWN" not in rendered


def test_rich_gds_packet_type_ids_unknown_id_raises() -> None:
    schema = PacketSchema(
        outputs=(),
        structs=(),
        packet_types=(PacketType(id="input", value="input"),),
        builders=(),
    )
    output = PacketOutput(
        id="client_packets",
        language="gdscript",
        path="client/scripts/generated/networking/packets/packets.gd",
        packet_type_ids=("missing",),
    )

    with pytest.raises(RichGdsPacketGenerationError):
        render_gdscript_output(schema, output)

