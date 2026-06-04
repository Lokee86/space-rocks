package game

import "github.com/Lokee86/space-rocks/server/internal/game/runtime"

func (game *Game) StatePacket(playerID string) StatePacket {
	game.mu.Lock()
	defer game.mu.Unlock()

	response := game.statePacket(playerID)
	game.pendingPresentationEvents[playerID] = nil

	return response
}

func (game *Game) statePacket(playerID string) StatePacket {
	players := make(map[string]runtime.ShipState, len(game.state.Players))
	for id, player := range game.state.Players {
		players[id] = player.State()
	}
	matchDecision := game.matchDecisionLocked()
	playerLifecycle := make(map[string]string, len(matchDecision.Players))
	for _, player := range matchDecision.Players {
		playerLifecycle[player.ID] = string(player.Status)
	}
	playerWorldStates := game.playerWorldStatesLocked()

	asteroids := make(map[string]runtime.AsteroidState, len(game.state.Asteroids))
	for id, asteroid := range game.state.Asteroids {
		asteroids[id] = asteroid.State()
	}

	bullets := make(map[string]runtime.BulletState, len(game.state.Projectiles))
	for id, bullet := range game.state.Projectiles {
		bullets[id] = bullet.State()
	}
	events := append(make([]EventState, 0, len(game.pendingPresentationEvents[playerID])), game.pendingPresentationEvents[playerID]...)

	return StatePacket{
		Type:            PacketTypeState,
		SelfID:          playerID,
		Lives:           game.playerLives(playerID),
		Players:         players,
		PlayerLifecycle: playerLifecycle,
		PlayerWorldStates: playerWorldStates,
		Bullets:         bullets,
		Asteroids:       asteroids,
		TotalAsteroids:  game.spawner.TotalAsteroidsSpawned(),
		Events:          events,
	}
}
