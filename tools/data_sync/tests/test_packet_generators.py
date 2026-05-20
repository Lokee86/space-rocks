from __future__ import annotations

from data_sync.generators import gds_packets, go_packets, ts_packets
from data_sync.model.packets import PacketDefinition, PacketField


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
        name="state",
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

const PacketState = 101

type StatePacket struct {
    SelfId string
    Lives  int
}
""".strip()


def test_gds_packet_output() -> None:
    assert gds_packets.generate_packets("packets", PACKETS) == """
const PACKET_PLAYER_INPUT := 100
const PACKET_PLAYER_INPUT_FIELDS := ["sequence", "turn", "thrust", "shoot"]
const PACKET_STATE := 101
const PACKET_STATE_FIELDS := ["self_id", "lives"]
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

export const PACKET_STATE = 101;

export interface StatePacket {
  self_id: string;
  lives: number;
}
""".strip()


def test_field_ordering_is_preserved() -> None:
    output = gds_packets.generate_packets("packets", PACKETS)

    assert '["sequence", "turn", "thrust", "shoot"]' in output
