package game

import (
	"reflect"
	"testing"
)

func TestPlayerMatchFactsReturnsOneFactWithScoreAndShipDeaths(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.AddPlayerScore(playerID, 150)
	session := game.playerSessions[playerID]
	session.ShipDeaths = 3

	facts := game.PlayerMatchFacts()

	if len(facts) != 1 {
		t.Fatalf("len(facts) = %d, want 1", len(facts))
	}
	fact := facts[0]
	if fact.GamePlayerID != playerID {
		t.Fatalf("GamePlayerID = %q, want %q", fact.GamePlayerID, playerID)
	}
	if fact.Score != 150 {
		t.Fatalf("Score = %d, want 150", fact.Score)
	}
	if fact.ShipDeaths != 3 {
		t.Fatalf("ShipDeaths = %d, want 3", fact.ShipDeaths)
	}
}

func TestPlayerMatchFactsReturnsTwoFacts(t *testing.T) {
	game := New()
	playerID1 := game.AddPlayer()
	playerID2 := game.AddPlayer()

	game.AddPlayerScore(playerID1, 100)
	game.AddPlayerScore(playerID2, 250)
	game.playerSessions[playerID1].ShipDeaths = 1
	game.playerSessions[playerID2].ShipDeaths = 2

	facts := game.PlayerMatchFacts()

	if len(facts) != 2 {
		t.Fatalf("len(facts) = %d, want 2", len(facts))
	}

	factsByID := map[string]PlayerMatchFact{}
	for _, fact := range facts {
		factsByID[fact.GamePlayerID] = fact
	}

	fact1, ok := factsByID[playerID1]
	if !ok {
		t.Fatalf("missing fact for %q", playerID1)
	}
	if fact1.Score != 100 || fact1.ShipDeaths != 1 {
		t.Fatalf("fact for %q = %+v, want score 100 shipDeaths 1", playerID1, fact1)
	}

	fact2, ok := factsByID[playerID2]
	if !ok {
		t.Fatalf("missing fact for %q", playerID2)
	}
	if fact2.Score != 250 || fact2.ShipDeaths != 2 {
		t.Fatalf("fact for %q = %+v, want score 250 shipDeaths 2", playerID2, fact2)
	}
}

func TestPlayerMatchFactsHasNoAccountOrLocalIdentityFields(t *testing.T) {
	factType := reflect.TypeOf(PlayerMatchFact{})

	for i := 0; i < factType.NumField(); i++ {
		fieldName := factType.Field(i).Name
		if fieldName == "AccountID" || fieldName == "LocalProfileID" {
			t.Fatalf("unexpected identity field %q on PlayerMatchFact", fieldName)
		}
	}
}
