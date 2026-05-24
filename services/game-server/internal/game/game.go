package game

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/motion"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

type Game struct {
	mu                   sync.Mutex
	stopSimulation       chan struct{}
	stopSimulationOnce   sync.Once
	nextID               int
	spawner              *spawning.Spawner
	asteroidSpawnElapsed float64
	worldDevTools        devtools.WorldOptions
	collisionShapes      physics.CollisionShapeCatalog
	state                entities.GameState
	cameraViews          map[string]*entities.CameraView
	playerSessions       map[string]*playerSession
	pendingEvents        map[string][]EventState
}

func New() *Game {
	collisionShapes, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Game.Warn("collision shapes unavailable", logging.FieldError, err)
	}

	return &Game{
		collisionShapes: collisionShapes,
		stopSimulation:  make(chan struct{}),
		cameraViews:     make(map[string]*entities.CameraView),
		playerSessions:  make(map[string]*playerSession),
		pendingEvents:   make(map[string][]EventState),
		spawner:         spawning.New(),
		state:           entities.NewGameState(),
	}
}

func (game *Game) Start() {
	go game.runSimulation()
}

func (game *Game) Stop() {
	game.stopSimulationOnce.Do(func() {
		close(game.stopSimulation)
	})
}

func (game *Game) IsGameOver() bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	if len(game.playerSessions) == 0 {
		return false
	}
	for _, session := range game.playerSessions {
		if session.Lives > 0 {
			return false
		}
	}
	if len(game.state.Players) > 0 {
		return false
	}

	return true
}

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	spawnPlan := game.planInitialPlayerSpawn(playerIndex, playerID)
	spawnPosition := spawnPlan.Position
	session := newPlayerSession(playerID, spawnPosition)
	player := session.NewShip(spawnPosition)
	game.playerSessions[playerID] = session
	game.state.Players[playerID] = player
	game.cameraViews[playerID] = &entities.CameraView{
		X:      player.X,
		Y:      player.Y,
		Config: player.Config,
	}
	game.pendingEvents[playerID] = nil
	logging.Game.Debug("player added",
		logging.FieldPlayerID, playerID,
		"x", spawnPosition.X,
		"y", spawnPosition.Y,
		"lives", session.Lives,
	)

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
	delete(game.cameraViews, playerID)
	delete(game.playerSessions, playerID)
	delete(game.pendingEvents, playerID)
	logging.Game.Debug("player removed", logging.FieldPlayerID, playerID)
}

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if packet.Type == PacketTypeRespawn {
		game.respawnPlayer(playerID)
		return
	}
	if packet.Type == PacketTypeClientConfig {
		if session, ok := game.playerSessions[playerID]; ok {
			session.Config = packet.Config
		}
		if cameraView, ok := game.cameraViews[playerID]; ok {
			cameraView.SetConfig(packet.Config)
		}
	}

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	switch packet.Type {
	case PacketTypeInput:
		if !player.CanReceiveInput() {
			return
		}
		player.SetInput(packet.Input)
	case PacketTypePausePlayer:
		if player.IsPendingDespawn() {
			return
		}
		player.Pause()
		logging.Game.Debug("player paused", logging.FieldPlayerID, playerID)
	case PacketTypeResumePlayer:
		if player.IsPendingDespawn() {
			logging.Game.Debug("resume ignored; player pending despawn", logging.FieldPlayerID, playerID)
			return
		}
		player.Resume(constants.PlayerResumeInvulnerabilitySeconds)
		logging.Game.Debug("player resumed",
			logging.FieldPlayerID, playerID,
			"invulnerability", constants.PlayerResumeInvulnerabilitySeconds,
		)
	case PacketTypeToggleDebugInvincible:
		enabled := player.DevTools.ToggleInvincible()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug invincibility toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugInfiniteLives:
		enabled := player.DevTools.ToggleInfiniteLives()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug infinite lives toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugFreezeWorld:
		enabled := game.worldDevTools.ToggleFreezeWorld()
		logging.Game.Info("debug world freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugFreezePlayer:
		enabled := player.DevTools.ToggleFreezePlayer()
		player.ClearInput()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug player freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
	}
}

