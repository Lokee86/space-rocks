from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_validation_project(tmp_path: Path) -> Path:
    for directory in ["shared", "go", "gds", "ts"]:
        (tmp_path / directory).mkdir()

    write_sot(
        tmp_path,
        """
[constants.gameplay]
player_speed = 420.0

[constants.client]
client_scale = 2

[constants.network]
max_players = 2

[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
sequence = "uint32"
shoot = "bool"

[packets.state]
id = 101
direction = "server_to_client"

[packets.state.fields]
self_id = "string"
""",
    )

    (tmp_path / "go/constants.go").write_text(
        block("//", "constants.gameplay") + block("//", "constants.network"),
        encoding="utf-8",
    )
    (tmp_path / "gds/constants.gd").write_text(
        block("#", "constants.gameplay") + block("#", "constants.client"),
        encoding="utf-8",
    )
    (tmp_path / "ts/constants.ts").write_text(
        block("//", "constants.network") + block("//", "constants.client"),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text(block("//", "packets"), encoding="utf-8")
    (tmp_path / "gds/packets.gd").write_text(block("#", "packets"), encoding="utf-8")
    (tmp_path / "ts/packets.ts").write_text(block("//", "packets"), encoding="utf-8")

    config_path = tmp_path / "config.toml"
    config_path.write_text(valid_config_text(), encoding="utf-8")
    return config_path


def valid_config_text() -> str:
    return """
[sot]
path = "shared/game_data.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay", "constants.network"]
owns = ["constants.gameplay"]

[constants.gds]
files = ["gds/constants.gd"]
sections = ["constants.gameplay", "constants.client"]
owns = ["constants.client"]

[constants.ts]
files = ["ts/constants.ts"]
sections = ["constants.network", "constants.client"]
owns = ["constants.network"]

[packets.go]
files = ["go/packets.go"]
sections = ["packets"]
owns = ["packets"]

[packets.gds]
files = ["gds/packets.gd"]
sections = ["packets"]
owns = []

[packets.ts]
files = ["ts/packets.ts"]
sections = ["packets"]
owns = []
""".strip() + "\n"


def write_sot(tmp_path: Path, text: str) -> None:
    (tmp_path / "shared/game_data.toml").write_text(text.strip() + "\n", encoding="utf-8")


def block(comment: str, section: str) -> str:
    return f"{comment} data-sync:start {section}\nold\n{comment} data-sync:end {section}\n"


def test_validate_valid_config_and_sot(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    assert run(["-validate", "-config", str(config_path)]) == 0


def test_validate_invalid_constants_ownership_overlap(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    config_path.write_text(
        valid_config_text().replace(
            'owns = ["constants.client"]',
            'owns = ["constants.gameplay"]',
        ),
        encoding="utf-8",
    )

    assert run(["-validate", "-config", str(config_path)]) == 2


def test_validate_duplicate_packet_ids(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    write_sot(
        tmp_path,
        """
[constants.gameplay]
player_speed = 420.0

[constants.client]
client_scale = 2

[constants.network]
max_players = 2

[packets.player_input]
id = 100
direction = "client_to_server"

[packets.state]
id = 100
direction = "server_to_client"
""",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_unsupported_packet_field_type(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    write_sot(
        tmp_path,
        """
[constants.gameplay]
player_speed = 420.0

[constants.client]
client_scale = 2

[constants.network]
max_players = 2

[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
position = "vector2"
""",
    )

    assert run(["-validate", "-packets", "-config", str(config_path)]) == 1


def test_validate_bad_constant_name(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    write_sot(
        tmp_path,
        """
[constants.gameplay]
PlayerSpeed = 420.0

[constants.client]
client_scale = 2

[constants.network]
max_players = 2
""",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 1


def test_validate_missing_generated_block(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "go/constants.go").write_text(
        block("//", "constants.gameplay"),
        encoding="utf-8",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 1


def test_validate_missing_configured_file(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "ts/packets.ts").unlink()

    assert run(["-validate", "-packets", "-ts", "-config", str(config_path)]) == 1
