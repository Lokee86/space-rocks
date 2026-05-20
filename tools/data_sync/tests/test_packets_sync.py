from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_project(tmp_path: Path) -> Path:
    for directory in ["shared", "go", "gds", "ts"]:
        (tmp_path / directory).mkdir()

    (tmp_path / "shared/game_data.toml").write_text(
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
turn = "float32"
thrust = "bool"
shoot = "bool"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
old
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "gds/constants.gd").write_text(
        """
# data-sync:start constants.client
old
# data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "ts/constants.ts").write_text(
        """
// data-sync:start constants.network
old
// data-sync:end constants.network
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text(
        """
package packets

// before packets
// data-sync:start packets
old
// data-sync:end packets
// after packets
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "gds/packets.gd").write_text(
        """
# data-sync:start packets
old
# data-sync:end packets
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "ts/packets.ts").write_text(
        """
// data-sync:start packets
old
// data-sync:end packets
""".lstrip(),
        encoding="utf-8",
    )

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
sections = ["constants.network"]
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
""".strip()
        + "\n",
        encoding="utf-8",
    )
    return config_path


def test_push_packets_updates_managed_block(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-packets", "-go", "-config", str(config_path)]) == 0

    output = (tmp_path / "go/packets.go").read_text(encoding="utf-8")
    assert "// before packets" in output
    assert "// after packets" in output
    assert "const PacketPlayerInput = 100" in output
    assert "type PlayerInputPacket struct {" in output
    assert "    Sequence uint32" in output


def test_diff_packets_writes_nothing(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_project(tmp_path)
    before = (tmp_path / "ts/packets.ts").read_text(encoding="utf-8")

    assert run(["-diff", "-packets", "-ts", "-config", str(config_path)]) == 0

    captured = capsys.readouterr()
    assert "+export const PACKET_PLAYER_INPUT = 100;" in captured.out
    assert (tmp_path / "ts/packets.ts").read_text(encoding="utf-8") == before


def test_check_packets_exit_codes(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-check", "-packets", "-gds", "-config", str(config_path)]) == 1
    assert run(["-push", "-packets", "-gds", "-config", str(config_path)]) == 0
    assert run(["-check", "-packets", "-gds", "-config", str(config_path)]) == 0


def test_combined_constants_and_packets_push(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-packets", "-go", "-config", str(config_path)]) == 0

    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(
        encoding="utf-8"
    )
    assert "const PacketPlayerInput = 100" in (tmp_path / "go/packets.go").read_text(
        encoding="utf-8"
    )


def test_combined_constants_and_packets_check(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-packets", "-go", "-gds", "-ts", "-config", str(config_path)]) == 0

    assert run(["-check", "-constants", "-packets", "-go", "-gds", "-ts", "-config", str(config_path)]) == 0
