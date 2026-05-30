from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.constants_store import ConstantsStore, ConstantsStoreError


pytest.importorskip("tomlkit")


def write_toml(path: Path, body: str) -> Path:
    path.write_text(body.strip() + "\n", encoding="utf-8")
    return path


def test_load_single_file_reads_constants_section(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "game_data.toml",
        """
[constants.gameplay]
player_speed = 420.0
tick_rate = 60
""",
    )

    store = ConstantsStore.load([path])
    section = store.constants("constants.gameplay")

    assert section.name == "constants.gameplay"
    assert section.values == (
        ("player_speed", 420.0),
        ("tick_rate", 60),
    )


def test_load_multiple_files_merges_same_section_namespace(tmp_path: Path) -> None:
    base_path = write_toml(
        tmp_path / "game_data.base.toml",
        """
[constants.gameplay]
player_speed = 420.0
""",
    )
    override_path = write_toml(
        tmp_path / "game_data.override.toml",
        """
[constants.gameplay]
tick_rate = 60
""",
    )

    store = ConstantsStore.load([base_path, override_path])
    section = store.constants("constants.gameplay")

    assert section.values == (
        ("player_speed", 420.0),
        ("tick_rate", 60),
    )


def test_load_multiple_files_raises_on_duplicate_section_key(tmp_path: Path) -> None:
    base_path = write_toml(
        tmp_path / "game_data.base.toml",
        """
[constants.gameplay]
tick_rate = 60
""",
    )
    override_path = write_toml(
        tmp_path / "game_data.override.toml",
        """
[constants.gameplay]
tick_rate = 120
""",
    )

    store = ConstantsStore.load([base_path, override_path])

    with pytest.raises(
        ConstantsStoreError, match=r"duplicate constants key.*constants\.gameplay.*tick_rate"
    ) as excinfo:
        store.constants("constants.gameplay")

    error_text = str(excinfo.value)
    assert str(base_path) in error_text
    assert str(override_path) in error_text
