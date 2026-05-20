from __future__ import annotations

import subprocess
import sys
from pathlib import Path

import pytest

from main import run
from tests.test_packets_sync import write_project
from tests.test_validate import write_validation_project


pytest.importorskip("tomlkit")


def test_full_push_constants_and_packets(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-packets", "-go", "-gds", "-ts", "-config", str(config_path)]) == 0

    assert "const PlayerSpeed = 420.0" in (tmp_path / "go/constants.go").read_text(
        encoding="utf-8"
    )
    assert "const PacketPlayerInput = 100" in (tmp_path / "go/packets.go").read_text(
        encoding="utf-8"
    )
    assert "const PACKET_PLAYER_INPUT := 100" in (tmp_path / "gds/packets.gd").read_text(
        encoding="utf-8"
    )
    assert "export interface PlayerInputPacket" in (tmp_path / "ts/packets.ts").read_text(
        encoding="utf-8"
    )


def test_full_check(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    assert run(["-push", "-constants", "-packets", "-go", "-gds", "-ts", "-config", str(config_path)]) == 0
    assert run(["-check", "-constants", "-packets", "-go", "-gds", "-ts", "-config", str(config_path)]) == 0


def test_validate_final_flow(tmp_path: Path) -> None:
    config_path = write_validation_project(tmp_path)

    assert run(["-validate", "-config", str(config_path)]) == 0


def test_json_migration_script_smoke() -> None:
    script = Path(__file__).resolve().parents[2] / "migrations/json_to_toml.py"

    result = subprocess.run(
        [sys.executable, str(script), "--help"],
        check=False,
        capture_output=True,
        text=True,
    )

    assert result.returncode == 0
    assert "--constants-input" in result.stdout
