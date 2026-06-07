from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_discovered_pull_project(tmp_path: Path) -> Path:
    for directory in ["shared", "go", "gds"]:
        (tmp_path / directory).mkdir()

    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.client]
client_scale = 2
""".strip()
        + "\n",
        encoding="utf-8",
    )

    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 420.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )

    (tmp_path / "gds/constants.gd").write_text(
        """
extends RefCounted

# data-sync:start constants.client
const CLIENT_SCALE := 2
# data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )

    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd"]
exclude = []
""".strip()
        + "\n",
        encoding="utf-8",
    )
    return config_path


def test_pull_constants_updates_toml_from_discovered_go_marker(tmp_path: Path) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot


def test_pull_constants_updates_toml_from_discovered_gds_marker(tmp_path: Path) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    (tmp_path / "gds/constants.gd").write_text(
        """
extends RefCounted

# data-sync:start constants.client
const CLIENT_SCALE := 4
# data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-gds", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "client_scale = 4" in sot


def test_pull_constants_accepts_duplicate_discovered_blocks_when_values_match(tmp_path: Path) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/constants_copy.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot


def test_pull_constants_rejects_duplicate_discovered_blocks_when_values_differ(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/constants_copy.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 600.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    before = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")

    exit_code = run(["-pull", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 1
    assert (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8") == before
    assert "conflicting values" in captured.err or "constants.gameplay" in captured.err


def test_pull_constants_rejects_key_renames_from_discovered_block(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    game_data_path = tmp_path / "shared/game_data.toml"
    game_data_path.write_text(
        """
[constants.gameplay]
player_speed = 420.0
tick_rate = 60
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
const TickSpeed = 30
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    before = game_data_path.read_text(encoding="utf-8")

    exit_code = run(["-pull", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 1
    assert game_data_path.read_text(encoding="utf-8") == before
    assert "pull may only update existing values" in captured.err
    assert "expected keys" in captured.err
    assert "found" in captured.err


def test_pull_constants_updates_multiple_sections_in_same_toml_file(tmp_path: Path) -> None:
    config_path = write_discovered_pull_project(tmp_path)
    game_data_path = tmp_path / "shared/game_data.toml"
    game_data_path.write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.server.damage]
collision_damage = 1
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
// data-sync:start constants.server.damage
const CollisionDamage = 9
// data-sync:end constants.server.damage
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = game_data_path.read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot
    assert "collision_damage = 9" in sot
