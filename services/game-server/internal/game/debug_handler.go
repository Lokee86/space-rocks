package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugPacket(playerID string, player *entities.Ship, packet ClientPacket) bool {
	switch packet.Type {
	case PacketTypeToggleDebugInvincible:
		enabled := !player.DamageOptions.Invincible
		player.DamageOptions.SetInvincible(enabled)
		if session, ok := game.playerSessions[playerID]; ok {
			session.DamageOptions.SetInvincible(enabled)
		}
		logging.Game.Info("debug invincibility toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugInfiniteLives:
		enabled := false
		if session, ok := game.playerSessions[playerID]; ok {
			enabled = !session.LifeOptions.InfiniteLives
			session.LifeOptions.SetInfiniteLives(enabled)
		}
		logging.Game.Info("debug infinite lives toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugFreezeWorld:
		enabled := !game.worldSimulationOptions.IsWorldFrozen()
		game.worldSimulationOptions.SetFreezeWorld(enabled)
		logging.Game.Info("debug world freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeToggleDebugFreezePlayer:
		session, ok := game.playerSessions[playerID]
		if !ok {
			return true
		}
		enabled := !session.Suspension.DevFrozen
		session.Suspension.SetDevFrozen(enabled)
		player.ClearInput()
		logging.Game.Info("debug player freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
		return true
	case PacketTypeDebugKillPlayer:
		targetPlayerID := packet.TargetPlayerID
		if targetPlayerID == "" {
			targetPlayerID = playerID
		}
		targetPlayer, ok := game.state.Players[targetPlayerID]
		if !ok || targetPlayer == nil {
			return true
		}
		damageRequest := damage.DamageRequest{
			TargetEntityID:   targetPlayerID,
			TargetEntityType: damage.EntityTypePlayer,
			SourceEntityID:   playerID,
			SourceEntityType: damage.EntityTypePlayer,
			CurrentHealth:    targetPlayer.Health,
			Amount:           targetPlayer.Health,
			Type:             damage.DamageTypeDebug,
		}
		damageResult := damage.Resolve(damageRequest)
		targetPlayer.Health = damageResult.RemainingHealth
		if damageResult.Fatal {
			game.applyFatalPlayerDamage(targetPlayerID, targetPlayer)
		}
		return true
	default:
		return false
	}
}
