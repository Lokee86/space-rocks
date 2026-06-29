package realtime

import "testing"

func TestCompareLaneRecordsEmitsCreateForMissingFromPrevious(t *testing.T) {
	delta := CompareLaneRecords(nil, []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 1 || delta.Creates[0].ID != "ship-a" {
		t.Fatalf("expected create for ship-a, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected only a create, got %#v", delta)
	}
}

func TestCompareLaneRecordsEmitsUpdateForChangedRecord(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}
	current := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 11}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 0 {
		t.Fatalf("expected no creates, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 1 || delta.Updates[0].ID != "ship-a" || delta.Updates[0].X != 11 {
		t.Fatalf("expected update for ship-a, got %#v", delta.Updates)
	}
	if len(delta.Deletes) != 0 {
		t.Fatalf("expected no deletes, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordsEmitsNothingForUnchangedRecord(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}
	current := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected no delta for unchanged record, got %#v", delta)
	}
}

func TestCompareLaneRecordsEmitsDeleteForMissingFromCurrent(t *testing.T) {
	previous := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}, {ID: "player-b", ShipType: "v_wing", Score: 8}}
	current := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}}

	delta := CompareLaneRecords(previous, current, func(record SessionPlayerRecord) string { return record.ID }, func(left, right SessionPlayerRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 {
		t.Fatalf("expected only a delete, got %#v", delta)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "player-b" {
		t.Fatalf("expected delete for player-b, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordsTreatsMissingFromDeltaAsUnchanged(t *testing.T) {
	previous := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}, {ID: "player-b", ShipType: "v_wing", Score: 8}}
	current := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}}

	delta := CompareLaneRecords(previous, current, func(record SessionPlayerRecord) string { return record.ID }, func(left, right SessionPlayerRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 || len(delta.Deletes) != 1 || delta.Deletes[0] != "player-b" {
		t.Fatalf("expected missing-from-current to be the only delete, got %#v", delta)
	}
	if delta.Deletes[0] == "player-c" {
		t.Fatal("expected missing delta entity to remain unchanged, not deleted")
	}
}

func TestCompareLaneRecordsOrdersDeterministically(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-c", ShipType: "v_wing"}}
	current := []WorldShipRecord{{ID: "ship-b", ShipType: "v_wing"}, {ID: "ship-a", ShipType: "v_wing"}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 2 || delta.Creates[0].ID != "ship-a" || delta.Creates[1].ID != "ship-b" {
		t.Fatalf("expected creates sorted by ID, got %#v", delta.Creates)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "ship-c" {
		t.Fatalf("expected delete sorted deterministically, got %#v", delta.Deletes)
	}
}
