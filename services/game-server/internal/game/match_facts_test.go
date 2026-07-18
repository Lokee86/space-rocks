package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
)

func TestPlayerMatchFactsIncludesNewPlayer(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	facts := game.PlayerMatchFacts()

	if len(facts) != 1 || facts[0] != (PlayerMatchFact{GamePlayerID: playerID}) {
		t.Fatalf("facts = %+v, want new player with zero score and deaths", facts)
	}
}

func TestMatchDeathFactsPreserveNormalizedAttributionAndRemovedPlayers(t *testing.T) {
	game := New()
	game.SetMatchContext("match-1", "trace-1")
	game.SetModeContext("mode-1")
	playerID := game.AddPlayer()
	game.lifeRuntime.ApplyDeath(lives.DeathInput{
		PlayerID:        playerID,
		DestroyedShipID: "ship-1",
		MatchID:         "match-1",
		ModeID:          game.modeID,
		CauseCode:       "projectile",
		Attribution:     lives.AttributionPlayerCaused,
		KillerPlayerID:  "player-2",
	})
	game.RemovePlayer(playerID)
	facts := game.MatchDeathFacts()
	if len(facts) != 1 || facts[0].Input.Attribution != lives.AttributionPlayerCaused || facts[0].Input.ModeID != "mode-1" {
		t.Fatalf("unexpected match death facts: %+v", facts)
	}
	facts[0].Input.KillerPlayerID = "mutated"
	if game.MatchDeathFacts()[0].Input.KillerPlayerID != "player-2" {
		t.Fatal("match death facts were not defensively copied")
	}
}

func TestPlayerMatchFactsReturnsOneFactWithScoreAndShipDeaths(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.AddPlayerScore(playerID, 150)
	for death := 0; death < 3; death++ {
		game.applyFatalPlayerDamage(playerID, game.entities.Players[playerID])
		if death < 2 {
			game.lifeRuntime.Step(game.lifeRuntime.Policy().RespawnDelay)
			game.lifeRuntime.CommitRespawn(playerID)
			session := game.playerSessions[playerID]
			game.entities.Players[playerID] = session.NewShip(session.SpawnPosition)
		}
	}

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
	game.applyFatalPlayerDamage(playerID1, game.entities.Players[playerID1])
	for death := 0; death < 2; death++ {
		game.applyFatalPlayerDamage(playerID2, game.entities.Players[playerID2])
		if death == 0 {
			game.lifeRuntime.Step(game.lifeRuntime.Policy().RespawnDelay)
			game.lifeRuntime.CommitRespawn(playerID2)
			session := game.playerSessions[playerID2]
			game.entities.Players[playerID2] = session.NewShip(session.SpawnPosition)
		}
	}

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

func TestRemovePlayerPreservesFactsAndStopsActiveParticipation(t *testing.T) {
	game := New()
	removedPlayerID := game.AddPlayer()
	remainingPlayerID := game.AddPlayer()

	game.AddPlayerScore(removedPlayerID, 150)
	game.applyFatalPlayerDamage(removedPlayerID, game.entities.Players[removedPlayerID])
	game.AddPlayerScore(remainingPlayerID, 75)

	game.SetPlayerLives(remainingPlayerID, 0)
	game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: remainingPlayerID})
	delete(game.entities.Players, remainingPlayerID)

	game.RemovePlayer(removedPlayerID)
	game.RemovePlayer(removedPlayerID)

	if _, ok := game.playerSessions[removedPlayerID]; ok {
		t.Fatalf("player session %q still exists after removal", removedPlayerID)
	}
	if _, ok := game.entities.Players[removedPlayerID]; ok {
		t.Fatalf("active ship %q still exists after removal", removedPlayerID)
	}

	snapshot := game.matchSnapshot()
	if len(snapshot.Players) != 1 || snapshot.Players[0].ID != remainingPlayerID {
		t.Fatalf("match snapshot = %+v, want only %q", snapshot.Players, remainingPlayerID)
	}
	if decision := game.MatchDecision(); !decision.IsOver {
		t.Fatalf("match decision = %+v, want removed player not to block completion", decision)
	}

	factsByID := playerMatchFactsByID(game.PlayerMatchFacts())
	if len(factsByID) != 2 {
		t.Fatalf("facts = %+v, want exactly two historical participants", factsByID)
	}
	if fact := factsByID[removedPlayerID]; fact.Score != 150 || fact.ShipDeaths != 1 {
		t.Fatalf("removed player fact = %+v, want score 150 and one death", fact)
	}
	if fact := factsByID[remainingPlayerID]; fact.Score != 75 || fact.ShipDeaths != 1 {
		t.Fatalf("remaining player fact = %+v, want score 75 and one death", fact)
	}
}

func TestRemoveLastPlayerCompletesMatchAndRepeatedRemovalIsStable(t *testing.T) {
	game := New()
	if decision := game.MatchDecision(); decision.IsOver {
		t.Fatalf("new empty game decision = %+v, want not over", decision)
	}

	playerID := game.AddPlayer()
	game.AddPlayerScore(playerID, 90)
	game.RemovePlayer(playerID)

	if _, ok := game.playerSessions[playerID]; ok {
		t.Fatalf("player session %q still exists after removal", playerID)
	}
	if _, ok := game.entities.Players[playerID]; ok {
		t.Fatalf("active ship %q still exists after removal", playerID)
	}
	if snapshot := game.matchSnapshot(); len(snapshot.Players) != 0 || !snapshot.HadParticipants {
		t.Fatalf("match snapshot = %+v, want no active players with historical participation", snapshot)
	}
	if decision := game.MatchDecision(); !decision.IsOver || len(decision.Players) != 0 {
		t.Fatalf("match decision = %+v, want completed match with no active players", decision)
	}

	facts := game.PlayerMatchFacts()
	if len(facts) != 1 || facts[0] != (PlayerMatchFact{GamePlayerID: playerID, Score: 90}) {
		t.Fatalf("facts = %+v, want removed player's accumulated facts", facts)
	}

	game.RemovePlayer(playerID)

	if decision := game.MatchDecision(); !decision.IsOver || len(decision.Players) != 0 {
		t.Fatalf("decision after repeated removal = %+v, want stable completed match", decision)
	}
	repeatedFacts := game.PlayerMatchFacts()
	if len(repeatedFacts) != 1 || repeatedFacts[0] != facts[0] {
		t.Fatalf("facts after repeated removal = %+v, want unchanged %+v", repeatedFacts, facts)
	}
}

func TestDiscardPlayerRemovesActiveAndHistoricalParticipation(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.AddPlayerScore(playerID, 90)

	game.DiscardPlayer(playerID)
	game.DiscardPlayer(playerID)

	if _, ok := game.playerSessions[playerID]; ok {
		t.Fatalf("player session %q still exists after discard", playerID)
	}
	if _, ok := game.entities.Players[playerID]; ok {
		t.Fatalf("active ship %q still exists after discard", playerID)
	}
	if facts := game.PlayerMatchFacts(); len(facts) != 0 {
		t.Fatalf("facts = %+v, want abandoned participation discarded", facts)
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

func playerMatchFactsByID(facts []PlayerMatchFact) map[string]PlayerMatchFact {
	factsByID := make(map[string]PlayerMatchFact, len(facts))
	for _, fact := range facts {
		factsByID[fact.GamePlayerID] = fact
	}
	return factsByID
}
