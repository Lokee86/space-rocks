package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
)

func (target *Control) SetWorldFrozen(enabled bool) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	target.game.worldSimulationOptions.SetFreezeWorld(enabled)
}

func (target *Control) ToggleFreezeWorld() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.ToggleFreezeWorld()
}

func (target *Control) ToggleFreezeAsteroids() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.ToggleFreezeAsteroids()
}

func (target *Control) ToggleFreezeBullets() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.ToggleFreezeBullets()
}

func (target *Control) ToggleFreezeSpawning() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.ToggleFreezeSpawning()
}

func (target *Control) ToggleFreezeCollisions() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.ToggleFreezeCollisions()
}

func (target *Control) PlayerInvincible(playerID string) (bool, bool) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
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
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
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
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.lifeRuntime.InfiniteOverride(playerID)
}

func (target *Control) SetInfiniteLives(playerID string, enabled bool) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.lifeRuntime.SetInfiniteOverride(playerID, enabled)
}

func (target *Control) PlayerFrozen(playerID string) (bool, bool) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	session, ok := target.game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.Suspension.DevFrozen, true
}

func (target *Control) SetPlayerFrozen(playerID string, enabled bool) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
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
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	targetPlayer, ok := target.game.entities.Players[targetPlayerID]
	if !ok || targetPlayer == nil {
		return true
	}
	damageRequest := damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:                 sourcePlayerID,
			EntityType:               damage.EntityTypePlayer,
			Cause:                    damage.DamageCauseDebug,
			ResponsiblePlayerID:      sourcePlayerID,
			OriginalInstigatorID:     sourcePlayerID,
			BypassInvulnerability:    true,
			AuthorizedDevAdminSource: true,
		},
		Target: damage.DamageTarget{
			EntityID:     targetPlayerID,
			EntityType:   damage.EntityTypePlayer,
			Health:       targetPlayer.Health,
			MaxHealth:    targetPlayer.Stats.MaxHealth,
			Shield:       targetPlayer.Shields,
			MaxShield:    targetPlayer.Stats.MaxShields,
			Invulnerable: targetPlayer.IsInvulnerable() || !targetPlayer.DamageOptions.CanTakeDamage(),
		},
		Spec: damage.DamageSpec{
			Amount:       targetPlayer.Health,
			Type:         damage.DamageTypeKinetic,
			Cause:        damage.DamageCauseDebug,
			BypassShield: true,
		},
	}
	damageResult := target.game.resolveDamageRequest(damageRequest)
	applyDamageResultToPlayer(targetPlayer, damageResult)
	target.game.acceptCreatedDamageOverTime(damageResult)
	if damageResult.Fatal {
		attribution := lives.AttributionUnattributed
		if sourcePlayerID == targetPlayerID {
			attribution = lives.AttributionSelfDestruction
		} else if sourcePlayerID != "" {
			if _, ok := target.game.playerSessions[sourcePlayerID]; ok {
				attribution = lives.AttributionPlayerCaused
			}
		}
		input := lives.DeathInput{CauseCode: "devtools", Attribution: attribution}
		if attribution == lives.AttributionPlayerCaused {
			input.KillerPlayerID = sourcePlayerID
		}
		target.game.applyFatalPlayerDamageWithInput(targetPlayerID, targetPlayer, input)
	}
	target.game.publishPresentationFrameLocked()
	return true
}
