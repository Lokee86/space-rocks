from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.drop_tables_toml import DropTablesTomlError, load_drop_table, load_drop_tables
from data_sync.model.drop_tables import DropTableEntry, DropTable


pytest.importorskip("tomlkit")


def write_toml(path: Path, body: str) -> Path:
    path.write_text(body.strip() + "\n", encoding="utf-8")
    return path


def test_load_drop_table_parses_basic_asteroids_file(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "basicasteroids.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 1
max_source_size = 4
""",
    )

    drop_table = load_drop_table(path)

    assert drop_table.id == "basicasteroids"
    assert drop_table.source_type == "asteroid"
    assert drop_table.drop_mode == "single"
    assert drop_table.max_drops_per_source == 1
    assert drop_table.max_active_pickups == 2
    assert drop_table.entries == (
        DropTableEntry(pickup_type="1_up", chance=0.05, min_source_size=1, max_source_size=4),
    )


def test_load_drop_table_rejects_invalid_chance(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "invalid_chance.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 1.5
min_source_size = 1
max_source_size = 4
""",
    )

    with pytest.raises(DropTablesTomlError, match=r"entries.*chance.*between 0\.0 and 1\.0"):
        load_drop_table(path)


def test_load_drop_table_rejects_invalid_source_size_range(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "invalid_range.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 4
max_source_size = 1
""",
    )

    with pytest.raises(
        DropTablesTomlError, match=r"entries.*min_source_size.*less than or equal to.*max_source_size"
    ):
        load_drop_table(path)


def test_load_drop_tables_parses_multiple_files(tmp_path: Path) -> None:
    first_path = write_toml(
        tmp_path / "basicasteroids.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 1
max_source_size = 4
""",
    )
    second_path = write_toml(
        tmp_path / "heavies.toml",
        """
[table]
id = "heavies"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 0

[[entries]]
pickup_type = "1_up"
chance = 0.1
min_source_size = 2
max_source_size = 5
""",
    )

    model = load_drop_tables([first_path, second_path])

    assert model.tables == (
        DropTable(
            id="basicasteroids",
            source_type="asteroid",
            drop_mode="single",
            max_drops_per_source=1,
            max_active_pickups=2,
            entries=(DropTableEntry(pickup_type="1_up", chance=0.05, min_source_size=1, max_source_size=4),),
        ),
        DropTable(
            id="heavies",
            source_type="asteroid",
            drop_mode="single",
            max_drops_per_source=1,
            max_active_pickups=0,
            entries=(DropTableEntry(pickup_type="1_up", chance=0.1, min_source_size=2, max_source_size=5),),
        ),
    )


def test_load_drop_tables_rejects_duplicate_ids(tmp_path: Path) -> None:
    first_path = write_toml(
        tmp_path / "first.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 1
max_source_size = 4
""",
    )
    second_path = write_toml(
        tmp_path / "second.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 1
max_active_pickups = 0

[[entries]]
pickup_type = "1_up"
chance = 0.1
min_source_size = 2
max_source_size = 5
""",
    )

    with pytest.raises(DropTablesTomlError, match=r"duplicate drop table id: basicasteroids"):
        load_drop_tables([first_path, second_path])


def test_load_drop_table_rejects_unknown_drop_mode(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "unknown_mode.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "burst"
max_drops_per_source = 1
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 1
max_source_size = 4
""",
    )

    with pytest.raises(DropTablesTomlError, match=r"\[table\]\.drop_mode must be one of: single, multi"):
        load_drop_table(path)


def test_load_drop_table_rejects_invalid_max_drops_per_source(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "invalid_max_drops.toml",
        """
[table]
id = "basicasteroids"
source_type = "asteroid"
drop_mode = "single"
max_drops_per_source = 0
max_active_pickups = 2

[[entries]]
pickup_type = "1_up"
chance = 0.05
min_source_size = 1
max_source_size = 4
""",
    )

    with pytest.raises(
        DropTablesTomlError,
        match=r"\[table\]\.max_drops_per_source must be greater than or equal to 1",
    ):
        load_drop_table(path)
