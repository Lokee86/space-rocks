from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.packet_toml import PacketTomlError, load_packet_schema, load_packet_schema_files


pytest.importorskip("tomlkit")


def migrated_packet_toml_paths(tmp_path: Path) -> tuple[Path, ...]:
    repo_root = Path(__file__).resolve().parents[3]
    relative_paths = (
        "shared/packets/outputs.toml",
        "shared/packets/gameplay.toml",
        "shared/packets/debug.toml",
        "shared/packets/lobby.toml",
    )
    outputs: list[Path] = []
    for relative_path in relative_paths:
        source = repo_root / relative_path
        output = tmp_path / Path(relative_path).name
        output.write_text(source.read_text(encoding="utf-8"), encoding="utf-8")
        outputs.append(output)
    return tuple(outputs)


def test_loads_migrated_packet_schema_outputs(tmp_path: Path) -> None:
    schema = load_packet_schema_files(migrated_packet_toml_paths(tmp_path))

    output_ids = {output.id for output in schema.outputs if output.id}
    assert {"server_game_packets", "server_realtime_packets", "server_devtools_packets"}.issubset(output_ids)
    assert "server_devtools_packets" in output_ids

    game_output = schema.output_for_path("services/game-server/internal/game/packets.go")
    assert game_output.language == "go"
    assert game_output.package == "game"
    assert game_output.packet_types is True
    assert "ClientPacket" in game_output.structs
    assert "EventState" in game_output.structs
    assert "ExcludedPacket" not in game_output.structs
    assert any(name in game_output.structs for name in ("RoomSnapshot", "CreateRoomRequest"))
    assert game_output.imports == {
        "runtime": "github.com/Lokee86/space-rocks/server/internal/game/runtime",
    }

    realtime_output = schema.output_for_id("server_realtime_packets")
    assert realtime_output.language == "go"
    assert realtime_output.package == "realtime"
    assert realtime_output.path == "services/game-server/internal/protocol/realtime/packets_generated.go"
    assert realtime_output.packet_types is True
    assert realtime_output.structs == ()
    assert realtime_output.packet_type_ids == (
        "world_full",
        "world_delta",
        "overlay_full",
        "overlay_delta",
        "session_full",
        "session_delta",
        "event_batch",
        "resync_request",
        "resync_required",
    )

    gds_output = schema.output_for_path("client/scripts/generated/networking/packets/packets.gd")
    assert gds_output.language == "gdscript"
    assert gds_output.base == "RefCounted"
    assert "input_packet" in gds_output.builders

    devtools_output = schema.output_for_id("server_devtools_packets")
    assert devtools_output.language == "go"
    assert devtools_output.package == "devtools"
    assert devtools_output.path == "services/game-server/internal/devtools/packets_generated.go"
    assert "DebugCommand" in devtools_output.structs
    assert "DebugStatus" in devtools_output.structs


def test_loads_migrated_structs_and_field_overrides(tmp_path: Path) -> None:
    schema = load_packet_schema_files(migrated_packet_toml_paths(tmp_path))

    with pytest.raises(KeyError):
        schema.struct("ExamplePacket")

    fields = {field.name: field for field in schema.struct("ClientPacket").fields}

    assert fields["input"].type == "InputState"
    assert fields["input"].go_type == "runtime.InputState"


def test_loads_packet_types_and_builders(tmp_path: Path) -> None:
    schema = load_packet_schema_files(migrated_packet_toml_paths(tmp_path))

    assert [packet_type.id for packet_type in schema.packet_types][:2] == [
        "input",
        "client_config",
    ]

    builder = schema.builder("input_packet")
    assert builder.args == ("forward", "back", "right", "left", "primary_fire", "secondary_fire")
    assert builder.body["type"] == "input"
    assert builder.body["input"]["forward"] == "$forward"
    assert builder.body["input"]["primary_fire"] == "$primary_fire"
    assert builder.body["input"]["secondary_fire"] == "$secondary_fire"


def test_preserves_rich_type_strings(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
[[outputs]]
language = "go"
path = "out.go"
package = "game"
builders = []
structs = ["ExamplePacket"]

[[structs]]
id = "ExamplePacket"

[[structs.fields]]
name = "players"
json = "players"
type = "map<string,ShipState>"

[[structs.fields]]
name = "events"
json = "events"
type = "array<EventState>"

[[packet_types]]
id = "client_config"
value = "client_config"
""".lstrip(),
        encoding="utf-8",
    )

    schema = load_packet_schema(path)

    fields = schema.struct("ExamplePacket").fields
    assert fields[0].type == "map<string,ShipState>"
    assert fields[1].type == "array<EventState>"


def test_supports_output_lookup_by_id_and_path(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
[[outputs]]
id = "go_packets"
language = "go"
path = "some/path.go"
package = "game"
builders = []
structs = ["ExamplePacket"]

[[outputs]]
id = "gds_packets"
language = "gdscript"
path = "some/path.gd"
base = "RefCounted"
builders = []
structs = ["ExamplePacket"]

[[structs]]
id = "ExamplePacket"

[[structs.fields]]
name = "players"
json = "players"
type = "map<string,ShipState>"

[[packet_types]]
id = "client_config"
value = "client_config"
""".lstrip(),
        encoding="utf-8",
    )

    schema = load_packet_schema(path)

    assert schema.outputs[0].id == "go_packets"
    assert schema.outputs[1].id == "gds_packets"
    assert schema.output_for_id("gds_packets").path == "some/path.gd"
    assert schema.output_for_path("some/path.gd").id == "gds_packets"
    with pytest.raises(KeyError):
        schema.output_for_id("missing_output")


