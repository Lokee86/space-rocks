from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.config import load_config
from data_sync.drop_tables_sync import plan_drop_tables_updates
from main import run
from tests.test_packets_sync import write_project


pytest.importorskip("tomlkit")


def test_plan_drop_tables_updates_targets_only_generated_go_file(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    updates = plan_drop_tables_updates(load_config(config_path))

    assert len(updates) == 1
    assert updates[0].path == tmp_path / "go/drop_tables.go"


def test_push_drop_tables_go_isolated_from_constants(tmp_path: Path) -> None:
    config_path = write_project(tmp_path)

    before = (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8")

    assert run(["-push", "-constants", "-go", "-config", str(config_path)]) == 0
    assert (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8") == before

    assert run(["-push", "-drop-tables", "-go", "-config", str(config_path)]) == 0
    after = (tmp_path / "go/drop_tables.go").read_text(encoding="utf-8")
    assert after != before
