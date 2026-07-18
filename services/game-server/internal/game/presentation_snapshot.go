package game

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

// gameplayPresentationFrame is immutable after publication and may be read after game.mu is released.
type gameplayPresentationFrame struct {
	players         map[string]runtime.ShipState
	playerSessions  map[string]PlayerSessionState
	playerLifecycle map[string]string
	playerLocators  map[string]PlayerLocatorState
	bullets         map[string]runtime.BulletState
	asteroids       map[string]runtime.AsteroidState
	pickups         map[string]runtime.PickupState
	totalAsteroids  int
	serverSentMsec  int
	generation      uint64
}

func (game *Game) publishPresentationFrameLocked() {
	players := make(map[string]runtime.ShipState, len(game.entities.Players))
	for id, player := range game.entities.Players {
		state := player.State()
		if session := game.playerSessions[id]; session != nil {
			state.TeamID = string(session.TeamID)
		}
		players[id] = state
	}

	playerLocators := make(map[string]PlayerLocatorState, len(game.playerSessions))
	for id, session := range game.playerSessions {
		locator := PlayerLocatorState{ID: id, X: session.SpawnPosition.X, Y: session.SpawnPosition.Y}
		if ship := game.entities.Players[id]; ship != nil && !ship.IsPendingDespawn() {
			locator.X = ship.X
			locator.Y = ship.Y
			locator.VelocityX = ship.Velocity.X
			locator.VelocityY = ship.Velocity.Y
			locator.Active = true
		} else if view := game.cameraViews[id]; view != nil {
			locator.X = view.X
			locator.Y = view.Y
		}
		playerLocators[id] = locator
	}

	matchDecision := game.matchDecisionLocked()
	playerLifecycle := make(map[string]string, len(matchDecision.Players))
	for _, player := range matchDecision.Players {
		playerLifecycle[player.ID] = string(player.Status)
	}

	asteroids := make(map[string]runtime.AsteroidState, len(game.entities.Asteroids))
	for id, asteroid := range game.entities.Asteroids {
		asteroids[id] = asteroid.State()
	}

	bullets := make(map[string]runtime.BulletState, len(game.entities.Projectiles))
	for id, bullet := range game.entities.Projectiles {
		bullets[id] = bullet.State()
	}

	generation := uint64(1)
	if game.presentationFrame != nil {
		generation = game.presentationFrame.generation + 1
	}

	game.presentationFrame = &gameplayPresentationFrame{
		players:         players,
		playerSessions:  game.playerSessionStatesLocked(),
		playerLifecycle: playerLifecycle,
		playerLocators:  playerLocators,
		bullets:         bullets,
		asteroids:       asteroids,
		pickups:         game.pickupStatesLocked(),
		totalAsteroids:  game.spawner.TotalAsteroidsSpawned(),
		serverSentMsec:  int(time.Now().UnixMilli()),
		generation:      generation,
	}
}

// GameplayPresentationSnapshot is the game-facing DTO for realtime presentation projection.
type GameplayPresentationSnapshot struct {
	SelfID          string
	Lives           int
	Players         map[string]runtime.ShipState
	PlayerSessions  map[string]PlayerSessionState
	PlayerLifecycle map[string]string
	PlayerLocators  map[string]PlayerLocatorState
	CameraView      runtime.CameraView
	HasCameraView   bool
	Bullets         map[string]runtime.BulletState
	Asteroids       map[string]runtime.AsteroidState
	Pickups         map[string]runtime.PickupState
	TotalAsteroids  int
	PendingEvents   []PendingPresentationEvent
	ServerSentMsec  int
	Generation      uint64
}

type GameplayPresentationSnapshotDurations struct {
	SnapshotCapture  time.Duration
	PendingEventCopy time.Duration
}

// GameplayPresentationSnapshot returns a receiver-scoped view over the current
// immutable frame and copies only the receiver's pending events.
func (game *Game) GameplayPresentationSnapshot(playerID string) GameplayPresentationSnapshot {
	snapshot, _ := game.GameplayPresentationSnapshotMeasured(playerID)
	return snapshot
}

// GameplayPresentationSnapshotMeasured returns the same receiver-scoped view
// with exclusive snapshot-capture and pending-event-copy durations.
func (game *Game) GameplayPresentationSnapshotMeasured(playerID string) (GameplayPresentationSnapshot, GameplayPresentationSnapshotDurations) {
	snapshotStarted := time.Now()
	game.mu.Lock()
	frame := game.presentationFrame
	if frame == nil {
		game.publishPresentationFrameLocked()
		frame = game.presentationFrame
	}

	pendingCopyStarted := time.Now()
	pending := game.pendingPresentationEvents[playerID]
	pendingEvents := make([]PendingPresentationEvent, len(pending))
	copy(pendingEvents, pending)
	pendingEventCopyDuration := time.Since(pendingCopyStarted)
	cameraView := runtime.CameraView{}
	hasCameraView := false
	if view := game.cameraViews[playerID]; view != nil {
		cameraView = *view
		hasCameraView = true
	}
	game.mu.Unlock()

	lives := 0
	if playerSessionState, ok := frame.playerSessions[playerID]; ok {
		lives = playerSessionState.Lives
	}

	snapshot := GameplayPresentationSnapshot{
		SelfID:          playerID,
		Lives:           lives,
		Players:         frame.players,
		PlayerSessions:  frame.playerSessions,
		PlayerLifecycle: frame.playerLifecycle,
		PlayerLocators:  frame.playerLocators,
		CameraView:      cameraView,
		HasCameraView:   hasCameraView,
		Bullets:         frame.bullets,
		Asteroids:       frame.asteroids,
		Pickups:         frame.pickups,
		TotalAsteroids:  frame.totalAsteroids,
		PendingEvents:   pendingEvents,
		ServerSentMsec:  frame.serverSentMsec,
		Generation:      frame.generation,
	}
	snapshotCaptureDuration := time.Since(snapshotStarted) - pendingEventCopyDuration
	if snapshotCaptureDuration < 0 {
		snapshotCaptureDuration = 0
	}
	return snapshot, GameplayPresentationSnapshotDurations{
		SnapshotCapture:  snapshotCaptureDuration,
		PendingEventCopy: pendingEventCopyDuration,
	}
}
