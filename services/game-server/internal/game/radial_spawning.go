package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func (game *Game) spawnRadialEffectFromBullet(bullet *runtime.Bullet, sourcePlayerID string, impactPosition physics.Vector2) {
	if bullet == nil || bullet.ImpactEffect.Kind != weapons.ImpactEffectRadial {
		return
	}

	effectID := game.spawner.NextBulletID() + "-radial"
	game.radialEffects.Add(radial.NewEffect(radial.SpawnRequest{
		ID:             effectID,
		SourceID:       bullet.ID,
		SourcePlayerID: sourcePlayerID,
		Origin:         impactPosition,
		Spec:           bullet.ImpactEffect.Radial,
	}))
	game.recordDomainEvent(events.Event{
		Type:       events.EventRadialEffectStarted,
		SourceID:   effectID,
		EffectType: string(bullet.ImpactEffect.Kind),
		X:          impactPosition.X,
		Y:          impactPosition.Y,
	})
}
