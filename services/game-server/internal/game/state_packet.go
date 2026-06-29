package game

import "github.com/Lokee86/space-rocks/server/internal/game/runtime"

// PendingPresentationEvent keeps the queued event ID alongside the event payload
// in the game-facing presentation layer.
type PendingPresentationEvent struct {
	EventID string
	Event   EventState
}

func (game *Game) pendingPresentationEventStates(playerID string) []EventState {
	pending := game.pendingPresentationEvents[playerID]
	events := make([]EventState, 0, len(pending))
	for _, pendingEvent := range pending {
		events = append(events, pendingEvent.Event)
	}
	return events
}

// PendingPresentationEvents returns a copy of the queued presentation events without draining them.
func (game *Game) PendingPresentationEvents(playerID string) []PendingPresentationEvent {
	game.mu.Lock()
	defer game.mu.Unlock()

	pending := game.pendingPresentationEvents[playerID]
	if len(pending) == 0 {
		return nil
	}

	events := make([]PendingPresentationEvent, len(pending))
	copy(events, pending)
	return events
}

// DrainPendingPresentationEvents removes only the matching pending presentation events for the player.
func (game *Game) DrainPendingPresentationEvents(playerID string, eventIDs ...string) []PendingPresentationEvent {
	game.mu.Lock()
	defer game.mu.Unlock()

	pending := game.pendingPresentationEvents[playerID]
	if len(pending) == 0 || len(eventIDs) == 0 {
		return nil
	}

	wanted := make(map[string]struct{}, len(eventIDs))
	for _, eventID := range eventIDs {
		wanted[eventID] = struct{}{}
	}

	kept := make([]PendingPresentationEvent, 0, len(pending))
	drained := make([]PendingPresentationEvent, 0, len(pending))
	for _, pendingEvent := range pending {
		if _, ok := wanted[pendingEvent.EventID]; ok {
			drained = append(drained, pendingEvent)
			continue
		}
		kept = append(kept, pendingEvent)
	}

	if len(kept) == 0 {
		game.pendingPresentationEvents[playerID] = nil
	} else {
		game.pendingPresentationEvents[playerID] = kept
	}

	return drained
}

func (game *Game) StatePacket(playerID string) StatePacket {
	game.mu.Lock()
	defer game.mu.Unlock()

	response := game.statePacket(playerID)
	game.pendingPresentationEvents[playerID] = nil

	return response
}

func (game *Game) statePacket(playerID string) StatePacket {
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
	events := game.pendingPresentationEventStates(playerID)

	return StatePacket{
		Type:            PacketTypeState,
		SelfID:          playerID,
		Lives:           game.playerLives(playerID),
		Players:         players,
		PlayerSessions:  playerSessions,
		PlayerLifecycle: playerLifecycle,
		Bullets:         bullets,
		Asteroids:       asteroids,
		Pickups:         pickups,
		TotalAsteroids:  game.spawner.TotalAsteroidsSpawned(),
		Events:          events,
	}
}
