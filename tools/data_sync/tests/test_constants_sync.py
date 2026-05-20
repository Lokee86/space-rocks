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
    (tmp_path / "gds/constants.gd").write_text(
        """
extends RefCounted

# data-sync:start constants.client
old
# data-sync:end constants.client
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


def test_push_updates_only_managed_block(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    exit_code = run(["-push", "-constants", "-go", "-config", str(config_path)])

    assert exit_code == 0
    assert (tmp_path / "go/constants.go").read_text(encoding="utf-8") == """
package constants

// keep before
// data-sync:start constants.gameplay
const PlayerSpeed = 420.0
const TickRate = 60
const DebugEnabled = true
const WelcomeText = "hello"
// data-sync:end constants.gameplay
// keep after
""".lstrip()


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


def test_check_exits_zero_when_synced(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    assert run(["-check", "-constants", "-go", "-config", str(config_path)]) == 0


def test_check_exits_one_when_out_of_sync(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-check", "-constants", "-go", "-config", str(config_path)]) == 1


def test_language_filtering_works(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    ts_before = (tmp_path / "ts/constants.ts").read_text(encoding="utf-8")

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(
        encoding="utf-8"
    )
    assert (tmp_path / "ts/constants.ts").read_text(encoding="utf-8") == ts_before
