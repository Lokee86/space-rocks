package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func (game *Game) damageOverTime() *damage.DamageOverTimeRuntime {
	if game.damageOverTimeRuntime == nil {
		game.damageOverTimeRuntime = damage.NewDamageOverTimeRuntime()
	}
	return game.damageOverTimeRuntime
}

func (game *Game) acceptCreatedDamageOverTime(result damage.DamageResult) {
	for _, effect := range result.CreatedDamageOverTime {
		outcome := game.damageOverTime().Add(effect)
		if outcome.Added {
			game.recordDomainEvent(damageOverTimeStartedEvent(effect))
		}
	}
}

func (game *Game) stepDamageOverTime(delta float64) {
	ticks := game.damageOverTime().Step(delta, game.worldSimulationOptions.IsWorldFrozen())
	for _, tick := range ticks {
		target, position, exists := game.damageOverTimeTarget(tick.Effect.Target)
		if !exists {
			game.damageOverTime().RemoveTarget(tick.Effect.Target.EntityID)
			continue
		}
		result := game.resolveDamageRequest(damage.DamageResolutionRequest{
			Source: tick.Effect.Source,
			Target: target,
			Spec: damage.DamageSpec{
				Amount: tick.Effect.AmountPerTick,
				Type:   tick.Effect.Type,
				Cause:  damage.DamageCauseDot,
			},
			Modifiers: tick.Effect.Modifiers,
		})
		game.applyDamageOverTimeResult(tick.Effect, result, position)
	}
}

func (game *Game) damageOverTimeTarget(reference damage.DamageTargetRef) (damage.DamageTarget, physics.Vector2, bool) {
	switch reference.EntityType {
	case damage.EntityTypePlayer:
		player := game.entities.Players[reference.EntityID]
		if player == nil || player.IsPendingDespawn() {
			return damage.DamageTarget{}, physics.Vector2{}, false
		}
		return damage.DamageTarget{
			EntityID:     player.ID,
			EntityType:   damage.EntityTypePlayer,
			Health:       player.Health,
			MaxHealth:    player.Stats.MaxHealth,
			Shield:       player.Shields,
			MaxShield:    player.Stats.MaxShields,
			Invulnerable: player.IsInvulnerable() || !player.DamageOptions.CanTakeDamage(),
			Modifiers:    player.DamageModifiers,
		}, player.Position(), true
	case damage.EntityTypeAsteroid:
		asteroid := game.entities.Asteroids[reference.EntityID]
		if asteroid == nil || asteroid.IsPendingDespawn() {
			return damage.DamageTarget{}, physics.Vector2{}, false
		}
		return damage.DamageTarget{
			EntityID:   asteroid.ID,
			EntityType: damage.EntityTypeAsteroid,
			Health:     asteroid.Health,
			MaxHealth:  asteroid.Health,
			Modifiers:  asteroid.DamageModifiers,
		}, asteroid.Position(), true
	case damage.EntityTypeEnemy:
		enemy := game.entities.Enemies[reference.EntityID]
		if enemy == nil || enemy.IsPendingDespawn() {
			return damage.DamageTarget{}, physics.Vector2{}, false
		}
		return damage.DamageTarget{
			EntityID:     enemy.ID,
			EntityType:   damage.EntityTypeEnemy,
			Health:       enemy.Health,
			MaxHealth:    enemy.Stats.MaxHealth,
			Shield:       enemy.Shields,
			MaxShield:    enemy.Stats.MaxShields,
			Invulnerable: enemy.IsInvulnerable() || !enemy.DamageOptions.CanTakeDamage(),
			Modifiers:    enemy.DamageModifiers,
		}, enemy.Position(), true
	default:
		return damage.DamageTarget{}, physics.Vector2{}, false
	}
}

func (game *Game) applyDamageOverTimeResult(effect damage.ActiveDamageOverTime, result damage.DamageResult, position physics.Vector2) {
	switch effect.Target.EntityType {
	case damage.EntityTypePlayer:
		player := game.entities.Players[effect.Target.EntityID]
		if player == nil {
			return
		}
		applyDamageResultToPlayer(player, result)
		if event, ok := damageOverTimeTickEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		} else if event, ok := damageResultEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		}
		if result.Fatal {
			attribution := lives.AttributionUnattributed
			if effect.Source.ResponsiblePlayerID == player.ID {
				attribution = lives.AttributionSelfDestruction
			} else if effect.Source.ResponsiblePlayerID != "" {
				attribution = lives.AttributionPlayerCaused
			}
			input := lives.DeathInput{CauseCode: "dot", Attribution: attribution}
			if attribution == lives.AttributionPlayerCaused {
				input.KillerPlayerID = effect.Source.ResponsiblePlayerID
			}
			game.applyFatalPlayerDamageWithInput(player.ID, player, input)
		}
	case damage.EntityTypeAsteroid:
		asteroid := game.entities.Asteroids[effect.Target.EntityID]
		if asteroid == nil {
			return
		}
		applyDamageResultToAsteroid(asteroid, result)
		if event, ok := damageOverTimeTickEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		} else if event, ok := damageResultEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		}
		if result.Destroyed {
			game.applyProjectileAsteroidDestruction(effect.Source.ResponsiblePlayerID, asteroid)
		}
	case damage.EntityTypeEnemy:
		enemy := game.entities.Enemies[effect.Target.EntityID]
		if enemy == nil {
			return
		}
		applyDamageResultToEnemy(enemy, result)
		if event, ok := damageOverTimeTickEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		} else if event, ok := damageResultEvent(result, position.X, position.Y); ok {
			game.recordDomainEvent(event)
		}
	}
}
