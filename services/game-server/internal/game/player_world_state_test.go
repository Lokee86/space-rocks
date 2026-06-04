package game

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/server/internal/game/player"
)

func TestPlayerWorldStateLocked_ActivePlayer(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	state, ok := game.playerWorldStateLocked(playerID)
	game.mu.Unlock()

	if !ok {
		t.Fatal("expected player world state to exist")
	}
	if state.Status != playerstate.StatusActive {
		t.Fatalf("expected status %q, got %q", playerstate.StatusActive, state.Status)
	}
	if !state.HasActiveShip {
		t.Fatal("expected HasActiveShip true")
	}
	if !state.Targetable {
		t.Fatal("expected Targetable true")
	}
	if !state.Damageable {
		t.Fatal("expected Damageable true")
	}
	if !state.Collidable {
		t.Fatal("expected Collidable true")
	}
}

func TestPlayerWorldStateLocked_PendingRespawnWithoutActiveShip(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	delete(game.entities.Players, playerID)
	state, ok := game.playerWorldStateLocked(playerID)
	game.mu.Unlock()

	if !ok {
		t.Fatal("expected player world state to exist")
	}
	if state.Status != playerstate.StatusPendingRespawn {
		t.Fatalf("expected status %q, got %q", playerstate.StatusPendingRespawn, state.Status)
	}
	if state.HasActiveShip {
		t.Fatal("expected HasActiveShip false")
	}
	if state.Targetable {
		t.Fatal("expected Targetable false")
	}
	if state.Damageable {
		t.Fatal("expected Damageable false")
	}
	if state.Collidable {
		t.Fatal("expected Collidable false")
	}
}

func TestPlayerWorldStateLocked_EliminatedWithoutActiveShip(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	delete(game.entities.Players, playerID)
	game.playerSessions[playerID].Lives = 0
	state, ok := game.playerWorldStateLocked(playerID)
	game.mu.Unlock()

	if !ok {
		t.Fatal("expected player world state to exist")
	}
	if state.Status != playerstate.StatusEliminated {
		t.Fatalf("expected status %q, got %q", playerstate.StatusEliminated, state.Status)
	}
}
