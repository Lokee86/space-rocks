from __future__ import annotations

import re

import pytest

from data_sync.generators.rich_go_packets import RichGoPacketGenerationError, render_go_output
from data_sync.generators import gds_packets, go_packets, ts_packets
from data_sync.model.packets import (
    PacketBuilder,
    PacketDefinition,
    PacketField,
    PacketOutput,
    PacketSchema,
    PacketSchemaField,
    PacketStruct,
    PacketType,
)


PACKETS = (
    PacketDefinition(
        name="player_input",
        id=100,
        direction="client_to_server",
        fields=(
            PacketField("sequence", "uint32"),
            PacketField("turn", "float32"),
            PacketField("thrust", "bool"),
            PacketField("shoot", "bool"),
        ),
    ),
    PacketDefinition(
        name="client_config",
        id=101,
        direction="server_to_client",
        fields=(PacketField("self_id", "string"), PacketField("lives", "int")),
    ),
)


def test_go_packet_output() -> None:
    assert go_packets.generate_packets("packets", PACKETS) == """
const PacketPlayerInput = 100

type PlayerInputPacket struct {
    Sequence uint32
    Turn     float32
    Thrust   bool
    Shoot    bool
}

const PacketClientConfig = 101

type ClientConfigPacket struct {
    SelfId string
    Lives  int
}
""".strip()


def test_gds_packet_output() -> None:
    assert gds_packets.generate_packets("packets", PACKETS) == """
const PACKET_PLAYER_INPUT := 100
const PACKET_PLAYER_INPUT_FIELDS := ["sequence", "turn", "thrust", "shoot"]
const PACKET_CLIENT_CONFIG := 101
const PACKET_CLIENT_CONFIG_FIELDS := ["self_id", "lives"]
""".strip()


def test_ts_packet_output() -> None:
    assert ts_packets.generate_packets("packets", PACKETS) == """
export const PACKET_PLAYER_INPUT = 100;

export interface PlayerInputPacket {
  sequence: number;
  turn: number;
  thrust: boolean;
  shoot: boolean;
}

export const PACKET_CLIENT_CONFIG = 101;

export interface ClientConfigPacket {
  self_id: string;
  lives: number;
}
""".strip()


def test_field_ordering_is_preserved() -> None:
    output = gds_packets.generate_packets("packets", PACKETS)

    assert '["sequence", "turn", "thrust", "shoot"]' in output


def test_rich_go_packet_type_ids_scopes_packet_type_constants() -> None:
    schema = PacketSchema(
        outputs=(),
        structs=(
            PacketStruct(
                id="ExamplePacket",
                fields=(PacketSchemaField(name="type", json="type", type="string"),),
            ),
        ),
        packet_types=(
            PacketType(id="input", value="input"),
            PacketType(id="respawn", value="respawn"),
            PacketType(id="client_config", value="client_config"),
        ),
        builders=(PacketBuilder(id="input_packet", args=(), body={"type": "input"}),),
    )
    output = PacketOutput(
        id="server_game_packets",
        language="go",
        path="services/game-server/internal/game/packets.go",
        package="game",
        packet_types=True,
        packet_type_ids=("input", "client_config"),
        structs=("ExamplePacket",),
    )

    rendered = render_go_output(schema, output)

    assert re.search(r'^\s*PacketTypeInput\s*=\s*"input"$', rendered, re.MULTILINE)
    assert re.search(r'^\s*PacketTypeClientConfig\s*=\s*"client_config"$', rendered, re.MULTILINE)
    assert "PacketTypeRespawn" not in rendered


def test_rich_go_packet_type_ids_unknown_id_raises() -> None:
    schema = PacketSchema(
        outputs=(),
        structs=(),
        packet_types=(PacketType(id="input", value="input"),),
        builders=(),
    )
    output = PacketOutput(
        id="server_game_packets",
        language="go",
        path="services/game-server/internal/game/packets.go",
        package="game",
        packet_type_ids=("missing",),
    )

    with pytest.raises(RichGoPacketGenerationError):
        render_go_output(schema, output)
