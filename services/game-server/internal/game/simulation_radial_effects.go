package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func (game *Game) stepRadialEffects(delta float64) {
	if game.radialEffects.Len() == 0 {
		return
	}

	candidates := game.radialCandidates()
	expiredEffectIDs := make([]string, 0)
	for id, effect := range game.radialEffects.All() {
		if effect == nil {
			expiredEffectIDs = append(expiredEffectIDs, id)
			continue
		}

		result := radial.Step(effect, delta, candidates)
		for _, hit := range result.Hits {
			game.applyRadialHit(hit)
		}
		if result.Expired {
			expiredEffectIDs = append(expiredEffectIDs, id)
		}
	}

	for _, id := range expiredEffectIDs {
		game.radialEffects.Remove(id)
	}
}

func (game *Game) applyRadialHit(hit radial.Hit) {
	switch hit.TargetKind {
	case radial.TargetAsteroid:
		asteroid, ok := game.entities.Asteroids[hit.TargetID]
		if !ok || asteroid == nil || asteroid.IsPendingDespawn() {
			return
		}
		game.applyRadialHitToAsteroid(hit, asteroid)
	case radial.TargetEnemy:
		enemy, ok := game.entities.Enemies[hit.TargetID]
		if !ok || enemy == nil || enemy.IsPendingDespawn() {
			return
		}
		game.applyRadialHitToEnemy(hit, enemy)
	case radial.TargetPlayer:
		player, ok := game.entities.Players[hit.TargetID]
		if !ok || player == nil || player.IsPendingDespawn() {
			return
		}
		game.applyRadialHitToPlayer(hit, player)
	}
}

func (game *Game) applyRadialHitToAsteroid(hit radial.Hit, asteroid *runtime.Asteroid) {
	damageResult := game.resolveDamageRequest(radialDamageRequestFromHitAndAsteroid(hit, asteroid))
	applyDamageResultToAsteroid(asteroid, damageResult)
	if event, ok := damageResultEvent(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	game.acceptCreatedDamageOverTime(damageResult)
	if damageResult.Destroyed {
		game.applyProjectileAsteroidDestruction(hit.SourcePlayerID, asteroid)
	}
}

func (game *Game) applyRadialHitToEnemy(hit radial.Hit, enemy *runtime.Ship) {
	damageResult := game.resolveDamageRequest(radialDamageRequestFromHitAndEnemy(hit, enemy))
	applyDamageResultToEnemy(enemy, damageResult)
	if event, ok := damageResultEvent(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	game.acceptCreatedDamageOverTime(damageResult)
	if damageResult.Destroyed {
		// Enemy death consequences are not wired yet.
	}
}

func (game *Game) applyRadialHitToPlayer(hit radial.Hit, player *runtime.Ship) {
	damageResult := game.resolveDamageRequest(radialDamageRequestFromHitAndPlayer(hit, player))
	applyDamageResultToPlayer(player, damageResult)
	if event, ok := damageAppliedEventForResult(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	game.acceptCreatedDamageOverTime(damageResult)
	if damageResult.Fatal {
		attribution := lives.AttributionUnattributed
		if hit.SourcePlayerID == player.ID {
			attribution = lives.AttributionSelfDestruction
		} else if hit.SourcePlayerID != "" {
			if _, ok := game.playerSessions[hit.SourcePlayerID]; ok {
				attribution = lives.AttributionPlayerCaused
			}
		}
		input := lives.DeathInput{CauseCode: "radial_effect", Attribution: attribution}
		if attribution == lives.AttributionPlayerCaused {
			input.KillerPlayerID = hit.SourcePlayerID
		}
		game.applyFatalPlayerDamageWithInput(player.ID, player, input)
	}
}
