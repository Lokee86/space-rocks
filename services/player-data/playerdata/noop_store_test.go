package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestNoopStoreLoadStats(t *testing.T) {
	store := NewNoopStore()

	stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindGuest,
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if found {
		t.Fatal("LoadStats returned found=true for guest identity")
	}
	if stats != (protocol.PlayerDataStats{}) {
		t.Fatalf("LoadStats returned %+v, want zero stats", stats)
	}
}

func TestNoopStoreRecordMatchResult(t *testing.T) {
	store := NewNoopStore()

	stats, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindGuest,
		},
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}
	if duplicate {
		t.Fatal("RecordMatchResult returned duplicate=true for guest identity")
	}
	if stats != (protocol.PlayerDataStats{}) {
		t.Fatalf("RecordMatchResult returned %+v, want zero stats", stats)
	}
}
