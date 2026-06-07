from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_project(tmp_path: Path) -> Path:
    (tmp_path / "shared").mkdir()
    (tmp_path / "go").mkdir()
    (tmp_path / "gds").mkdir()
    (tmp_path / "ts").mkdir()

    (tmp_path / "shared/game_data.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0
tick_rate = 60
debug_enabled = true
welcome_text = "hello"

[constants.client]
client_scale = 2

[constants.network]
max_players = 2

[constants.server.weapons.basic_cannon]
basic_cannon_projectile_speed = 1200.0
basic_cannon_projectile_lifetime = 1.75
basic_cannon_cooldown = 0.22
basic_cannon_projectile_spawn_offset = 42.0
basic_cannon_damage = 1

[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1200.0
torpedo_projectile_lifetime = 1.75
torpedo_cooldown = 0.22
torpedo_projectile_spawn_offset = 42.0
torpedo_impact_damage = 1
torpedo_radial_damage = 1
torpedo_radial_zone_spawn_seconds = 0.1
torpedo_radial_tick_seconds = 0.1
torpedo_radial_total_seconds = 0.4
torpedo_radial_zone_lifetime_seconds = 0.4

[constants.shared.weapons.torpedo_radial_shape]
torpedo_radial_zone_count = 4
torpedo_radial_zone_width = 10
""".strip()
        + "\n",
        encoding="utf-8",
    )

    (tmp_path / "go/constants.go").write_text(
        """
package constants

// keep before
// data-sync:start constants.gameplay
old
// data-sync:end constants.gameplay
// keep after
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
old
// data-sync:end constants.server.weapons.basic_cannon
// data-sync:start constants.server.weapons.torpedo
old
// data-sync:end constants.server.weapons.torpedo
// data-sync:start constants.shared.weapons.torpedo_radial_shape
old
// data-sync:end constants.shared.weapons.torpedo_radial_shape
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "gds/constants.gd").write_text(
        """
extends RefCounted

# data-sync:start constants.client
old
# data-sync:end constants.client
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "gds/weapons.gd").write_text(
        """
extends RefCounted

# data-sync:start constants.server.weapons.basic_cannon
old
# data-sync:end constants.server.weapons.basic_cannon
""".lstrip(),
        encoding="utf-8",
    )
    (tmp_path / "ts/constants.ts").write_text(
        """
// untouched ts
// data-sync:start constants.network
old
// data-sync:end constants.network
""".lstrip(),
        encoding="utf-8",
    )

    config_path = tmp_path / "config.toml"
    config_path.write_text(
        """
[sot]
path = "shared/game_data.toml"

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd", "ts/**/*.ts"]
exclude = []
""".strip()
        + "\n",
        encoding="utf-8",
    )
    return config_path


def test_push_updates_only_managed_block(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    exit_code = run(["-push", "-constants", "-go", "-config", str(config_path)])

    assert exit_code == 0
    assert (tmp_path / "go/constants.go").read_text(encoding="utf-8") == (
        """
package constants

// keep before
// data-sync:start constants.gameplay
const PlayerSpeed = 420.0
const TickRate = 60
const DebugEnabled = true
const WelcomeText = "hello"
// data-sync:end constants.gameplay
// keep after
"""
        .lstrip()
        .strip()
        + "\n"
    )


def test_push_updates_new_toml_section_with_matching_marker_without_config_section_list(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    game_data_path = tmp_path / "shared/game_data.toml"
    game_data_path.write_text(
        """
[constants.gameplay]
player_speed = 420.0
tick_rate = 60
debug_enabled = true
welcome_text = "hello"

[constants.client]
client_scale = 2

[constants.network]
max_players = 2

[constants.server.weapons.basic_cannon]
basic_cannon_projectile_speed = 1200.0
basic_cannon_projectile_lifetime = 1.75
basic_cannon_cooldown = 0.22
basic_cannon_projectile_spawn_offset = 42.0
basic_cannon_damage = 1

[constants.server.weapons.torpedo]
torpedo_projectile_speed = 1200.0
torpedo_projectile_lifetime = 1.75
torpedo_cooldown = 0.22
torpedo_projectile_spawn_offset = 42.0
torpedo_impact_damage = 1
torpedo_radial_damage = 1
torpedo_radial_zone_spawn_seconds = 0.1
torpedo_radial_tick_seconds = 0.1
torpedo_radial_total_seconds = 0.4
torpedo_radial_zone_lifetime_seconds = 0.4

[constants.shared.weapons.torpedo_radial_shape]
torpedo_radial_zone_count = 4
torpedo_radial_zone_width = 10

[constants.server.damage]
collision_damage = 9
""".strip()
        + "\n",
        encoding="utf-8",
    )
    damage_path = tmp_path / "go/damage.go"
    damage_path.write_text(
        """
package constants

// data-sync:start constants.server.damage
old
// data-sync:end constants.server.damage
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0
    assert damage_path.read_text(encoding="utf-8") == (
        """
package constants

// data-sync:start constants.server.damage
const CollisionDamage = 9
// data-sync:end constants.server.damage
""".lstrip()
    )


def test_push_updates_all_constants_outputs_for_language(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    exit_code = run(["-push", "-constants", "-go", "-config", str(config_path)])

    assert exit_code == 0
    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(encoding="utf-8")
    assert "const BasicCannonProjectileSpeed = " in (tmp_path / "go/weapons.go").read_text(encoding="utf-8")

    exit_code = run(["-push", "-constants", "-gds", "-config", str(config_path)])

    assert exit_code == 0
    assert "CLIENT_SCALE := 2" in (tmp_path / "gds/constants.gd").read_text(encoding="utf-8")
    assert "BASIC_CANNON_PROJECTILE_SPEED := " in (tmp_path / "gds/weapons.gd").read_text(encoding="utf-8")


def test_push_does_not_alter_surrounding_content(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    before = (tmp_path / "go/constants.go").read_text(encoding="utf-8")
    assert "// keep before" in before
    assert "// keep after" in before

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    after = (tmp_path / "go/constants.go").read_text(encoding="utf-8")
    assert "// keep before" in after
    assert "// keep after" in after


def test_diff_writes_nothing(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_project(tmp_path)
    before = (tmp_path / "go/constants.go").read_text(encoding="utf-8")

    exit_code = run(["-diff", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 0
    assert "-old" in captured.out
    assert "+const PlayerSpeed = 420.0" in captured.out
    assert (tmp_path / "go/constants.go").read_text(encoding="utf-8") == before


def test_diff_includes_all_constants_outputs_when_stale(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_project(tmp_path)

    exit_code = run(["-diff", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 0
    assert str(tmp_path / "go/constants.go") in captured.out
    assert str(tmp_path / "go/weapons.go") in captured.out
    assert "const PlayerSpeed = 420.0" in captured.out
    assert "BasicCannonProjectileSpeed" in captured.out


def test_check_exits_zero_when_synced(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    assert run(["-check", "-constants", "-go", "-config", str(config_path)]) == 0


def test_check_exits_one_when_out_of_sync(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-check", "-constants", "-go", "-config", str(config_path)]) == 1


def test_check_fails_when_either_go_constants_output_is_stale(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    (tmp_path / "go/constants.go").write_text(
        """
package constants

// keep before
// data-sync:start constants.gameplay
const PlayerSpeed = 420.0
const TickRate = 60
const DebugEnabled = true
const WelcomeText = "hello"
// data-sync:end constants.gameplay
// keep after
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
old
// data-sync:end constants.server.weapons.basic_cannon
// data-sync:start constants.server.weapons.torpedo
old
// data-sync:end constants.server.weapons.torpedo
// data-sync:start constants.shared.weapons.torpedo_radial_shape
old
// data-sync:end constants.shared.weapons.torpedo_radial_shape
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-check", "-constants", "-go", "-config", str(config_path)]) == 1


def test_language_filtering_works(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    ts_before = (tmp_path / "ts/constants.ts").read_text(encoding="utf-8")

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(
        encoding="utf-8"
    )
    assert (tmp_path / "ts/constants.ts").read_text(encoding="utf-8") == ts_before


def test_push_fails_when_destination_marker_section_is_missing_from_toml(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    config_path = write_project(tmp_path)
    missing_path = tmp_path / "go/missing.go"
    missing_path.write_text(
        """
package constants

// data-sync:start constants.missing.example
old
// data-sync:end constants.missing.example
""".lstrip(),
        encoding="utf-8",
    )

    exit_code = run(["-push", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 1
    assert "constants.missing.example" in captured.err
