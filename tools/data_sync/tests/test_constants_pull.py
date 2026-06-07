from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_pull_project(tmp_path: Path) -> Path:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
debug_enabled = true
welcome_text = "hello"

[constants.server.weapons.basic_cannon]
basic_cannon_projectile_speed = 1200.0

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
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
const BasicCannonProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.basic_cannon
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text("// data-sync:start packets\nold\n// data-sync:end packets\n")

    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.basic_cannon"]
owns = ["constants.server.weapons.basic_cannon"]

""".strip()
        + "\n",
        encoding="utf-8",
    )
    return config_path


def test_pull_constants_updates_gameplay_and_preserves_packets(tmp_path: Path) -> None:
    config_path = write_pull_project(tmp_path)

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot
    assert "debug_enabled = false" in sot
    assert 'welcome_text = "hi"' in sot
    assert "client_scale = 2" in sot


def test_pull_constants_can_load_only_go_constants_outputs(tmp_path: Path) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
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
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot


def test_pull_constants_can_use_arbitrary_go_outputs_without_legacy_constants_table(tmp_path: Path) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.server.weapons.basic_cannon]
basic_cannon_projectile_speed = 1200.0
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
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
const BasicCannonProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.basic_cannon
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.basic_cannon"]
owns = ["constants.server.weapons.basic_cannon"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    sot = (tmp_path / "shared/game_data.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in sot
    assert "basic_cannon_projectile_speed = 900.0" in sot


def test_pull_constants_updates_general_and_weapons_sot_files(tmp_path: Path) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/constants.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
debug_enabled = true

""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/weapons.toml").write_text(
        """
[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1200.0

""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/constants.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
const DebugEnabled = false
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.torpedo
const TorpedoProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.torpedo
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/constants.toml", "shared/weapons.toml"]

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.torpedo"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    general_sot = (tmp_path / "shared/constants.toml").read_text(encoding="utf-8")
    weapons_sot = (tmp_path / "shared/weapons.toml").read_text(encoding="utf-8")
    assert "player_speed = 500.0" in general_sot
    assert "debug_enabled = false" in general_sot
    assert "torpedo_projectile_speed = 900.0" in weapons_sot


def test_pull_constants_succeeds_when_sections_are_split_across_sot_files(tmp_path: Path) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/general.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/weapons.toml").write_text(
        """
[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1200.0
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
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.torpedo
const TorpedoProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.torpedo
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/general.toml", "shared/weapons.toml"]

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.torpedo"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 0

    assert "player_speed = 500.0" in (tmp_path / "shared/general.toml").read_text(encoding="utf-8")
    assert "torpedo_projectile_speed = 900.0" in (tmp_path / "shared/weapons.toml").read_text(encoding="utf-8")


def test_pull_constants_fails_when_owned_section_exists_in_zero_sot_files(tmp_path: Path) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/general.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.torpedo
const TorpedoProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.torpedo
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/general.toml"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.torpedo"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 1


def test_pull_constants_fails_when_owned_section_exists_in_multiple_sot_files(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/general.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1100.0
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/weapons.toml").write_text(
        """
[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1200.0
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.torpedo
const TorpedoProjectileSpeed = 900.0
// data-sync:end constants.server.weapons.torpedo
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/general.toml", "shared/weapons.toml"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.torpedo"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 1
    captured = capsys.readouterr()
    assert "duplicate constants source" in captured.err or "duplicate source section" in captured.err


def test_pull_constants_fails_when_section_is_owned_by_multiple_outputs(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    for directory in ["shared", "go"]:
        (tmp_path / directory).mkdir()
    (tmp_path / "shared/general.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
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
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.gameplay
const PlayerSpeed = 500.0
// data-sync:end constants.gameplay
""".lstrip(),
        encoding="utf-8",
    )
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot.constants]
path = "shared/general.toml"

[constants.go]
files = ["go/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["go/weapons.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert run(["-pull", "-constants", "-go", "-config", str(config_path)]) == 2
    captured = capsys.readouterr()
    assert "config error" in captured.err
    assert "owned by multiple targets" in captured.err


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
