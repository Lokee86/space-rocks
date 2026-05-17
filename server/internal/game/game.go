package game

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type Game struct {
	mu                   sync.Mutex
	nextID               int
	nextAsteroidID       int
	nextBulletID         int
	asteroidSpawnElapsed float64
	collisionShapes      CollisionShapeCatalog
	state                GameState
	pendingEvents        map[string][]EventState
}

func New() *Game {
	collisionShapes, err := LoadCollisionShapeCatalog()
	if err != nil {
		log.Println("collision shapes unavailable:", err)
	}

	return &Game{
		collisionShapes: collisionShapes,
		pendingEvents:   make(map[string][]EventState),
		state:           NewGameState(),
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
	game.state.Players[playerID] = &Ship{
		ID: playerID,
		X:  576 + float64(playerIndex%4)*80,
		Y:  320 + float64(playerIndex/4)*80,
		Config: ClientConfig{
			VisibleWorldWidth:  constants.WorldWidth,
			VisibleWorldHeight: constants.WorldHeight,
		},
	}
	game.pendingEvents[playerID] = nil

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
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
	case "input":
		player.Input = packet.Input
	case "client_config":
		player.Config = packet.Config
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
		player.applyInput(delta)
		if player.Input.Shoot && player.ShootCooldown == 0 {
			game.spawnBullet(player)
			player.ShootCooldown = constants.BulletCooldown
		}
	}

	if len(game.state.Players) > 0 {
		game.asteroidSpawnElapsed += delta
		if game.asteroidSpawnElapsed >= constants.AsteroidSpawnInterval {
			game.asteroidSpawnElapsed = 0
			for _, player := range game.state.Players {
				game.spawnAsteroidBatch(player)
			}
		}
	} else {
		game.asteroidSpawnElapsed = 0
	}

	for id, asteroid := range game.state.Asteroids {
		asteroid.step(delta)
		if asteroid.PendingDespawn && asteroid.DespawnDelay <= 0 {
			delete(game.state.Asteroids, id)
			continue
		}
		if game.isAsteroidFarFromAllPlayers(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}

	for id, bullet := range game.state.Projectiles {
		bullet.step(delta)
		if bullet.PendingDespawn && bullet.DespawnDelay <= 0 {
			delete(game.state.Projectiles, id)
			continue
		}
		if bullet.Life <= 0 || game.isBulletFarFromAllPlayers(bullet) {
			delete(game.state.Projectiles, id)
		}
	}

	game.handleBulletAsteroidCollisions()
}

func (game *Game) statePacket(playerID string) StatePacket {
	players := make(map[string]ShipState, len(game.state.Players))
	for id, player := range game.state.Players {
		players[id] = player.State()
	}

	asteroids := make(map[string]AsteroidState, len(game.state.Asteroids))
	for id, asteroid := range game.state.Asteroids {
		asteroids[id] = asteroid.State()
	}

	bullets := make(map[string]BulletState, len(game.state.Projectiles))
	for id, bullet := range game.state.Projectiles {
		bullets[id] = bullet.State()
	}
	events := append(make([]EventState, 0, len(game.pendingEvents[playerID])), game.pendingEvents[playerID]...)

	return StatePacket{
		Type:      "state",
		SelfID:    playerID,
		Players:   players,
		Bullets:   bullets,
		Asteroids: asteroids,
		Events:    events,
	}
}
