package drops

import "testing"

func TestGeneratedTablesAreWellFormed(t *testing.T) {
	for tableID, table := range GeneratedTables.ByID {
		if tableID == "" {
			t.Fatal("expected generated table map key to be non-empty")
		}
		if table.ID == "" {
			t.Fatalf("expected table %q to have a non-empty id", tableID)
		}
		if tableID != table.ID {
			t.Fatalf("expected table key %q to match table id %q", tableID, table.ID)
		}
		if table.SourceType == "" {
			t.Fatalf("expected table %q to have a source type", tableID)
		}
		if table.DropMode != DropModeSingle && table.DropMode != DropModeMulti {
			t.Fatalf("expected table %q to have a valid drop mode, got %q", tableID, table.DropMode)
		}
		if table.MaxDropsPerSource < 1 {
			t.Fatalf("expected table %q to allow at least one drop per source, got %d", tableID, table.MaxDropsPerSource)
		}
		if table.MaxActivePickups < 0 {
			t.Fatalf("expected table %q to have non-negative max active pickups, got %d", tableID, table.MaxActivePickups)
		}
		if len(table.Entries) == 0 {
			t.Fatalf("expected table %q to have at least one entry", tableID)
		}
		for index, entry := range table.Entries {
			if entry.PickupType == "" {
				t.Fatalf("expected table %q entry %d to have a pickup type", tableID, index)
			}
			if entry.Chance < 0.0 || entry.Chance > 1.0 {
				t.Fatalf("expected table %q entry %d chance to be between 0.0 and 1.0, got %v", tableID, index, entry.Chance)
			}
			if entry.MinSourceSize > entry.MaxSourceSize {
				t.Fatalf(
					"expected table %q entry %d source size bounds to be ordered, got %d and %d",
					tableID,
					index,
					entry.MinSourceSize,
					entry.MaxSourceSize,
				)
			}
		}
	}
}
