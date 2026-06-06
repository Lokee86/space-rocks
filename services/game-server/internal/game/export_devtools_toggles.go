package game

import "github.com/Lokee86/space-rocks/server/internal/game/damage"

func (game *Game) DevtoolsWorldFrozen() bool {
	return game.worldSimulationOptions.IsWorldFrozen()
}

func (game *Game) DevtoolsSetWorldFrozen(enabled bool) {
	game.worldSimulationOptions.SetFreezeWorld(enabled)
}

func (game *Game) DevtoolsToggleFreezeWorld() bool {
	return game.worldSimulationOptions.ToggleFreezeWorld()
}

func (game *Game) DevtoolsToggleFreezeAsteroids() bool {
	return game.worldSimulationOptions.ToggleFreezeAsteroids()
}

func (game *Game) DevtoolsToggleFreezeBullets() bool {
	return game.worldSimulationOptions.ToggleFreezeBullets()
}

func (game *Game) DevtoolsToggleFreezeSpawning() bool {
	return game.worldSimulationOptions.ToggleFreezeSpawning()
}

func (game *Game) DevtoolsToggleFreezeCollisions() bool {
	return game.worldSimulationOptions.ToggleFreezeCollisions()
}

func (game *Game) DevtoolsPlayerInvincible(playerID string) (bool, bool) {
	found := false
	invincible := false

	if session, ok := game.playerSessions[playerID]; ok {
		invincible = session.DamageOptions.Invincible
		found = true
	}

	if player, ok := game.entities.Players[playerID]; ok {
		invincible = player.DamageOptions.Invincible
		found = true
	}

	return invincible, found
}

func (game *Game) DevtoolsSetPlayerInvincible(playerID string, enabled bool) bool {
	found := false

	if session, ok := game.playerSessions[playerID]; ok {
		session.DamageOptions.Invincible = enabled
		found = true
	}

	if player, ok := game.entities.Players[playerID]; ok {
		player.DamageOptions.Invincible = enabled
		found = true
	}

	return found
}

func (game *Game) DevtoolsInfiniteLives(playerID string) (bool, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.LifeOptions.InfiniteLives, true
}

func (game *Game) DevtoolsSetInfiniteLives(playerID string, enabled bool) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.LifeOptions.InfiniteLives = enabled
	return true
}

func (game *Game) DevtoolsPlayerFrozen(playerID string) (bool, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.Suspension.DevFrozen, true
}

func (game *Game) DevtoolsSetPlayerFrozen(playerID string, enabled bool) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.Suspension.SetDevFrozen(enabled)
	if enabled {
		if player, ok := game.entities.Players[playerID]; ok {
			player.ClearInput()
		}
	}
	return true
}

func (game *Game) DevtoolsKillPlayer(sourcePlayerID string, targetPlayerID string) bool {
	targetPlayer, ok := game.entities.Players[targetPlayerID]
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
			Kind:   damage.DamageKindKinetic,
			Cause:  damage.DamageCauseDebug,
		},
	}
	damageResult := damage.ResolveSingle(damageRequest)
	targetPlayer.Health = damageResult.RemainingHealth
	targetPlayer.Shields = damageResult.RemainingShield
	if damageResult.Fatal {
		game.applyFatalPlayerDamage(targetPlayerID, targetPlayer)
	}
	return true
}