def test_output_packet_type_ids_are_loaded(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
builders = []

[[outputs]]
id = "server_game_packets"
language = "go"
path = "services/game-server/internal/game/packets.go"
package = "game"
packet_types = true
packet_type_ids = ["input"]
structs = ["ExamplePacket"]

[[structs]]
id = "ExamplePacket"

[[structs.fields]]
name = "type"
json = "type"
type = "string"

[[packet_types]]
id = "input"
value = "input"

""".lstrip(),
        encoding="utf-8",
    )

    schema = load_packet_schema(path)

    output = schema.output_for_id("server_game_packets")
    assert output.packet_type_ids == ("input",)


def test_output_packet_type_ids_default_empty(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
builders = []

[[outputs]]
id = "server_game_packets"
language = "go"
path = "services/game-server/internal/game/packets.go"
package = "game"
packet_types = true
structs = ["ExamplePacket"]

[[structs]]
id = "ExamplePacket"

[[structs.fields]]
name = "type"
json = "type"
type = "string"

[[packet_types]]
id = "input"
value = "input"
""".lstrip(),
        encoding="utf-8",
    )

    schema = load_packet_schema(path)

    output = schema.output_for_id("server_game_packets")
    assert output.packet_type_ids == ()


def test_output_packet_type_ids_requires_list_of_strings(tmp_path: Path) -> None:
    path = tmp_path / "packets.toml"
    path.write_text(
        """
[[outputs]]
id = "server_game_packets"
language = "go"
path = "services/game-server/internal/game/packets.go"
package = "game"
packet_types = true
packet_type_ids = [1, "client_config"]
structs = ["ExamplePacket"]

[[structs]]
id = "ExamplePacket"

[[structs.fields]]
name = "type"
json = "type"
type = "string"

""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema(path)


def test_multi_file_rejects_duplicate_output_ids(tmp_path: Path) -> None:
    first = tmp_path / "first.toml"
    second = tmp_path / "second.toml"
    first.write_text(
        """
[[outputs]]
id = "dup_output"
language = "go"
path = "one.go"
""".lstrip(),
        encoding="utf-8",
    )
    second.write_text(
        """
[[outputs]]
id = "dup_output"
language = "gdscript"
path = "two.gd"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((first, second))


def test_multi_file_rejects_duplicate_output_paths(tmp_path: Path) -> None:
    first = tmp_path / "first.toml"
    second = tmp_path / "second.toml"
    first.write_text(
        """
[[outputs]]
id = "go_output"
language = "go"
path = "same/path.go"
""".lstrip(),
        encoding="utf-8",
    )
    second.write_text(
        """
[[outputs]]
id = "gds_output"
language = "gdscript"
path = "same/path.go"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((first, second))


def test_multi_file_rejects_duplicate_struct_ids(tmp_path: Path) -> None:
    first = tmp_path / "first.toml"
    second = tmp_path / "second.toml"
    first.write_text(
        """
[[structs]]
id = "ShipState"
[[structs.fields]]
name = "x"
json = "x"
type = "float"
""".lstrip(),
        encoding="utf-8",
    )
    second.write_text(
        """
[[structs]]
id = "ShipState"
[[structs.fields]]
name = "y"
json = "y"
type = "float"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((first, second))


def test_multi_file_rejects_duplicate_packet_type_ids(tmp_path: Path) -> None:
    first = tmp_path / "first.toml"
    second = tmp_path / "second.toml"
    first.write_text(
        """
[[packet_types]]
id = "input"
value = "input"
""".lstrip(),
        encoding="utf-8",
    )
    second.write_text(
        """
[[packet_types]]
id = "input"
value = "input_two"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((first, second))


def test_multi_file_rejects_duplicate_builder_ids(tmp_path: Path) -> None:
    first = tmp_path / "first.toml"
    second = tmp_path / "second.toml"
    first.write_text(
        """
[[builders]]
id = "input_packet"
args = []
[builders.body]
type = "input"
""".lstrip(),
        encoding="utf-8",
    )
    second.write_text(
        """
[[builders]]
id = "input_packet"
args = []
[builders.body]
type = "input_alt"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((first, second))


def test_multi_file_legacy_packet_toml_raises_packet_toml_error(tmp_path: Path) -> None:
    legacy = tmp_path / "legacy_packets.toml"
    legacy.write_text(
        """
[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
sequence = "uint32"
shoot = "bool"
""".lstrip(),
        encoding="utf-8",
    )

    with pytest.raises(PacketTomlError):
        load_packet_schema_files((legacy,))
