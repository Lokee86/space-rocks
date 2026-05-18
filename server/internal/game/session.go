package game

import (
	"math"

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

	spacing := respawnSearchSpacing()
	for ring := 1; ; ring++ {
		for x := -ring; x <= ring; x++ {
			top := session.SpawnPosition.Add(physics.Vector2{X: float64(x) * spacing, Y: -float64(ring) * spacing})
			if game.isSafeRespawnPosition(top) {
				return top
			}

			bottom := session.SpawnPosition.Add(physics.Vector2{X: float64(x) * spacing, Y: float64(ring) * spacing})
			if game.isSafeRespawnPosition(bottom) {
				return bottom
			}
		}

		for y := -ring + 1; y <= ring-1; y++ {
			left := session.SpawnPosition.Add(physics.Vector2{X: -float64(ring) * spacing, Y: float64(y) * spacing})
			if game.isSafeRespawnPosition(left) {
				return left
			}

			right := session.SpawnPosition.Add(physics.Vector2{X: float64(ring) * spacing, Y: float64(y) * spacing})
			if game.isSafeRespawnPosition(right) {
				return right
			}
		}
	}
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
		if !hasRespawnClearance(shipBody, asteroidBody, constants.PlayerRespawnBuffer) {
			return false
		}
	}

	return true
}

func respawnSearchSpacing() float64 {
	return max(64, constants.PlayerRespawnBuffer)
}

func hasRespawnClearance(shipBody physics.CollisionBody, asteroidBody physics.CollisionBody, buffer float64) bool {
	clearance := collisionShapeRadius(shipBody.Shape) + collisionShapeRadius(asteroidBody.Shape) + buffer
	return shipBody.Position.Subtract(asteroidBody.Position).LengthSquared() > clearance*clearance
}

func collisionShapeRadius(shape physics.CollisionShape) float64 {
	switch shape.Type {
	case physics.CollisionShapeCircle:
		return shape.Radius
	case physics.CollisionShapeCapsule:
		return shape.Height * 0.5
	case physics.CollisionShapeRectangle:
		return shape.Size.Multiply(0.5).Length()
	case physics.CollisionShapePolygon:
		var radius float64
		for _, point := range shape.Points {
			radius = math.Max(radius, point.Length())
		}
		return radius
	default:
		return 0
	}
}
