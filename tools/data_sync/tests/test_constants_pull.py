from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_pull_project(tmp_path: Path) -> Path:
    for directory in ["shared", "go", "gds", "ts"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
debug_enabled = true
welcome_text = "hello"

[constants.client]
client_scale = 2

[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
sequence = "uint32"
shoot = "bool"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/constants.go").write_text(
        """
package constants

func untouched() {}
// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
const DebugEnabled = false
const WelcomeText = "hi"
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "gds/constants.gd").write_text(
        """
# data-sync:start constants.client
const CLIENT_SCALE := 4
# data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "ts/constants.ts").write_text(
        """
// data-sync:start constants.client
export const CLIENT_SCALE = 5;
// data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text("// data-sync:start packets\nold\n// data-sync:end packets\n")
    (tmp_path / "gds/packets.gd").write_text("# data-sync:start packets\nold\n# data-sync:end packets\n")
    (tmp_path / "ts/packets.ts").write_text("// data-sync:start packets\nold\n// data-sync:end packets\n")

    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot]
path = "shared/game_data.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[constants.gds]
files = ["gds/constants.gd"]
sections = ["constants.client"]
owns = ["constants.client"]

[constants.ts]
files = ["ts/constants.ts"]
sections = ["constants.client"]
owns = []

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
""".strip()
        + "\n",
        encoding="utf-8",
    )
    return config_path


def test_pull_constants_updates_only_owned_section_and_preserves_packets(tmp_path: Path) -> None:
    config_path = write_pull_project(tmp_path)

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot
    assert "debug_enabled = false" in sot
    assert 'welcome_text = "hi"' in sot
    assert "client_scale = 2" in sot
    assert "[packets.player_input.fields]" in sot
    assert 'sequence = "uint32"' in sot


def test_pull_constants_refuses_non_owned_section(tmp_path: Path) -> None:
    config_path = write_pull_project(tmp_path)

    assert run(["-pull", "-constants", "-ts", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "client_scale = 2" in sot
    assert "CLIENT_SCALE" not in sot


def test_pull_constants_refuses_noncanonical_formatting(tmp_path: Path) -> None:
    config_path = write_pull_project(tmp_path)
    (tmp_path / "go/constants.go").write_text(
        """
// data-sync:start constants.gameplay
const PlayerSpeed=500.0
const DebugEnabled = false
const WelcomeText = "hi"
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 1


def test_packet_pull_is_refused(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_pull_project(tmp_path)

    assert run(["-pull", "-packets", "-go", "-config", str(config_path)]) == 2
    captured = capsys.readouterr()
    assert "Packet pull is not supported. Edit packet schema files under shared/packets/." in captured.err


def test_pull_does_not_rewrite_language_file(tmp_path: Path) -> None:
    config_path = write_pull_project(tmp_path)
    before = (tmp_path / "go/constants.go").read_text(encoding="utf-8")

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    assert (tmp_path / "go/constants.go").read_text(encoding="utf-8") == before
