package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func HandleCommand(target *game.Game, playerID string, command DebugCommand) bool {
	switch command.Type {
	case PacketTypeToggleDebugInvincible:
		current, _ := target.DevtoolsPlayerInvincible(playerID)
		enabled := !current
		target.DevtoolsSetPlayerInvincible(playerID, enabled)
		logging.Game.Info("debug invincibility toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugInfiniteLives:
		current, _ := target.DevtoolsInfiniteLives(playerID)
		enabled := !current
		target.DevtoolsSetInfiniteLives(playerID, enabled)
		logging.Game.Info("debug infinite lives toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugFreezeWorld:
		enabled := !target.DevtoolsWorldFrozen()
		target.DevtoolsSetWorldFrozen(enabled)
		logging.Game.Info("debug world freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugFreezePlayer:
		current, _ := target.DevtoolsPlayerFrozen(playerID)
		enabled := !current
		target.DevtoolsSetPlayerFrozen(playerID, enabled)
		logging.Game.Info("debug player freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeDebugKillPlayer:
		targetPlayerID := command.TargetPlayerID
		if targetPlayerID == "" {
			targetPlayerID = playerID
		}
		target.DevtoolsKillPlayer(playerID, targetPlayerID)
		return true
	case PacketTypeDebugSpawnEntity:
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
		target.DevtoolsHandleSpawnEntity(playerID, request)
		return true
	case PacketTypeDebugRespawnPlayer:
		request := RespawnPlayerRequest{
			TargetPlayerID: command.TargetPlayerID,
			X:              command.X,
			Y:              command.Y,
		}
		target.DevtoolsHandleRespawnPlayer(playerID, request)
		return true
	default:
		return false
	}
}
