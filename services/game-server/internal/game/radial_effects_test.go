package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestRadialEffectsRespectTargetFilter(t *testing.T) {
	game := New()
	game.worldSimulationOptions.SetFreezeWorld(true)

	game.entities.Asteroids["asteroid-1"] = runtime.NewAsteroid("asteroid-1", physics.Vector2{X: 5}, physics.Vector2{}, 1, 0)
	game.entities.Enemies["enemy-1"] = &runtime.Ship{
		ID:      "enemy-1",
		X:       15,
		Y:       0,
		Health:  10,
		Shields: 0,
	}
	game.entities.Players["player-1"] = &runtime.Ship{
		ID:      "player-1",
		X:       15,
		Y:       0,
		Health:  10,
		Shields: 0,
	}
	game.entities.Projectiles["projectile-1"] = runtime.NewBullet("projectile-1", "owner-1", physics.Vector2{X: 15}, 0, physics.Vector2{}, 1)

	spec := radial.Spec{
		CoverageMode:   radial.CoverageAnnularWave,
		ExpirationMode: radial.ExpirationSimultaneous,
		TargetFilter: radial.TargetFilter{
			Asteroids: true,
			Enemies:   true,
			Players:   false,
			Projectiles: false,
		},
		ZoneCount:        2,
		ZoneWidth:        10,
		ZoneSpawnSeconds: 0,
		TickSeconds:      0.1,
		TotalSeconds:     0.1,
		Damage: damage.DamageSpec{
			Amount: 1,
			Type:   damage.DamageTypeExplosive,
			Cause:  damage.DamageCauseArea,
		},
	}
	game.radialEffects.Add(radial.Effect{
		ID:      "effect-1",
		Origin:  physics.Vector2{},
		Spec:    spec,
		Zones:   []radial.Zone{{Index: 0, InnerRadius: 0, OuterRadius: 10, StartsAt: 0, ExpiresAt: 0.1, NextTickAt: 0}, {Index: 1, InnerRadius: 10, OuterRadius: 20, StartsAt: 0, ExpiresAt: 0.1, NextTickAt: 0}},
	})

	asteroidHealthBefore := game.entities.Asteroids["asteroid-1"].Health
	enemyHealthBefore := game.entities.Enemies["enemy-1"].Health
	playerHealthBefore := game.entities.Players["player-1"].Health
	projectileLifeBefore := game.entities.Projectiles["projectile-1"].Life

	game.stepRadialEffects(0.1)

	if got := game.entities.Asteroids["asteroid-1"].Health; got == asteroidHealthBefore {
		t.Fatal("expected asteroid health to change")
	}
	if got := game.entities.Enemies["enemy-1"].Health; got == enemyHealthBefore {
		t.Fatal("expected enemy health to change")
	}
	if got := game.entities.Players["player-1"].Health; got != playerHealthBefore {
		t.Fatal("expected player health to remain unchanged")
	}
	if got := game.entities.Projectiles["projectile-1"].Life; got != projectileLifeBefore {
		t.Fatal("expected projectile to remain unaffected")
	}
}
