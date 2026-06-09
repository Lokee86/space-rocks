package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestStoreRouterLoadStats(t *testing.T) {
	accountStore := NewMemoryStore()
	localStore := NewMemoryStore()
	guestStore := NewNoopStore()
	router := NewStoreRouter(accountStore, localStore, guestStore)

	t.Run("account route", func(t *testing.T) {
		identity := protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		}
		if _, _, err := accountStore.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID:   "account-result",
			MatchID:    "match-1",
			Identity:   identity,
			Score:      5,
			ShipDeaths: 1,
			Won:        true,
		}); err != nil {
			t.Fatalf("seed account store: %v", err)
		}

		stats, found, err := router.LoadStats(identity)
		if err != nil {
			t.Fatalf("LoadStats returned error: %v", err)
		}
		if !found {
			t.Fatal("LoadStats returned found=false for account identity")
		}
		if stats.TotalScore != 5 || stats.Wins != 1 {
			t.Fatalf("LoadStats returned %+v, want account stats", stats)
		}
	})

	t.Run("local route", func(t *testing.T) {
		identity := protocol.PlayerDataIdentity{
			IdentityKind:   IdentityKindLocalProfile,
			LocalProfileID: "local-123",
		}
		if _, _, err := localStore.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID:   "local-result",
			MatchID:    "match-2",
			Identity:   identity,
			Score:      7,
			ShipDeaths: 2,
			Won:        false,
		}); err != nil {
			t.Fatalf("seed local store: %v", err)
		}

		stats, found, err := router.LoadStats(identity)
		if err != nil {
			t.Fatalf("LoadStats returned error: %v", err)
		}
		if !found {
			t.Fatal("LoadStats returned found=false for local identity")
		}
		if stats.TotalScore != 7 || stats.ShipDeaths != 2 {
			t.Fatalf("LoadStats returned %+v, want local stats", stats)
		}
	})

	t.Run("guest route", func(t *testing.T) {
		stats, found, err := router.LoadStats(protocol.PlayerDataIdentity{
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
	})

	t.Run("unknown route", func(t *testing.T) {
		if _, _, err := router.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: "unknown",
		}); err == nil {
			t.Fatal("LoadStats returned nil error for unknown identity_kind")
		}
	})
}

func TestStoreRouterRecordMatchResult(t *testing.T) {
	accountStore := NewMemoryStore()
	localStore := NewMemoryStore()
	guestStore := NewNoopStore()
	router := NewStoreRouter(accountStore, localStore, guestStore)

	t.Run("account route", func(t *testing.T) {
		stats, duplicate, err := router.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "account-result",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
			Score:      5,
			ShipDeaths: 1,
			Won:        true,
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("RecordMatchResult returned duplicate=true for account route")
		}
		if stats.TotalScore != 5 || stats.Wins != 1 {
			t.Fatalf("RecordMatchResult returned %+v, want account stats", stats)
		}
	})

	t.Run("local route", func(t *testing.T) {
		stats, duplicate, err := router.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "local-result",
			MatchID:  "match-2",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   IdentityKindLocalProfile,
				LocalProfileID: "local-123",
			},
			Score:      7,
			ShipDeaths: 2,
			Won:        false,
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("RecordMatchResult returned duplicate=true for local route")
		}
		if stats.TotalScore != 7 || stats.ShipDeaths != 2 {
			t.Fatalf("RecordMatchResult returned %+v, want local stats", stats)
		}
	})

	t.Run("guest route", func(t *testing.T) {
		stats, duplicate, err := router.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "guest-result",
			MatchID:  "match-3",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindGuest,
			},
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("RecordMatchResult returned duplicate=true for guest route")
		}
		if stats != (protocol.PlayerDataStats{}) {
			t.Fatalf("RecordMatchResult returned %+v, want zero stats", stats)
		}
	})

	t.Run("unknown route", func(t *testing.T) {
		if _, _, err := router.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: "unknown",
			},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for unknown identity_kind")
		}
	})
}
