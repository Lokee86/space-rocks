package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugSpawnEntity(playerID string, packet ClientPacket) bool {
	request := debugging.SpawnEntityRequest{
		EntityType:     packet.EntityType,
		X:              packet.X,
		Y:              packet.Y,
		HasDirection:   packet.HasDirection,
		DirectionX:     packet.DirectionX,
		DirectionY:     packet.DirectionY,
		TargetPlayerID: packet.TargetPlayerID,
	}
	if request.EntityType == debugging.EntityTypePlayer {
		spawnedPlayerID, ok := game.applyDebugSpawnPlayer(request)
		if !ok {
			logging.Game.Info("debug player spawn ignored",
				logging.FieldPlayerID, playerID,
				"target_player_id", request.TargetPlayerID,
			)
			return true
		}

		spawnedPlayer := game.state.Players[spawnedPlayerID]
		logging.Game.Info("debug player spawned",
			logging.FieldPlayerID, playerID,
			"spawned_player_id", spawnedPlayerID,
			"x", spawnedPlayer.X,
			"y", spawnedPlayer.Y,
			"has_target_player_id", request.TargetPlayerID != "",
		)
		return true
	}

	if request.EntityType == debugging.EntityTypeBullet {
		bulletID, ok := game.applyDebugSpawnBullet(playerID, request)
		if !ok {
			logging.Game.Info("debug bullet spawn ignored",
				logging.FieldPlayerID, playerID,
			)
			return true
		}

		bullet := game.state.Projectiles[bulletID]
		logging.Game.Info("debug bullet spawned",
			logging.FieldPlayerID, playerID,
			"bullet_id", bulletID,
			"owner_player_id", playerID,
			"x", bullet.X,
			"y", bullet.Y,
			"has_direction", request.HasDirection,
		)
		return true
	}

	if request.EntityType != debugging.EntityTypeAsteroid {
		logging.Game.Info("debug spawn entity not implemented for entity type",
			logging.FieldPlayerID, playerID,
			"entity_type", request.EntityType,
		)
		return true
	}

	plan := game.buildDebugAsteroidSpawnPlan(request)
	asteroid := game.applyAsteroidSpawn(plan)
	logging.Game.Info("debug asteroid spawned",
		logging.FieldPlayerID, playerID,
		"asteroid_id", asteroid.ID,
		"x", plan.Position.X,
		"y", plan.Position.Y,
		"has_direction", request.HasDirection,
	)
	return true
}
