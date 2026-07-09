package game

import "github.com/Lokee86/space-rocks/server/internal/game/damage"

func (target *Control) SetWorldFrozen(enabled bool) {
	target.game.worldSimulationOptions.SetFreezeWorld(enabled)
}

func (target *Control) ToggleFreezeWorld() bool {
	return target.game.worldSimulationOptions.ToggleFreezeWorld()
}

func (target *Control) ToggleFreezeAsteroids() bool {
	return target.game.worldSimulationOptions.ToggleFreezeAsteroids()
}

func (target *Control) ToggleFreezeBullets() bool {
	return target.game.worldSimulationOptions.ToggleFreezeBullets()
}

func (target *Control) ToggleFreezeSpawning() bool {
	return target.game.worldSimulationOptions.ToggleFreezeSpawning()
}

func (target *Control) ToggleFreezeCollisions() bool {
	return target.game.worldSimulationOptions.ToggleFreezeCollisions()
}

func (target *Control) PlayerInvincible(playerID string) (bool, bool) {
	found := false
	invincible := false

	if session, ok := target.game.playerSessions[playerID]; ok {
		invincible = session.DamageOptions.Invincible
		found = true
	}

	if player, ok := target.game.entities.Players[playerID]; ok {
		invincible = player.DamageOptions.Invincible
		found = true
	}

	return invincible, found
}

func (target *Control) SetPlayerInvincible(playerID string, enabled bool) bool {
	found := false

	if session, ok := target.game.playerSessions[playerID]; ok {
		session.DamageOptions.Invincible = enabled
		found = true
	}

	if player, ok := target.game.entities.Players[playerID]; ok {
		player.DamageOptions.Invincible = enabled
		found = true
	}

	return found
}

func (target *Control) InfiniteLives(playerID string) (bool, bool) {
	session, ok := target.game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.LifeOptions.InfiniteLives, true
}

func (target *Control) SetInfiniteLives(playerID string, enabled bool) bool {
	session, ok := target.game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.LifeOptions.InfiniteLives = enabled
	return true
}

func (target *Control) PlayerFrozen(playerID string) (bool, bool) {
	session, ok := target.game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.Suspension.DevFrozen, true
}

func (target *Control) SetPlayerFrozen(playerID string, enabled bool) bool {
	session, ok := target.game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.Suspension.SetDevFrozen(enabled)
	if enabled {
		if player, ok := target.game.entities.Players[playerID]; ok {
			player.ClearInput()
		}
	}
	return true
}

func (target *Control) ApplyPlayerDefeat(sourcePlayerID string, targetPlayerID string) bool {
	targetPlayer, ok := target.game.entities.Players[targetPlayerID]
	if !ok || targetPlayer == nil {
		return true
	}
	damageRequest := damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   sourcePlayerID,
			EntityType: damage.EntityTypePlayer,
			Cause:      damage.DamageCauseDebug,
		},
		Target: damage.DamageTarget{
			EntityID:   targetPlayerID,
			EntityType: damage.EntityTypePlayer,
			Health:     targetPlayer.Health,
			Shield:     targetPlayer.Shields,
		},
		Spec: damage.DamageSpec{
			Amount: targetPlayer.Health,
			Type:   damage.DamageTypeKinetic,
			Cause:  damage.DamageCauseDebug,
		},
	}
	damageResult := damage.ResolveSingle(damageRequest)
	targetPlayer.Health = damageResult.RemainingHealth
	targetPlayer.Shields = damageResult.RemainingShield
	if damageResult.Fatal {
		target.game.applyFatalPlayerDamage(targetPlayerID, targetPlayer)
	}
	return true
}
