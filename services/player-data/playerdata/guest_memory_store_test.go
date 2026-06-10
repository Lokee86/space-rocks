package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestGuestMemoryStoreLoadStats(t *testing.T) {
	store := NewGuestMemoryStore()

	stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindGuest,
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if !found {
		t.Fatal("LoadStats returned found=false for guest identity")
	}
	if stats != (protocol.PlayerDataStats{}) {
		t.Fatalf("LoadStats returned %+v, want zero stats", stats)
	}
}

func TestGuestMemoryStoreRecordMatchResult(t *testing.T) {
	store := NewGuestMemoryStore()

	first, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "guest-result-1",
		MatchID:    "match-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		Score:      5,
		ShipDeaths: 1,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}
	if duplicate {
		t.Fatal("RecordMatchResult returned duplicate=true for first guest result")
	}
	if first.TotalScore != 5 || first.HighScore != 5 || first.ShipDeaths != 1 || first.GamesPlayed != 1 || first.Wins != 1 {
		t.Fatalf("RecordMatchResult returned %+v, want first guest stats", first)
	}

	second, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "guest-result-1",
		MatchID:    "match-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		Score:      99,
		ShipDeaths: 99,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error on duplicate: %v", err)
	}
	if !duplicate {
		t.Fatal("RecordMatchResult returned duplicate=false for duplicate guest result")
	}
	if second != first {
		t.Fatalf("duplicate stats = %+v, want %+v", second, first)
	}
}
