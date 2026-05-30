from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.model.packets import PacketDefinition, PacketField
from data_sync.toml_store import TomlStore


tomlkit = pytest.importorskip("tomlkit")


def write_sot(tmp_path: Path, text: str) -> Path:
    path = tmp_path / "game_data.toml"
    path.write_text(text.strip() + "\n", encoding="utf-8")
    return path


def sample_sot() -> str:
    return """
[constants.gameplay]
player_speed = 420.0
bullet_speed = 900.0
asteroid_spawn_interval = 1.5

[constants.network]
tick_rate = 60
max_players_per_room = 2

[unrelated]
keep_me = "yes"

[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
sequence = "uint32"
turn = "float32"
thrust = "bool"
shoot = "bool"

[packets.state]
id = "state"
direction = "server_to_client"

[packets.state.fields]
self_id = "string"
lives = "int"
""".strip()


def test_loading_constants_preserves_toml_order(tmp_path: Path) -> None:
    store = TomlStore.load(write_sot(tmp_path, sample_sot()))

    section = store.constants("constants.gameplay")

    assert section.name == "constants.gameplay"
    assert section.values == (
        ("player_speed", 420.0),
        ("bullet_speed", 900.0),
        ("asteroid_spawn_interval", 1.5),
    )


def test_loading_packets(tmp_path: Path) -> None:
    store = TomlStore.load(write_sot(tmp_path, sample_sot()))

    packets = store.packets()

    assert [packet.name for packet in packets] == ["player_input", "state"]
    assert packets[0].id == 100
    assert packets[0].direction == "client_to_server"
    assert packets[0].field_types() == {
        "sequence": "uint32",
        "turn": "float32",
        "thrust": "bool",
        "shoot": "bool",
    }


def test_packet_fields_preserve_toml_order(tmp_path: Path) -> None:
    store = TomlStore.load(write_sot(tmp_path, sample_sot()))

    packet = store.packet("player_input")

    assert [field.name for field in packet.fields] == ["sequence", "turn", "thrust", "shoot"]


def test_constants_section_with_child_table_returns_only_direct_scalars(tmp_path: Path) -> None:
    path = write_sot(
        tmp_path,
        """
[constants.client.presentation]
hud_scale = 1.25

[constants.client.presentation.sound]
master_volume = 0.6
""",
    )
    store = TomlStore.load(path)

    presentation = store.constants("constants.client.presentation")
    sound = store.constants("constants.client.presentation.sound")

    assert presentation.values == (("hud_scale", 1.25),)
    assert sound.values == (("master_volume", 0.6),)


def test_constants_section_with_direct_table_value_exposes_value_for_validation(tmp_path: Path) -> None:
    path = write_sot(
        tmp_path,
        """
[constants.client.presentation]
hud_scale = 1.25
sound = { master_volume = 0.6 }
""",
    )
    store = TomlStore.load(path)

    presentation = store.constants("constants.client.presentation")

    assert presentation.values == (
        ("hud_scale", 1.25),
        ("sound", {"master_volume": 0.6}),
    )


def test_update_one_constants_section_preserves_other_sections(tmp_path: Path) -> None:
    path = write_sot(tmp_path, sample_sot())
    store = TomlStore.load(path)

    store.update_constants(
        "constants.gameplay",
        {
            "player_speed": 500.0,
            "bullet_speed": 950.0,
        },
    )
    store.write()

    reloaded = TomlStore.load(path)
    assert reloaded.constants("constants.gameplay").values == (
        ("player_speed", 500.0),
        ("bullet_speed", 950.0),
    )
    assert reloaded.constants("constants.network").values == (
        ("tick_rate", 60),
        ("max_players_per_room", 2),
    )
    assert reloaded.document["unrelated"]["keep_me"] == "yes"
    assert reloaded.packet("player_input").direction == "client_to_server"


def test_update_packet_and_write_back_valid_toml(tmp_path: Path) -> None:
    path = write_sot(tmp_path, sample_sot())
    store = TomlStore.load(path)

    store.update_packet(
        PacketDefinition(
            name="player_input",
            id=101,
            direction="client_to_server",
            fields=(
                PacketField("sequence", "uint32"),
                PacketField("shoot", "bool"),
            ),
        )
    )
    store.write()

    parsed = tomlkit.parse(path.read_text(encoding="utf-8"))
    assert parsed["packets"]["player_input"]["id"] == 101
    assert list(parsed["packets"]["player_input"]["fields"]) == ["sequence", "shoot"]

    reloaded = TomlStore.load(path)
    assert [field.name for field in reloaded.packet("player_input").fields] == ["sequence", "shoot"]
