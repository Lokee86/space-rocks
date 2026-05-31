package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func handleDebugSpawnEntity(target *game.Game, playerID string, command DebugCommand) bool {
	request := SpawnEntityRequest{
		EntityType:     command.EntityType,
		X:              command.X,
		Y:              command.Y,
		HasDirection:   command.HasDirection,
		DirectionX:     command.DirectionX,
		DirectionY:     command.DirectionY,
		TargetPlayerID: command.TargetPlayerID,
	}
	if request.EntityType == EntityTypePlayer {
		spawnedPlayerID, spawnPosition, ok := applyDebugSpawnPlayer(target, request)
		if !ok {
			logging.Game.Info("debug player spawn ignored",
				logging.FieldPlayerID, playerID,
				"target_player_id", request.TargetPlayerID,
			)
			return true
		}
		logging.Game.Info("debug player spawned",
			logging.FieldPlayerID, playerID,
			"spawned_player_id", spawnedPlayerID,
			"x", spawnPosition.X,
			"y", spawnPosition.Y,
			"has_target_player_id", request.TargetPlayerID != "",
		)
		return true
	}
	if request.EntityType == EntityTypeBullet {
		bullet, ok := applyDebugSpawnBullet(target, playerID, request)
		if !ok {
			logging.Game.Info("debug bullet spawn ignored",
				logging.FieldPlayerID, playerID,
			)
			return true
		}
		logging.Game.Info("debug bullet spawned",
			logging.FieldPlayerID, playerID,
			"bullet_id", bullet.ID,
			"owner_player_id", playerID,
			"x", bullet.X,
			"y", bullet.Y,
			"has_direction", request.HasDirection,
		)
		return true
	}
	if request.EntityType == EntityTypeAsteroid {
		asteroid, plan, ok := applyDebugSpawnAsteroid(target, request)
		if !ok {
			logging.Game.Info("debug asteroid spawn ignored",
				logging.FieldPlayerID, playerID,
			)
			return true
		}
		logging.Game.Info("debug asteroid spawned",
			logging.FieldPlayerID, playerID,
			"asteroid_id", asteroid.ID,
			"x", plan.Position.X,
			"y", plan.Position.Y,
			"has_direction", request.HasDirection,
		)
		return true
	}
	logging.Game.Info("debug spawn entity not implemented for entity type",
		logging.FieldPlayerID, playerID,
		"entity_type", request.EntityType,
	)
	return true
}
