package game

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type Game struct {
	mu                   sync.Mutex
	nextID               int
	nextAsteroidID       int
	nextBulletID         int
	asteroidSpawnElapsed float64
	collisionShapes      physics.CollisionShapeCatalog
	state                entities.GameState
	cameraViews          map[string]*entities.CameraView
	pendingEvents        map[string][]EventState
}

func New() *Game {
	collisionShapes, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		log.Println("collision shapes unavailable:", err)
	}

	return &Game{
		collisionShapes: collisionShapes,
		cameraViews:     make(map[string]*entities.CameraView),
		pendingEvents:   make(map[string][]EventState),
		state:           entities.NewGameState(),
	}
}

func (game *Game) Start() {
	go game.runSimulation()
}

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	player := &entities.Ship{
		ID: playerID,
		X:  576 + float64(playerIndex%4)*80,
		Y:  320 + float64(playerIndex/4)*80,
		Config: entities.ClientConfig{
			VisibleWorldWidth:  constants.WorldWidth,
			VisibleWorldHeight: constants.WorldHeight,
		},
	}
	game.state.Players[playerID] = player
	game.cameraViews[playerID] = &entities.CameraView{
		X:      player.X,
		Y:      player.Y,
		Config: player.Config,
	}
	game.pendingEvents[playerID] = nil

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
	delete(game.cameraViews, playerID)
	delete(game.pendingEvents, playerID)
}

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	switch packet.Type {
	case PacketTypeInput:
		if player.IsPendingDespawn() {
			return
		}
		player.SetInput(packet.Input)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
		if cameraView, ok := game.cameraViews[playerID]; ok {
			cameraView.SetConfig(packet.Config)
		}
	}
}

func (game *Game) State(playerID string) []byte {
	game.mu.Lock()
	defer game.mu.Unlock()

	response, err := json.Marshal(game.statePacket(playerID))
	if err != nil {
		log.Println(err)
		return nil
	}
	game.pendingEvents[playerID] = nil

	return response
}

func (game *Game) runSimulation() {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	delta := 1.0 / float64(constants.ServerTickRate)
	for range ticker.C {
		game.Step(delta)
	}
}

func (game *Game) Step(delta float64) {
	game.mu.Lock()
	defer game.mu.Unlock()

	for _, player := range game.state.Players {
		player.ApplyInput(delta)
		if cameraView, ok := game.cameraViews[player.ID]; ok {
			cameraView.SetPosition(player.Position())
		}
		if player.IsPendingDespawn() {
			continue
		}
		if player.WantsToShoot() && player.CanShoot() {
			game.spawnBullet(player)
			player.ResetShootCooldown()
		}
	}

	for id, player := range game.state.Players {
		if player.ReadyForRemoval() {
			delete(game.state.Players, id)
		}
	}

	if game.hasCameraViews() {
		game.asteroidSpawnElapsed += delta
		if game.asteroidSpawnElapsed >= constants.AsteroidSpawnInterval {
			game.asteroidSpawnElapsed = 0
			for _, cameraView := range game.cameraViews {
				game.spawnAsteroidBatch(cameraView)
			}
		}
	} else {
		game.asteroidSpawnElapsed = 0
	}

	for id, asteroid := range game.state.Asteroids {
		asteroid.Step(delta)
		if asteroid.ReadyForRemoval() {
			delete(game.state.Asteroids, id)
			continue
		}
		if game.isAsteroidFarFromAllCameras(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}

	for id, bullet := range game.state.Projectiles {
		bullet.Step(delta)
		if bullet.ReadyForRemoval() {
			delete(game.state.Projectiles, id)
			continue
		}
		if bullet.IsExpired() || game.isBulletFarFromAllCameras(bullet) {
			delete(game.state.Projectiles, id)
		}
	}

	game.handleShipAsteroidCollisions()
	game.handleBulletAsteroidCollisions()
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
		Players:   players,
		Bullets:   bullets,
		Asteroids: asteroids,
		Events:    events,
	}
}
