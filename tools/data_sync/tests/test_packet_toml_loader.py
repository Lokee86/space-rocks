from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.packet_toml import load_packet_schema


pytest.importorskip("tomlkit")


def migrated_packet_toml(tmp_path: Path) -> Path:
    repo_root = Path(__file__).resolve().parents[3]
    source = repo_root / "shared/packets/packets.toml"
    output = tmp_path / "packets.toml"
    output.write_text(source.read_text(encoding="utf-8"), encoding="utf-8")
    return output


def test_loads_migrated_packet_schema_outputs(tmp_path: Path) -> None:
    schema = load_packet_schema(migrated_packet_toml(tmp_path))

    assert len(schema.outputs) == 3

    game_output = schema.output_for_path("services/game-server/internal/game/packets.go")
    assert game_output.language == "go"
    assert game_output.package == "game"
    assert game_output.packet_types is True
    for required_struct in ("ClientPacket", "EventState", "StatePacket"):
        assert required_struct in game_output.structs
    assert any(name in game_output.structs for name in ("RoomSnapshot", "CreateRoomRequest"))
    assert game_output.imports == {
        "entities": "github.com/Lokee86/space-rocks/server/internal/game/entities",
    }

    gds_output = schema.output_for_path("client/scripts/networking/packets/packets.gd")
    assert gds_output.language == "gdscript"
    assert gds_output.base == "RefCounted"
    assert "input_packet" in gds_output.builders


def test_loads_migrated_structs_and_field_overrides(tmp_path: Path) -> None:
    schema = load_packet_schema(migrated_packet_toml(tmp_path))

    state_packet = schema.struct("StatePacket")
    fields = {field.name: field for field in state_packet.fields}

    assert fields["players"].type == "map"
    assert fields["players"].key_type == "string"
    assert fields["players"].value_type == "ShipState"
    assert fields["players"].go_value_type == "entities.ShipState"
    assert fields["events"].type == "array"
    assert fields["events"].item_type == "EventState"
    assert fields["self_id"].go_name == "SelfID"


def test_loads_packet_types_and_builders(tmp_path: Path) -> None:
    schema = load_packet_schema(migrated_packet_toml(tmp_path))

    assert [packet_type.id for packet_type in schema.packet_types][:3] == [
        "input",
        "client_config",
        "state",
    ]

    builder = schema.builder("input_packet")
    assert builder.args == ("forward", "back", "right", "left", "shoot")
    assert builder.body["type"] == "input"
    assert builder.body["input"]["forward"] == "$forward"


def test_preserves_rich_type_strings(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
[[outputs]]
language = "go"
path = "out.go"
package = "game"
structs = ["StatePacket"]

[[structs]]
id = "StatePacket"

[[structs.fields]]
name = "players"
json = "players"
type = "map<string,ShipState>"

[[structs.fields]]
name = "events"
json = "events"
type = "array<EventState>"

[[packet_types]]
id = "state"
value = "state"

[[builders]]
id = "state_packet"
args = []

[builders.body]
type = "state"
""".lstrip(),
        encoding="utf-8",
    )

    schema = load_packet_schema(path)

    fields = schema.struct("StatePacket").fields
    assert fields[0].type == "map<string,ShipState>"
    assert fields[1].type == "array<EventState>"
