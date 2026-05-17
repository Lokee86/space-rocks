package game

import (
	"math"
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type playerSession struct {
	ID              string
	SpawnPosition   physics.Vector2
	Config          entities.ClientConfig
	Score           int
	Lives           int
	RespawnCooldown float64
}

func newPlayerSession(id string, spawnPosition physics.Vector2) *playerSession {
	return &playerSession{
		ID:            id,
		SpawnPosition: spawnPosition,
		Config: entities.ClientConfig{
			VisibleWorldWidth:  constants.WorldWidth,
			VisibleWorldHeight: constants.WorldHeight,
		},
		Lives: constants.PlayerStartingLives,
	}
}

func (session *playerSession) Step(delta float64) {
	if session.RespawnCooldown > 0 {
		session.RespawnCooldown = max(0, session.RespawnCooldown-delta)
	}
}

func (session *playerSession) CanRespawn() bool {
	return session.Lives > 0 && session.RespawnCooldown == 0
}

func (session *playerSession) RecordDeath() {
	if session.Lives > 0 {
		session.Lives--
	}
	if session.Lives > 0 {
		session.RespawnCooldown = constants.PlayerRespawnDelay
	}
}

func (session *playerSession) NewShip(position physics.Vector2) *entities.Ship {
	return &entities.Ship{
		ID:     session.ID,
		X:      position.X,
		Y:      position.Y,
		Config: session.Config,
		Score:  session.Score,
		Lives:  session.Lives,
	}
}

func (game *Game) respawnPlayer(playerID string) {
	session, ok := game.playerSessions[playerID]
	if !ok || !session.CanRespawn() {
		return
	}
	if _, ok := game.state.Players[playerID]; ok {
		return
	}

	spawnPosition := game.safeRespawnPosition(session)
	player := session.NewShip(spawnPosition)
	game.state.Players[playerID] = player
	game.cameraViews[playerID] = &entities.CameraView{
		X:      player.X,
		Y:      player.Y,
		Config: player.Config,
	}
}

func (game *Game) safeRespawnPosition(session *playerSession) physics.Vector2 {
	if game.isSafeRespawnPosition(session.SpawnPosition) {
		return session.SpawnPosition
	}

	for attempt := 0; attempt < 32; attempt++ {
		radius := constants.PlayerRespawnSearchRadius * (1 + float64(attempt/8))
		position := session.SpawnPosition.Add(randomUnitVector().Multiply(rand.Float64() * radius))
		if game.isSafeRespawnPosition(position) {
			return position
		}
	}

	return session.SpawnPosition.Add(physics.Vector2{X: constants.PlayerRespawnSearchRadius, Y: 0}.Rotated(rand.Float64() * math.Pi * 2))
}

func (game *Game) isSafeRespawnPosition(position physics.Vector2) bool {
	shape, err := game.collisionShapes.ShipShape()
	if err != nil {
		return true
	}

	shipBody := physics.CollisionBody{
		ID:       "respawn",
		Position: position,
		Shape:    shape,
	}
	for _, asteroid := range game.state.Asteroids {
		if asteroid.IsPendingDespawn() {
			continue
		}

		asteroidBody, ok := asteroid.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		if _, ok := physics.DetectCollision(shipBody, asteroidBody); ok {
			return false
		}
	}

	return true
}
