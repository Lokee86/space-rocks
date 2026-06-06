package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func TestRadialProjectileCreatesRadialEffectOnAsteroidHit(t *testing.T) {
	game := New()

	game.entities.Players["player-1"] = &runtime.Ship{
		ID:      "player-1",
		Health:  10,
		Shields: 0,
	}
	game.entities.Asteroids["asteroid-1"] = runtime.NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	game.entities.Projectiles["bullet-1"] = &runtime.Bullet{
		ID:             "bullet-1",
		OwnerID:        "player-1",
		WeaponID:       weapons.Torpedo,
		ProjectileType: "torpedo",
		ImpactEffect: weapons.ImpactEffectSpec{
			Kind: weapons.ImpactEffectRadial,
			Radial: radial.Spec{
				CoverageMode:   radial.CoverageAnnularWave,
				ExpirationMode: radial.ExpirationSimultaneous,
				ZoneCount:      4,
				ZoneWidth:      10,
				TargetFilter: radial.TargetFilter{
					Asteroids: true,
					Enemies:   true,
				},
			},
		},
		X: 0,
		Y: 0,
	}

	before := game.radialEffects.Len()

	game.handleBulletAsteroidCollisions()

	if got, want := game.radialEffects.Len(), before+1; got != want {
		t.Fatalf("radial effect count = %d, want %d", got, want)
	}
}
