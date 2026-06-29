package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

// GameplayPresentationSnapshot is the game-facing DTO for realtime presentation projection.
type GameplayPresentationSnapshot struct {
	SelfID          string
	Lives           int
	Players         map[string]runtime.ShipState
	PlayerSessions  map[string]PlayerSessionState
	PlayerLifecycle map[string]string
	Bullets         map[string]runtime.BulletState
	Asteroids       map[string]runtime.AsteroidState
	Pickups         map[string]runtime.PickupState
	TotalAsteroids  int
	PendingEvents   []PendingPresentationEvent
	ServerSentMsec  int
}
// GameplayPresentationSnapshot returns a non-draining copy of the authoritative
// presentation state for realtime projection.
func (game *Game) GameplayPresentationSnapshot(playerID string) GameplayPresentationSnapshot {
	game.mu.Lock()
	defer game.mu.Unlock()

	players := make(map[string]runtime.ShipState, len(game.entities.Players))
	for id, player := range game.entities.Players {
		players[id] = player.State()
	}

	matchDecision := game.matchDecisionLocked()
	playerLifecycle := make(map[string]string, len(matchDecision.Players))
	for _, player := range matchDecision.Players {
		playerLifecycle[player.ID] = string(player.Status)
	}

	playerSessions := game.playerSessionStatesLocked()

	asteroids := make(map[string]runtime.AsteroidState, len(game.entities.Asteroids))
	for id, asteroid := range game.entities.Asteroids {
		asteroids[id] = asteroid.State()
	}

	pickups := game.pickupStatesLocked()
	bullets := make(map[string]runtime.BulletState, len(game.entities.Projectiles))
	for id, bullet := range game.entities.Projectiles {
		bullets[id] = bullet.State()
	}

	pending := game.pendingPresentationEvents[playerID]
	pendingEvents := make([]PendingPresentationEvent, len(pending))
	copy(pendingEvents, pending)

	return GameplayPresentationSnapshot{
		SelfID:          playerID,
		Lives:           game.playerLives(playerID),
		Players:         players,
		PlayerSessions:  playerSessions,
		PlayerLifecycle: playerLifecycle,
		Bullets:         bullets,
		Asteroids:       asteroids,
		Pickups:         pickups,
		TotalAsteroids:  game.spawner.TotalAsteroidsSpawned(),
		PendingEvents:   pendingEvents,
		ServerSentMsec:  0,
	}
}