func (game *Game) State(playerID string) []byte {
	game.mu.Lock()
	defer game.mu.Unlock()

	response, err := json.Marshal(game.statePacket(playerID))
	if err != nil {
		logging.Game.Error("state marshal failed", err, logging.FieldPlayerID, playerID)
		return nil
	}
	game.pendingEvents[playerID] = nil

	return response
}

func (game *Game) runSimulation() {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	delta := 1.0 / float64(constants.ServerTickRate)
	for {
		select {
		case <-game.stopSimulation:
			return
		case <-ticker.C:
			game.Step(delta)
		}
	}
}

func (game *Game) Step(delta float64) {
	game.mu.Lock()
	defer game.mu.Unlock()

	bounds := space.DefaultBounds()

	for _, session := range game.playerSessions {
		session.Step(delta)
	}

	for _, player := range game.state.Players {
		motion.AdvanceShip(player, delta, bounds)
		if cameraView, ok := game.cameraViews[player.ID]; ok {
			cameraView.SetPosition(player.Position())
		}
		if player.IsPendingDespawn() {
			continue
		}
		if game.worldDevTools.BulletsCanMove() && player.WantsToShoot() && player.CanShoot() {
			game.spawnBullet(player)
			player.ResetShootCooldown()
		}
	}

	for id, player := range game.state.Players {
		if player.ReadyForRemoval() {
			if session, ok := game.playerSessions[id]; ok {
				session.Score = player.Score
			}
			delete(game.state.Players, id)
		}
	}

	if game.worldDevTools.CanSpawnAsteroids() && game.hasCameraViews() {
		game.asteroidSpawnElapsed += delta
		if game.asteroidSpawnElapsed >= constants.AsteroidSpawnInterval {
			game.asteroidSpawnElapsed = 0
			for _, cameraView := range game.cameraViews {
				game.spawnAsteroidBatch(cameraView)
			}
		}
	} else if !game.hasCameraViews() {
		game.asteroidSpawnElapsed = 0
	}

	for id, asteroid := range game.state.Asteroids {
		if game.worldDevTools.AsteroidsCanMove() {
			motion.AdvanceAsteroid(asteroid, delta, bounds)
		}
		if asteroid.ReadyForRemoval() {
			delete(game.state.Asteroids, id)
			continue
		}
		if game.isAsteroidFarFromAllCameras(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}

	for id, bullet := range game.state.Projectiles {
		if game.worldDevTools.BulletsCanMove() {
			motion.AdvanceBullet(bullet, delta, bounds)
		}
		if bullet.ReadyForRemoval() {
			delete(game.state.Projectiles, id)
			continue
		}
		if bullet.IsExpired() || game.isBulletFarFromAllCameras(bullet) {
			delete(game.state.Projectiles, id)
		}
	}

	if game.worldDevTools.CanRunCollisions() {
		game.handleShipAsteroidCollisions()
		game.handleBulletAsteroidCollisions()
	}
}

func (game *Game) statePacket(playerID string) StatePacket {
	players := make(map[string]entities.ShipState, len(game.state.Players))
	for id, player := range game.state.Players {
		players[id] = player.State()
	}

	asteroids := make(map[string]entities.AsteroidState, len(game.state.Asteroids))
	for id, asteroid := range game.state.Asteroids {
		asteroids[id] = asteroid.State()
	}

	bullets := make(map[string]entities.BulletState, len(game.state.Projectiles))
	for id, bullet := range game.state.Projectiles {
		bullets[id] = bullet.State()
	}
	events := append(make([]EventState, 0, len(game.pendingEvents[playerID])), game.pendingEvents[playerID]...)

	return StatePacket{
		Type:      PacketTypeState,
		SelfID:    playerID,
		Lives:     game.playerLives(playerID),
		Players:   players,
		Bullets:   bullets,
		Asteroids: asteroids,
		Events:    events,
	}
}

func (game *Game) playerLives(playerID string) int {
	if session, ok := game.playerSessions[playerID]; ok {
		return session.Lives
	}
	if player, ok := game.state.Players[playerID]; ok {
		return player.Lives
	}

	return 0
}
