package game

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

const deathmatchSpawnCandidateCount = 48

type playerSpawnContext struct {
	Reason              SpawnReason
	PlayerID            string
	TeamID              teams.ID
	PlayerIndex         int
	PreviousPosition    physics.Vector2
	HasPreviousPosition bool
	CollisionShapeID    string
}

func (game *Game) planInitialPlayerSpawn(playerIndex int, playerID string, teamID teams.ID) PlayerSpawnPlan {
	return game.planPlayerSpawn(playerSpawnContext{
		Reason:           SpawnReasonInitialPlayer,
		PlayerID:         playerID,
		TeamID:           teamID,
		PlayerIndex:      playerIndex,
		CollisionShapeID: runtime.ResolveShipStats(runtime.DefaultShipTypeID).CollisionShapeID,
	})
}

func (game *Game) planPlayerRespawn(session *playerSession) PlayerSpawnPlan {
	return game.planPlayerSpawn(playerSpawnContext{
		Reason:              SpawnReasonPlayerRespawn,
		PlayerID:            session.ID,
		TeamID:              session.TeamID,
		PreviousPosition:    session.SpawnPosition,
		HasPreviousPosition: true,
		CollisionShapeID:    session.Stats.CollisionShapeID,
	})
}

func (game *Game) planPlayerSpawn(context playerSpawnContext) PlayerSpawnPlan {
	preferred := game.preferredPlayerSpawnPosition(context)
	return PlayerSpawnPlan{
		EntityType: SpawnEntityTypePlayer,
		Reason:     context.Reason,
		PlayerID:   context.PlayerID,
		Position:   game.safePlayerSpawnPosition(preferred, context.PlayerID, context.CollisionShapeID),
	}
}

func (game *Game) safeRespawnPosition(session *playerSession) physics.Vector2 {
	return game.planPlayerRespawn(session).Position
}

func (game *Game) preferredPlayerSpawnPosition(context playerSpawnContext) physics.Vector2 {
	switch game.resolvedMatchRules.PlayerSpawnProfileID {
	case modes.DeathmatchSpawnProfileID:
		return game.preferredDeathmatchSpawnPosition(context)
	case modes.BasicSafeSpawnProfileID:
		return preferredBasicSafeSpawnPosition(context)
	default:
		return preferredBasicSafeSpawnPosition(context)
	}
}

func preferredBasicSafeSpawnPosition(context playerSpawnContext) physics.Vector2 {
	if context.Reason == SpawnReasonPlayerRespawn && context.HasPreviousPosition {
		return context.PreviousPosition
	}
	return physics.Vector2{
		X: 576 + float64(context.PlayerIndex%4)*80,
		Y: 320 + float64(context.PlayerIndex/4)*80,
	}
}

func (game *Game) preferredDeathmatchSpawnPosition(context playerSpawnContext) physics.Vector2 {
	bounds := space.DefaultBounds()
	maxDistance := math.Hypot(bounds.Width*0.5, bounds.Height*0.5)
	bestPosition := game.randomWorldPosition(bounds)
	bestScore := math.Inf(-1)
	foundSafe := false

	for candidateIndex := 0; candidateIndex < deathmatchSpawnCandidateCount; candidateIndex++ {
		candidate := game.randomWorldPosition(bounds)
		safe := game.isSafeRespawnPosition(candidate, context.PlayerID, context.CollisionShapeID)
		if foundSafe && !safe {
			continue
		}

		score := game.deathmatchSpawnScore(candidate, context, bounds, maxDistance)
		if safe && !foundSafe {
			foundSafe = true
			bestScore = math.Inf(-1)
		}
		if safe == foundSafe && score > bestScore {
			bestPosition = candidate
			bestScore = score
		}
	}

	return bestPosition
}

func (game *Game) deathmatchSpawnScore(candidate physics.Vector2, context playerSpawnContext, bounds space.Bounds, maxDistance float64) float64 {
	nearestEnemy := maxDistance
	nearestTeammate := maxDistance
	hasEnemy := false
	hasTeammate := false

	for playerID, player := range game.entities.Players {
		if playerID == context.PlayerID || player.IsPendingDespawn() {
			continue
		}
		distance := space.WrappedDistance(candidate, physics.Vector2{X: player.X, Y: player.Y}, bounds)
		session := game.playerSessions[playerID]
		if context.TeamID != teams.NoTeam && session != nil && session.TeamID == context.TeamID {
			hasTeammate = true
			nearestTeammate = math.Min(nearestTeammate, distance)
			continue
		}
		hasEnemy = true
		nearestEnemy = math.Min(nearestEnemy, distance)
	}

	score := nearestEnemy
	if !hasEnemy {
		score = maxDistance
	}
	if hasTeammate {
		preferredTeammateDistance := math.Min(bounds.Width, bounds.Height) * 0.08
		score -= math.Abs(nearestTeammate-preferredTeammateDistance) * 0.7
	}
	if context.HasPreviousPosition {
		previousDistance := space.WrappedDistance(candidate, context.PreviousPosition, bounds)
		score += math.Min(previousDistance, maxDistance) * 0.25
	}
	return score
}

func (game *Game) randomWorldPosition(bounds space.Bounds) physics.Vector2 {
	return physics.Vector2{
		X: game.rngSource.Float64() * bounds.Width,
		Y: game.rngSource.Float64() * bounds.Height,
	}
}
