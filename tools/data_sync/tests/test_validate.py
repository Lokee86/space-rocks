from __future__ import annotations

from pathlib import Path

import pytest

from main import run


pytest.importorskip("tomlkit")


def write_validation_project(tmp_path: Path) -> Path:
    for directory in ["shared", "shared/player_data", "go", "gds", "ts"]:
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
    (tmp_path / "go/weapons.go").write_text(
        block("//", "constants.server.weapons.basic_cannon")
        + block("//", "constants.server.weapons.torpedo")
        + block("//", "constants.shared.weapons.torpedo_radial_shape"),
        encoding="utf-8",
    )
    (tmp_path / "gds/constants.gd").write_text(
        block("#", "constants.gameplay") + block("#", "constants.client"),
        encoding="utf-8",
    )
    (tmp_path / "gds/weapons.gd").write_text(
        block("#", "constants.server.weapons.basic_cannon"),
        encoding="utf-8",
    )
    (tmp_path / "ts/constants.ts").write_text(
        block("//", "constants.network") + block("//", "constants.client"),
        encoding="utf-8",
    )
    (tmp_path / "go/packets.go").write_text(block("//", "packets"), encoding="utf-8")
    (tmp_path / "gds/packets.gd").write_text(block("#", "packets"), encoding="utf-8")
    (tmp_path / "ts/packets.ts").write_text(block("//", "packets"), encoding="utf-8")

    (tmp_path / "shared/player_data/stats.toml").write_text(
        """
schema_name = "stats"
schema_version = "v1.1"

[fields.total_score]
type = "integer"
default = 0

[fields.high_score]
type = "integer"
default = 0

[fields.ship_deaths]
type = "integer"
default = 0

[fields.games_played]
type = "integer"
default = 0

[fields.wins]
type = "integer"
default = 0
scope = "multiplayer_only"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/player_data/match_result.toml").write_text(
        """
schema_name = "match_result"
schema_version = "v1.1"

[MatchResultSummary.metadata]
winner_rule = "multiplayer_highest_score"
ties_award_no_wins = true

[MatchResultSummary.fields.match_id]
type = "string"
required = true

[MatchResultSummary.fields.mode]
type = "string"
required = true

[MatchResultSummary.fields.resolved_at]
type = "string"
optional = true

[PlayerMatchSummary.fields.game_player_id]
type = "string"
required = true

[PlayerMatchSummary.fields.account_user_id]
type = "integer"
optional = true

[PlayerMatchSummary.fields.local_profile_id]
type = "string"
optional = true

[PlayerMatchSummary.fields.score]
type = "integer"
default = 0

[PlayerMatchSummary.fields.ship_deaths]
type = "integer"
default = 0

[PlayerMatchSummary.fields.won]
type = "boolean"
default = false
""".strip()
        + "\n",
        encoding="utf-8",
    )

    config_path = tmp_path / "config.toml"
    config_path.write_text(valid_config_text(), encoding="utf-8")
    return config_path


def valid_config_text() -> str:
    return """
[sot]
path = "shared/game_data.toml"

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd", "ts/**/*.ts"]
exclude = []

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


def player_data_config_text() -> str:
    return """
[sot.constants]
paths = ["shared/game_data.toml"]

[sot.packets]
paths = ["shared/game_data.toml"]

[sot.player_data]
paths = ["shared/player_data/stats.toml", "shared/player_data/match_result.toml"]

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd", "ts/**/*.ts"]
exclude = []

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


def block(comment: str, section: str) -> str:
    return f"{comment} data-sync:start {section}\nold\n{comment} data-sync:end {section}\n"


def test_validate_valid_config_and_sot(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    assert run(["-validate", "-config", str(config_path)]) == 0


def test_validate_player_data_sot(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    config_path.write_text(player_data_config_text(), encoding="utf-8")

    assert run(["-validate", "-player_data", "-config", str(config_path)]) == 0


def test_validate_player_data_rejects_invalid_toml(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "shared/player_data/stats.toml").write_text(
        """
schema_name = "stats"
schema_version = "v1.1"

[fields.total_score]
type = "float"
default = 0
""".strip()
        + "\n",
        encoding="utf-8",
    )
    config_path.write_text(player_data_config_text(), encoding="utf-8")

    assert run(["-validate", "-player_data", "-config", str(config_path)]) == 1


def test_validate_constants_multiple_sot_files(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "shared/constants_a.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

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
    (tmp_path / "shared/constants_b.toml").write_text(
        """
[constants.client]
client_scale = 2
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/packets.toml").write_text(
        """
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
""".strip()
        + "\n",
        encoding="utf-8",
    )
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/constants_a.toml", "shared/constants_b.toml"]

[sot.packets]
path = "shared/packets.toml"

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd", "ts/**/*.ts"]
exclude = []

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

    assert run(["-validate", "-config", str(config_path)]) == 0


def test_validate_constants_multiple_sot_files_duplicate_section_key(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "shared/constants_a.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.network]
max_players = 2
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/constants_b.toml").write_text(
        """
[constants.gameplay]
player_speed = 420.0

[constants.client]
client_scale = 2
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (tmp_path / "shared/packets.toml").write_text(
        """
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
""".strip()
        + "\n",
        encoding="utf-8",
    )
    config_path.write_text(
        """
[sot.constants]
paths = ["shared/constants_a.toml", "shared/constants_b.toml"]

[sot.packets]
path = "shared/packets.toml"

[constants.scan]
include = ["go/**/*.go", "gds/**/*.gd", "ts/**/*.ts"]
exclude = []

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

    assert run(["-validate", "-config", str(config_path)]) == 1


def test_validate_files_without_data_sync_markers_are_ignored_for_constants(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "go/ignored.go").write_text(
        """
package constants

const Untouched = true
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 0


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
    (tmp_path / "go/weapons.go").write_text(
        block("//", "constants.server.weapons.basic_cannon"),
        encoding="utf-8",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 1


def test_validate_missing_source_section_for_constants_output(tmp_path: Path) -> None:
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
""",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 1


def test_validate_discovered_constants_block_missing_toml_section_fails_clearly(
    tmp_path: Path,
    capsys: pytest.CaptureFixture[str],
) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "go/missing.go").write_text(
        """
package constants

// data-sync:start constants.missing.example
old
// data-sync:end constants.missing.example
""".lstrip(),
        encoding="utf-8",
    )

    exit_code = run(["-validate", "-constants", "-go", "-config", str(config_path)])

    captured = capsys.readouterr()
    assert exit_code == 1
    assert "constants.missing.example" in captured.err


def test_validate_arbitrary_constants_output_is_checked(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    (tmp_path / "go/weapons.go").write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
const BasicCannonProjectileSpeed = 1200.0
// data-sync:end constants.server.weapons.basic_cannon
// data-sync:start constants.server.weapons.torpedo
const TorpedoProjectileSpeed = 1200.0
// data-sync:end constants.server.weapons.torpedo
// data-sync:start constants.shared.weapons.torpedo_radial_shape
const TorpedoRadialZoneCount = 4
// data-sync:end constants.shared.weapons.torpedo_radial_shape
""".lstrip(),
        encoding="utf-8",
    )

    assert run(["-validate", "-constants", "-go", "-config", str(config_path)]) == 0


def test_validate_missing_configured_file(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)
    (tmp_path / "ts/packets.ts").unlink()

    assert run(["-validate", "-packets", "-ts", "-config", str(config_path)]) == 1
