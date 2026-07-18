package encounterlifecycle

import (
	"reflect"
	"testing"
)

func TestOrderCleanupCandidates(t *testing.T) {
	candidates := []CleanupCandidate{
		{EntityID: "entity-d", NearestPlayerDistance: 100, Priority: 0},
		{EntityID: "entity-c", NearestPlayerDistance: 200, Priority: 9},
		{EntityID: "entity-b", NearestPlayerDistance: 200, Priority: 2},
		{EntityID: "entity-a", NearestPlayerDistance: 200, Priority: 2},
	}
	original := append([]CleanupCandidate(nil), candidates...)

	ordered := OrderCleanupCandidates(candidates)
	wantIDs := []string{"entity-a", "entity-b", "entity-c", "entity-d"}
	for index, wantID := range wantIDs {
		if ordered[index].EntityID != wantID {
			t.Fatalf("candidate %d: got %q, want %q", index, ordered[index].EntityID, wantID)
		}
	}
	if !reflect.DeepEqual(candidates, original) {
		t.Fatal("ordering mutated the input candidates")
	}
}
