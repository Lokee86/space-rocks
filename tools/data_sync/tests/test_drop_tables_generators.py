from __future__ import annotations

from data_sync.generators.go_drop_tables import generate_drop_tables
from data_sync.model.drop_tables import DropTable, DropTableEntry, DropTablesModel


def test_generate_drop_tables_renders_go_source() -> None:
    model = DropTablesModel(
        tables=(
            DropTable(
                id="basicasteroids",
                source_type="asteroid",
                drop_mode="single",
                max_drops_per_source=1,
                max_active_pickups=2,
                entries=(
                    DropTableEntry(
                        pickup_type="1_up",
                        chance=0.05,
                        min_source_size=1,
                        max_source_size=4,
                    ),
                ),
            ),
        )
    )

    rendered = generate_drop_tables(model)

    assert "package drops" in rendered
    assert "GeneratedTables" in rendered
    assert '"basicasteroids"' in rendered
    assert "DropMode: DropMode(\"single\")" in rendered
    assert "MaxDropsPerSource: 1" in rendered
    assert 'PickupType: "1_up"' in rendered
    assert "Chance: 0.05" in rendered
