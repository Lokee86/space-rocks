package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestMemoryStoreLoadStats(t *testing.T) {
	store := NewMemoryStore()

	t.Run("unknown account identity", func(t *testing.T) {
		stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		})
		if err != nil {
			t.Fatalf("LoadStats returned error: %v", err)
		}
		if found {
			t.Fatal("LoadStats returned found=true for unknown identity")
		}
		if stats != (protocol.PlayerDataStats{}) {
			t.Fatalf("LoadStats returned %+v, want zero stats", stats)
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
		}); err == nil {
			t.Fatal("LoadStats returned nil error for invalid identity")
		}
	})
}

func TestMemoryStoreRecordMatchResult(t *testing.T) {
	store := NewMemoryStore()
	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}

	first, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      12,
		ShipDeaths: 2,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}
	if duplicate {
		t.Fatal("RecordMatchResult returned duplicate=true for first result")
	}
	if first.TotalScore != 12 {
		t.Fatalf("TotalScore = %d, want 12", first.TotalScore)
	}
	if first.HighScore != 12 {
		t.Fatalf("HighScore = %d, want 12", first.HighScore)
	}
	if first.ShipDeaths != 2 {
		t.Fatalf("ShipDeaths = %d, want 2", first.ShipDeaths)
	}
	if first.GamesPlayed != 1 {
		t.Fatalf("GamesPlayed = %d, want 1", first.GamesPlayed)
	}
	if first.Wins != 1 {
		t.Fatalf("Wins = %d, want 1", first.Wins)
	}

	second, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "result-2",
		MatchID:    "match-2",
		Identity:   identity,
		Score:      7,
		ShipDeaths: 1,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}
	if duplicate {
		t.Fatal("RecordMatchResult returned duplicate=true for new result")
	}
	if second.TotalScore != 19 {
		t.Fatalf("TotalScore = %d, want 19", second.TotalScore)
	}
	if second.HighScore != 12 {
		t.Fatalf("HighScore = %d, want 12", second.HighScore)
	}
	if second.ShipDeaths != 3 {
		t.Fatalf("ShipDeaths = %d, want 3", second.ShipDeaths)
	}
	if second.GamesPlayed != 2 {
		t.Fatalf("GamesPlayed = %d, want 2", second.GamesPlayed)
	}
	if second.Wins != 1 {
		t.Fatalf("Wins = %d, want 1", second.Wins)
	}

	dup, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "result-2",
		MatchID:    "match-2",
		Identity:   identity,
		Score:      999,
		ShipDeaths: 999,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error on duplicate: %v", err)
	}
	if !duplicate {
		t.Fatal("RecordMatchResult returned duplicate=false for duplicate result")
	}
	if dup != second {
		t.Fatalf("duplicate stats = %+v, want %+v", dup, second)
	}
	if got, _, err := store.LoadStats(identity); err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	} else if got != second {
		t.Fatalf("LoadStats returned %+v, want %+v", got, second)
	}
}
