from __future__ import annotations

from pathlib import Path

import pytest

from main import run
from tests.test_packets_sync import write_project
from tests.test_validate import write_validation_project


pytest.importorskip("tomlkit")


def test_full_push_constants_and_packets(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    from tests.test_packets_sync import GO_PACKETS

    assert run(["-push", "-constants", "-packets", "-go", "-config", str(config_path)]) == 0

    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(
        encoding="utf-8"
    )
    assert (tmp_path / "go/packets.go").read_text(encoding="utf-8") == GO_PACKETS


def test_push_constants_does_not_touch_drop_tables(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    drop_tables_before = (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8")

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0

    assert (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8") == drop_tables_before


def test_push_drop_tables_go_writes_only_drop_tables_output(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-drop-tables", "-go", "-config", str(config_path)]) == 0

    assert "GeneratedTables" in (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8")
    assert (tmp_path / "go/packets.go").read_text(encoding="utf-8") == "stale go packets\n"
    assert (tmp_path / "gds/packets.gd").read_text(encoding="utf-8") == "stale gds packets\n"


def test_push_drop_tables_requires_go(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_project(tmp_path)

    with pytest.raises(SystemExit) as exc:
        run(["-push", "-drop-tables", "-config", str(config_path)])
    captured = capsys.readouterr()
    assert exc.value.code == 2
    assert "-push requires at least one language" in captured.err


def test_push_drop_tables_rejects_gds(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    config_path = write_project(tmp_path)

    with pytest.raises(SystemExit) as exc:
        run(["-push", "-drop-tables", "-gds", "-config", str(config_path)])
    captured = capsys.readouterr()
    assert exc.value.code == 2
    assert "-push with -drop-tables requires -go" in captured.err
    assert (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8") == "stale drop tables\n"


def test_full_check(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    from tests.test_packets_sync import GDS_PACKETS, GO_PACKETS

    assert run(["-push", "-constants", "-packets", "-go", "-gds", "-config", str(config_path)]) == 0

    assert run(["-check", "-constants", "-packets", "-go", "-gds", "-config", str(config_path)]) == 0


def test_validate_final_flow(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    assert run(["-validate", "-config", str(config_path)]) == 0
