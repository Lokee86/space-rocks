package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

type playerSession struct {
	ID              string
	ShipTypeID      string
	Stats           entities.ShipStats
	SpawnPosition   physics.Vector2
	Config          entities.ClientConfig
	Score           int
	Lives           int
	RespawnCooldown float64
	Suspension      entities.SuspensionState
	DamageOptions   entities.DamageOptions
	LifeOptions     entities.LifeOptions
}

func newPlayerSession(id string, spawnPosition physics.Vector2) *playerSession {
	return &playerSession{
		ID:            id,
		ShipTypeID:    entities.DefaultShipTypeID,
		Stats:         entities.ResolveShipStats(entities.DefaultShipTypeID),
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
	if session.LifeOptions.CanLoseLives() && session.Lives > 0 {
		session.Lives--
	}
	if session.Lives > 0 {
		session.RespawnCooldown = constants.PlayerRespawnDelay
	}
}

func (session *playerSession) NewShip(position physics.Vector2) *entities.Ship {
	return &entities.Ship{
		ID:         session.ID,
		ShipTypeID: session.ShipTypeID,
		Stats:      session.Stats,
		X:          position.X,
		Y:          position.Y,
		Config:     session.Config,
		Score:      session.Score,
		Lives:      session.Lives,
		Health:     session.Stats.MaxHealth,
		DamageOptions: session.DamageOptions,
	}
}

func (game *Game) respawnPlayer(playerID string) {
	logging.Game.Info("respawn requested", logging.FieldPlayerID, playerID)

	session, ok := game.playerSessions[playerID]
	if !ok {
		logging.Game.Warn("respawn blocked; session missing", logging.FieldPlayerID, playerID)
		return
	}
	if !session.CanRespawn() {
		logging.Game.Info("respawn blocked",
			logging.FieldPlayerID, playerID,
			"lives", session.Lives,
			"respawn_cooldown", session.RespawnCooldown,
		)
		return
	}
	if _, ok := game.state.Players[playerID]; ok {
		logging.Game.Info("respawn blocked; player already active", logging.FieldPlayerID, playerID)
		return
	}

	spawnPlan := game.planPlayerRespawn(session)
	spawnPosition := spawnPlan.Position
	player := session.NewShip(spawnPosition)
	game.state.Players[playerID] = player
	game.cameraViews[playerID] = &entities.CameraView{
		X:      player.X,
		Y:      player.Y,
		Config: player.Config,
	}
	logging.Game.Info("player respawned",
		logging.FieldPlayerID, playerID,
		"x", spawnPosition.X,
		"y", spawnPosition.Y,
		"lives", session.Lives,
	)
}

func (game *Game) planInitialPlayerSpawn(playerIndex int, playerID string) PlayerSpawnPlan {
	shapeID := entities.ResolveShipStats(entities.DefaultShipTypeID).CollisionShapeID
	return PlayerSpawnPlan{
		EntityType: SpawnEntityTypePlayer,
		Reason:     SpawnReasonInitialPlayer,
		PlayerID:   playerID,
		Position:   game.safePlayerSpawnPosition(preferredInitialSpawnPosition(playerIndex), playerID, shapeID),
	}
}

func preferredInitialSpawnPosition(playerIndex int) physics.Vector2 {
	return physics.Vector2{
		X: 576 + float64(playerIndex%4)*80,
		Y: 320 + float64(playerIndex/4)*80,
	}
}

func (game *Game) safeRespawnPosition(session *playerSession) physics.Vector2 {
	return game.safePlayerSpawnPosition(session.SpawnPosition, session.ID, session.Stats.CollisionShapeID)
}

func (game *Game) planPlayerRespawn(session *playerSession) PlayerSpawnPlan {
	return PlayerSpawnPlan{
		EntityType: SpawnEntityTypePlayer,
		Reason:     SpawnReasonPlayerRespawn,
		PlayerID:   session.ID,
		Position:   game.safeRespawnPosition(session),
	}
}

func (game *Game) safePlayerSpawnPosition(origin physics.Vector2, ignorePlayerID string, collisionShapeID string) physics.Vector2 {
	if game.isSafeRespawnPosition(origin, ignorePlayerID, collisionShapeID) {
		return origin
	}

	spacing := respawnSearchSpacing()
	for ring := 1; ; ring++ {
		for x := -ring; x <= ring; x++ {
			top := origin.Add(physics.Vector2{X: float64(x) * spacing, Y: -float64(ring) * spacing})
			if game.isSafeRespawnPosition(top, ignorePlayerID, collisionShapeID) {
				return top
			}

			bottom := origin.Add(physics.Vector2{X: float64(x) * spacing, Y: float64(ring) * spacing})
			if game.isSafeRespawnPosition(bottom, ignorePlayerID, collisionShapeID) {
				return bottom
			}
		}

		for y := -ring + 1; y <= ring-1; y++ {
			left := origin.Add(physics.Vector2{X: -float64(ring) * spacing, Y: float64(y) * spacing})
			if game.isSafeRespawnPosition(left, ignorePlayerID, collisionShapeID) {
				return left
			}

			right := origin.Add(physics.Vector2{X: float64(ring) * spacing, Y: float64(y) * spacing})
			if game.isSafeRespawnPosition(right, ignorePlayerID, collisionShapeID) {
				return right
			}
		}
	}
}

func (game *Game) isSafeRespawnPosition(position physics.Vector2, ignorePlayerID string, collisionShapeID string) bool {
	shape, err := game.collisionShapes.ShipShapeByID(collisionShapeID)
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
	for id, player := range game.state.Players {
		if id == ignorePlayerID || player.IsPendingDespawn() {
			continue
		}

		playerBody, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		if !hasRespawnClearance(shipBody, playerBody, constants.PlayerRespawnBuffer) {
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
	return space.Distance(shipBody.Position, asteroidBody.Position) > clearance
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
