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
	nextBulletID         int
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
		Config: ClientConfig{
			VisibleWorldWidth:  constants.WorldWidth,
			VisibleWorldHeight: constants.WorldHeight,
		},
	}

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
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
		if game.isAsteroidFarFromAllPlayers(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}

	for id, bullet := range game.state.Projectiles {
		bullet.step(delta)
		if bullet.Life <= 0 || game.isBulletFarFromAllPlayers(bullet) {
			delete(game.state.Projectiles, id)
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

	bullets := make(map[string]BulletState, len(game.state.Projectiles))
	for id, bullet := range game.state.Projectiles {
		bullets[id] = bullet.State()
	}

	return StatePacket{
		Type:      "state",
		SelfID:    playerID,
		Players:   players,
		Bullets:   bullets,
		Asteroids: asteroids,
	}
}

func (game *Game) spawnBullet(ship *Ship) {
	forward := Vector2{X: 0, Y: -1}.rotated(ship.Rotation)

	game.nextBulletID++
	bulletID := fmt.Sprintf("bullet-%d", game.nextBulletID)
	game.state.Projectiles[bulletID] = &Bullet{
		ID:       bulletID,
		OwnerID:  ship.ID,
		X:        ship.X + forward.X*constants.BulletSpawnOffset,
		Y:        ship.Y + forward.Y*constants.BulletSpawnOffset,
		Rotation: ship.Rotation,
		Velocity: Vector2{
			X: forward.X * constants.BulletSpeed,
			Y: forward.Y * constants.BulletSpeed,
		},
		Life: constants.BulletLifetime,
	}
}

func (game *Game) spawnAsteroidBatch(target *Ship) {
	for range constants.AsteroidSpawnBatchSize {
		game.spawnAsteroid(target)
	}
}

func (game *Game) spawnAsteroid(target *Ship) {
	targetPosition := Vector2{X: target.X, Y: target.Y}
	spawn := game.randomAsteroidSpawnPosition(target)
	direction := spawn.directionTo(targetPosition).rotated(randomRange(
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

func (game *Game) randomAsteroidSpawnPosition(target *Ship) Vector2 {
	margin := constants.AsteroidSpawnMargin
	for attempts := 0; ; attempts++ {
		spawn := randomOffscreenPosition(target, margin)
		if !game.isOnscreenForAnyPlayer(spawn) {
			return spawn
		}

		if attempts > 0 && attempts%16 == 0 {
			margin += constants.AsteroidSpawnMargin
		}
	}
}

func randomOffscreenPosition(target *Ship, margin float64) Vector2 {
	width := target.visibleWorldWidth()
	height := target.visibleWorldHeight()
	left := target.X - width*0.5
	right := target.X + width*0.5
	top := target.Y - height*0.5
	bottom := target.Y + height*0.5

	switch rand.Intn(4) {
	case 0:
		return Vector2{X: randomRange(left, right), Y: top - margin}
	case 1:
		return Vector2{
			X: right + margin,
			Y: randomRange(top, bottom),
		}
	case 2:
		return Vector2{
			X: randomRange(left, right),
			Y: bottom + margin,
		}
	default:
		return Vector2{X: left - margin, Y: randomRange(top, bottom)}
	}
}

func (game *Game) isOnscreenForAnyPlayer(position Vector2) bool {
	for _, player := range game.state.Players {
		if player.isInsideView(position) {
			return true
		}
	}

	return false
}

func (game *Game) isAsteroidFarFromAllPlayers(asteroid *Asteroid) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.isFarFromView(Vector2{X: asteroid.X, Y: asteroid.Y}) {
			return false
		}
	}

	return true
}

func (game *Game) isBulletFarFromAllPlayers(bullet *Bullet) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.isFarFromView(Vector2{X: bullet.X, Y: bullet.Y}) {
			return false
		}
	}

	return true
}

func (ship *Ship) isInsideView(position Vector2) bool {
	width := ship.visibleWorldWidth()
	height := ship.visibleWorldHeight()
	left := ship.X - width*0.5
	right := ship.X + width*0.5
	top := ship.Y - height*0.5
	bottom := ship.Y + height*0.5

	return position.X >= left &&
		position.X <= right &&
		position.Y >= top &&
		position.Y <= bottom
}

func (ship *Ship) isFarFromView(position Vector2) bool {
	width := ship.visibleWorldWidth()
	height := ship.visibleWorldHeight()
	left := ship.X - width*0.5 - constants.AsteroidDespawnMargin
	right := ship.X + width*0.5 + constants.AsteroidDespawnMargin
	top := ship.Y - height*0.5 - constants.AsteroidDespawnMargin
	bottom := ship.Y + height*0.5 + constants.AsteroidDespawnMargin

	return position.X < left ||
		position.X > right ||
		position.Y < top ||
		position.Y > bottom
}

func (ship *Ship) visibleWorldWidth() float64 {
	if ship.Config.VisibleWorldWidth > 0 {
		return ship.Config.VisibleWorldWidth
	}

	return constants.WorldWidth
}

func (ship *Ship) visibleWorldHeight() float64 {
	if ship.Config.VisibleWorldHeight > 0 {
		return ship.Config.VisibleWorldHeight
	}

	return constants.WorldHeight
}

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
