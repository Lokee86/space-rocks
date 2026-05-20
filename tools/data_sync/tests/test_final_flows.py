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


def test_full_check(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)
    from tests.test_packets_sync import GDS_PACKETS, GO_PACKETS

    assert run(["-push", "-constants", "-packets", "-go", "-gds", "-config", str(config_path)]) == 0

    assert run(["-check", "-constants", "-packets", "-go", "-gds", "-config", str(config_path)]) == 0


def test_validate_final_flow(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    assert run(["-validate", "-config", str(config_path)]) == 0

