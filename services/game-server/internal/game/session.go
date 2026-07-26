package game

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type playerSession struct {
	ID              string
	ShipTypeID      string
	Stats           runtime.ShipStats
	SpawnPosition   physics.Vector2
	Config          runtime.ClientConfig
	Targeting       PlayerTargeting
	Score           int
	Lives           int
	ShipDeaths      int
	RespawnCooldown float64
	Suspension      runtime.SuspensionState
	DamageOptions   runtime.DamageOptions
	LifeOptions     runtime.LifeOptions
	PlayerArmory    weapons.PlayerArmory
}

func newPlayerSession(id string, spawnPosition physics.Vector2) *playerSession {
	return &playerSession{
		ID:            id,
		ShipTypeID:    runtime.DefaultShipTypeID,
		Stats:         runtime.ResolveShipStats(runtime.DefaultShipTypeID),
		SpawnPosition: spawnPosition,
		Config:        runtime.DefaultCameraConfig(),
		Targeting:     EmptyPlayerTargeting(),
		Lives:         constants.PlayerStartingLives,
		PlayerArmory:  weapons.DefaultPlayerArmory(),
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

func (session *playerSession) NewShip(position physics.Vector2) *runtime.Ship {
	ship := &runtime.Ship{
		ID:            session.ID,
		ShipTypeID:    session.ShipTypeID,
		Stats:         session.Stats,
		X:             position.X,
		Y:             position.Y,
		Config:        session.Config,
		Health:        session.Stats.MaxHealth,
		DamageOptions: session.DamageOptions,
	}
	ship.ShipWeapons.Primary = session.PlayerArmory.Primary
	ship.ShipWeapons.Secondary = session.PlayerArmory.Secondary
	session.Targeting.ApplyToShip(ship)
	return ship
}

func (game *Game) respawnPlayer(playerID string) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		if game.matchID != "" && game.matchTraceID != "" {
			logging.Emit(observability.Request{
				Event: observability.EventNameRespawnBlocked,
				Context: observability.Context{
					TraceID:  game.matchTraceID,
					MatchID:  game.matchID,
					PlayerID: playerID,
				},
				Fields: observability.Fields{"reason_code": "session_missing"},
			})
		}
		return
	}
	if !session.CanRespawn() {
		if game.matchID != "" && game.matchTraceID != "" {
			logging.Emit(observability.Request{
				Event: observability.EventNameRespawnBlocked,
				Context: observability.Context{
					TraceID:  game.matchTraceID,
					MatchID:  game.matchID,
					PlayerID: playerID,
				},
				Fields: observability.Fields{
					"reason_code":      "respawn_cooldown_or_lives_exhausted",
					"lives":            session.Lives,
					"respawn_cooldown": session.RespawnCooldown,
				},
			})
		}
		return
	}
	if _, ok := game.entities.Players[playerID]; ok {
		if game.matchID != "" && game.matchTraceID != "" {
			logging.Emit(observability.Request{
				Event: observability.EventNameRespawnBlocked,
				Context: observability.Context{
					TraceID:  game.matchTraceID,
					MatchID:  game.matchID,
					PlayerID: playerID,
				},
				Fields: observability.Fields{"reason_code": "already_active"},
			})
		}
		return
	}

	spawnPlan := game.planPlayerRespawn(session)
	spawnPosition := spawnPlan.Position
	player := session.NewShip(spawnPosition)
	game.entities.Players[playerID] = player
	game.setPlayerCameraViewLocked(playerID, player)
}

func (game *Game) planInitialPlayerSpawn(playerIndex int, playerID string) PlayerSpawnPlan {
	shapeID := runtime.ResolveShipStats(runtime.DefaultShipTypeID).CollisionShapeID
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
	for _, asteroid := range game.entities.Asteroids {
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
	for id, player := range game.entities.Players {
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
