package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type Game struct {
	mu                   sync.Mutex
	nextID               int
	nextAsteroidID       int
	asteroidSpawnElapsed float64
	state                GameState
}

func New() *Game {
	return &Game{
		state: NewGameState(),
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
	}

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
}

func (game *Game) HandlePacket(playerID string, packet InputPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}

	switch packet.Type {
	case "input":
		player.Input = packet.Input

		if packet.Input.Shoot {
			log.Println("shoot")
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
	}

	if len(game.state.Players) > 0 {
		game.asteroidSpawnElapsed += delta
		if game.asteroidSpawnElapsed >= constants.AsteroidSpawnInterval {
			game.asteroidSpawnElapsed = 0
			for _, player := range game.state.Players {
				game.spawnAsteroid(Vector2{X: player.X, Y: player.Y})
			}
		}
	} else {
		game.asteroidSpawnElapsed = 0
	}

	for id, asteroid := range game.state.Asteroids {
		asteroid.step(delta)
		if game.isAsteroidFarFromAllPlayers(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}
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

	return StatePacket{
		Type:      "state",
		SelfID:    playerID,
		Players:   players,
		Asteroids: asteroids,
	}
}

func (game *Game) spawnAsteroid(target Vector2) {
	spawn := randomOffscreenPosition(target)
	direction := spawn.directionTo(target).rotated(randomRange(
		-degreesToRadians(constants.AsteroidAimRandomnessDegrees),
		degreesToRadians(constants.AsteroidAimRandomnessDegrees),
	))
	speed := randomRange(constants.AsteroidMinSpeed, constants.AsteroidMaxSpeed)

	game.nextAsteroidID++
	asteroidID := fmt.Sprintf("asteroid-%d", game.nextAsteroidID)
	game.state.Asteroids[asteroidID] = &Asteroid{
		ID: asteroidID,
		X:  spawn.X,
		Y:  spawn.Y,
		Velocity: Vector2{
			X: direction.X * speed,
			Y: direction.Y * speed,
		},
		Size:    rand.Intn(4) + 1,
		Variant: rand.Intn(4),
	}
}

func randomOffscreenPosition(target Vector2) Vector2 {
	left := target.X - constants.WorldWidth*0.5
	right := target.X + constants.WorldWidth*0.5
	top := target.Y - constants.WorldHeight*0.5
	bottom := target.Y + constants.WorldHeight*0.5

	switch rand.Intn(4) {
	case 0:
		return Vector2{X: randomRange(left, right), Y: top - constants.AsteroidSpawnMargin}
	case 1:
		return Vector2{
			X: right + constants.AsteroidSpawnMargin,
			Y: randomRange(top, bottom),
		}
	case 2:
		return Vector2{
			X: randomRange(left, right),
			Y: bottom + constants.AsteroidSpawnMargin,
		}
	default:
		return Vector2{X: left - constants.AsteroidSpawnMargin, Y: randomRange(top, bottom)}
	}
}

func (game *Game) isAsteroidFarFromAllPlayers(asteroid *Asteroid) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !isFarFromPlayerView(Vector2{X: asteroid.X, Y: asteroid.Y}, Vector2{X: player.X, Y: player.Y}) {
			return false
		}
	}

	return true
}

func isFarFromPlayerView(position Vector2, playerPosition Vector2) bool {
	left := playerPosition.X - constants.WorldWidth*0.5 - constants.AsteroidDespawnMargin
	right := playerPosition.X + constants.WorldWidth*0.5 + constants.AsteroidDespawnMargin
	top := playerPosition.Y - constants.WorldHeight*0.5 - constants.AsteroidDespawnMargin
	bottom := playerPosition.Y + constants.WorldHeight*0.5 + constants.AsteroidDespawnMargin

	return position.X < left ||
		position.X > right ||
		position.Y < top ||
		position.Y > bottom
}

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
