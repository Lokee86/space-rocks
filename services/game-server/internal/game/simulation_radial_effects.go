package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
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
	damageResult := damage.ResolveSingle(radialDamageRequestFromHitAndAsteroid(hit, asteroid))
	applyDamageResultToAsteroid(asteroid, damageResult)
	if event, ok := damageAppliedEventForResult(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	if damageResult.Destroyed {
		game.applyProjectileAsteroidDestruction(hit.SourcePlayerID, asteroid)
	}
}

func (game *Game) applyRadialHitToEnemy(hit radial.Hit, enemy *runtime.Ship) {
	damageResult := damage.ResolveSingle(radialDamageRequestFromHitAndEnemy(hit, enemy))
	applyDamageResultToEnemy(enemy, damageResult)
	if event, ok := damageAppliedEventForResult(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	if damageResult.Destroyed {
		// Enemy death consequences are not wired yet.
	}
}

func (game *Game) applyRadialHitToPlayer(hit radial.Hit, player *runtime.Ship) {
	damageResult := damage.ResolveSingle(radialDamageRequestFromHitAndPlayer(hit, player))
	applyDamageResultToPlayer(player, damageResult)
	if event, ok := damageAppliedEventForResult(damageResult, hit.TargetPosition.X, hit.TargetPosition.Y); ok {
		game.recordDomainEvent(event)
	}
	if damageResult.Fatal {
		game.applyFatalPlayerDamage(player.ID, player)
	}
}
