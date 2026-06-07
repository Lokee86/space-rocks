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

func TestTorpedoImpactDamageZeroStillDestroysAsteroidWithRadialDamage(t *testing.T) {
	profile, ok := weapons.Lookup(weapons.Torpedo)
	if !ok {
		t.Fatal("expected torpedo weapon profile")
	}
	if profile.Damage.Amount != 0 {
		t.Fatalf("torpedo impact damage = %d, want 0", profile.Damage.Amount)
	}
	if profile.ImpactEffect.Kind != weapons.ImpactEffectRadial {
		t.Fatalf("torpedo impact effect kind = %q, want %q", profile.ImpactEffect.Kind, weapons.ImpactEffectRadial)
	}
	if profile.ImpactEffect.Radial.Damage.Amount <= 0 {
		t.Fatalf("torpedo radial damage = %d, want > 0", profile.ImpactEffect.Radial.Damage.Amount)
	}

	game := New()

	game.entities.Players["player-1"] = &runtime.Ship{
		ID:      "player-1",
		Health:  10,
		Shields: 0,
	}
	asteroid := runtime.NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	game.entities.Asteroids[asteroid.ID] = asteroid
	game.entities.Projectiles["torpedo-1"] = &runtime.Bullet{
		ID:             "torpedo-1",
		OwnerID:        "player-1",
		WeaponID:       weapons.Torpedo,
		ProjectileType: profile.Projectile.Type,
		ImpactEffect:   profile.ImpactEffect,
		Damage:         profile.Damage.Amount,
		DamageSpec:     profile.Damage,
		X:              0,
		Y:              0,
	}

	game.handleBulletAsteroidCollisions()

	if asteroid.Health <= 0 || asteroid.IsPendingDespawn() {
		t.Fatalf("asteroid destroyed by torpedo impact damage before radial tick: health=%d pending=%v", asteroid.Health, asteroid.IsPendingDespawn())
	}
	if game.radialEffects.Len() == 0 {
		t.Fatal("expected torpedo impact to spawn a radial effect")
	}

	for i := 0; i < profile.ImpactEffect.Radial.ZoneCount; i++ {
		game.stepRadialEffects(profile.ImpactEffect.Radial.TickSeconds)
	}

	if asteroid.Health > 0 && !asteroid.IsPendingDespawn() {
		t.Fatalf("asteroid survived torpedo radial damage: health=%d pending=%v", asteroid.Health, asteroid.IsPendingDespawn())
	}
}
