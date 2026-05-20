from __future__ import annotations

from pathlib import Path

import pytest

from main import run
from tests.test_packets_sync import write_project


pytest.importorskip("tomlkit")


def test_validate_packets_accepts_rich_packet_toml(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-validate", "-packets", "-go", "-gds", "-config", str(config_path)]) == 0


def test_validate_packets_rejects_duplicate_packet_type_ids(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    append_packet_toml(
        tmp_path,
        """
[[packet_types]]
id = "input"
value = "input_duplicate"
""",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_packets_rejects_unknown_output_struct(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        packets_path.read_text(encoding="utf-8").replace(
            'structs = ["PlayerInputPacket"]',
            'structs = ["MissingPacket"]',
        ),
        encoding="utf-8",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_packets_rejects_builder_unknown_field(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        packets_path.read_text(encoding="utf-8").replace(
            'sequence = "$sequence"',
            'sequence = "$sequence"\nmissing = "$sequence"',
        ),
        encoding="utf-8",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_packets_rejects_incomplete_map_field(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        packets_path.read_text(encoding="utf-8").replace(
            """
name = "sequence"
json = "sequence"
type = "uint32"
""".strip(),
            """
name = "sequence"
json = "sequence"
type = "map"
key_type = "string"
""".strip(),
        ),
        encoding="utf-8",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_packets_accepts_rich_type_strings(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        rich_type_packet_toml(),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text("generated\n", encoding="utf-8")

    assert run(["-validate", "-packets", "-go", "-config", str(config_path)]) == 0


def test_validate_packets_rejects_absolute_output_path(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        packets_path.read_text(encoding="utf-8").replace(
            'path = "go/packets.go"',
            'path = "/tmp/packets.go"',
        ),
        encoding="utf-8",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def append_packet_toml(tmp_path: Path, text: str) -> None:
    packets_path = tmp_path / "shared/packets/packets.toml"
    packets_path.write_text(
        packets_path.read_text(encoding="utf-8") + "\n" + text.lstrip(),
        encoding="utf-8",
    )


def rich_type_packet_toml() -> str:
    return """
[[outputs]]
language = "go"
path = "go/packets.go"
package = "packets"
structs = ["ShipState", "EventState", "StatePacket"]

[[structs]]
id = "ShipState"

[[structs.fields]]
name = "id"
json = "id"
type = "string"

[[structs]]
id = "EventState"

[[structs.fields]]
name = "type"
json = "type"
type = "string"

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
""".lstrip()
