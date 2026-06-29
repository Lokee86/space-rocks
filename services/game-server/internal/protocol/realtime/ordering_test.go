package realtime

import "testing"

func TestSortOrderedRecordKeysUsesLaneFamilyKindID(t *testing.T) {
	keys := []OrderedRecordKey{
		{Lane: LaneSession, Family: PacketFamilySessionDelta, Kind: "snapshot", ID: "b"},
		{Lane: LaneWorld, Family: PacketFamilyWorldDelta, Kind: "ship", ID: "c"},
		{Lane: LaneWorld, Family: PacketFamilyWorldDelta, Kind: "bullet", ID: "a"},
		{Lane: LaneWorld, Family: PacketFamilyWorldFull, Kind: "ship", ID: "b"},
		{Lane: LaneEvent, Family: PacketFamilyEventBatch, Kind: "event", ID: "a"},
	}

	SortOrderedRecordKeys(keys)

	want := []OrderedRecordKey{
		{Lane: LaneEvent, Family: PacketFamilyEventBatch, Kind: "event", ID: "a"},
		{Lane: LaneSession, Family: PacketFamilySessionDelta, Kind: "snapshot", ID: "b"},
		{Lane: LaneWorld, Family: PacketFamilyWorldDelta, Kind: "bullet", ID: "a"},
		{Lane: LaneWorld, Family: PacketFamilyWorldDelta, Kind: "ship", ID: "c"},
		{Lane: LaneWorld, Family: PacketFamilyWorldFull, Kind: "ship", ID: "b"},
	}

	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("expected key %d to be %#v, got %#v", i, want[i], keys[i])
		}
	}
}
