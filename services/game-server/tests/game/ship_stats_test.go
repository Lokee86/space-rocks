package gametests

import (
	"math"
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/motion"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDefaultShipTypeResolvesBaselineEffectiveStats(t *testing.T) {
	stats := runtime.ResolveShipStats(runtime.DefaultShipTypeID)

	if stats.RotationSpeed != constants.PlayerRotationSpeed {
		t.Fatalf("expected effective rotation speed %v, got %v", constants.PlayerRotationSpeed, stats.RotationSpeed)
	}
	if stats.ThrustForce != constants.PlayerThrustForce {
		t.Fatalf("expected effective thrust force %v, got %v", constants.PlayerThrustForce, stats.ThrustForce)
	}
	if stats.MaxSpeed != constants.PlayerMaxSpeed {
		t.Fatalf("expected effective max speed %v, got %v", constants.PlayerMaxSpeed, stats.MaxSpeed)
	}
	if stats.Damping != constants.PlayerDamping {
		t.Fatalf("expected effective damping %v, got %v", constants.PlayerDamping, stats.Damping)
	}
	if stats.BulletCooldown != constants.BulletCooldown {
		t.Fatalf("expected effective bullet cooldown %v, got %v", constants.BulletCooldown, stats.BulletCooldown)
	}
	if stats.BulletSpeed != constants.BulletSpeed {
		t.Fatalf("expected effective bullet speed %v, got %v", constants.BulletSpeed, stats.BulletSpeed)
	}
	if stats.BulletLifetime != constants.BulletLifetime {
		t.Fatalf("expected effective bullet lifetime %v, got %v", constants.BulletLifetime, stats.BulletLifetime)
	}
	if stats.BulletSpawnOffset != constants.BulletSpawnOffset {
		t.Fatalf("expected effective bullet spawn offset %v, got %v", constants.BulletSpawnOffset, stats.BulletSpawnOffset)
	}
	if stats.CollisionShapeID != runtime.DefaultShipTypeID {
		t.Fatalf("expected effective collision shape ID %q, got %q", runtime.DefaultShipTypeID, stats.CollisionShapeID)
	}
}

func TestDefaultShipStatModifiersAreNeutral(t *testing.T) {
	modifiers := runtime.DefaultShipStatModifiers()

	if modifiers.RotationSpeed != 1.0 {
		t.Fatalf("expected rotation speed modifier 1.0, got %v", modifiers.RotationSpeed)
	}
	if modifiers.ThrustForce != 1.0 {
		t.Fatalf("expected thrust force modifier 1.0, got %v", modifiers.ThrustForce)
	}
	if modifiers.MaxSpeed != 1.0 {
		t.Fatalf("expected max speed modifier 1.0, got %v", modifiers.MaxSpeed)
	}
	if modifiers.Damping != 1.0 {
		t.Fatalf("expected damping modifier 1.0, got %v", modifiers.Damping)
	}
	if modifiers.BulletCooldown != 1.0 {
		t.Fatalf("expected bullet cooldown modifier 1.0, got %v", modifiers.BulletCooldown)
	}
	if modifiers.BulletSpeed != 1.0 {
		t.Fatalf("expected bullet speed modifier 1.0, got %v", modifiers.BulletSpeed)
	}
	if modifiers.BulletLifetime != 1.0 {
		t.Fatalf("expected bullet lifetime modifier 1.0, got %v", modifiers.BulletLifetime)
	}
	if modifiers.BulletSpawnOffset != 1.0 {
		t.Fatalf("expected bullet spawn offset modifier 1.0, got %v", modifiers.BulletSpawnOffset)
	}
	if modifiers.CollisionShapeID != "v_wing" {
		t.Fatalf("expected collision shape ID %q, got %q", "v_wing", modifiers.CollisionShapeID)
	}
}

func TestResolveShipStatModifiersFallsBackToDefaultModifiers(t *testing.T) {
	defaultModifiers := runtime.DefaultShipStatModifiers()

	if modifiers := runtime.ResolveShipStatModifiers(runtime.DefaultShipTypeID); modifiers != defaultModifiers {
		t.Fatalf("expected default ship modifiers %#v, got %#v", defaultModifiers, modifiers)
	}
	if modifiers := runtime.ResolveShipStatModifiers("unknown_ship"); modifiers != defaultModifiers {
		t.Fatalf("expected unknown ship type to resolve default modifiers %#v, got %#v", defaultModifiers, modifiers)
	}
}

func TestResolveShipStatsFallsBackToDefaultEffectiveStats(t *testing.T) {
	defaultStats := runtime.DefaultShipStats()

	if stats := runtime.ResolveShipStats(runtime.DefaultShipTypeID); stats != defaultStats {
		t.Fatalf("expected default effective stats %#v, got %#v", defaultStats, stats)
	}
	if stats := runtime.ResolveShipStats("unknown_ship"); stats != defaultStats {
		t.Fatalf("expected unknown ship type to resolve default effective stats %#v, got %#v", defaultStats, stats)
	}
}

func TestNewPlayerSessionCarriesResolvedDefaultStats(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	stats := scenario.sessionField(playerID, "Stats").Interface().(runtime.ShipStats)
	if stats != runtime.DefaultShipStats() {
		t.Fatalf("expected session default stats %#v, got %#v", runtime.DefaultShipStats(), stats)
	}
}

func TestSessionCreatedShipsCopySessionStats(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	customStats := runtime.DefaultShipStats()
	customStats.MaxSpeed = 1234
	scenario.sessionField(playerID, "Stats").Set(reflect.ValueOf(customStats))
	scenario.removePlayerEntity(playerID)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	ship := scenario.player(playerID)
	if ship.Stats != customStats {
		t.Fatalf("expected respawned ship stats %#v, got %#v", customStats, ship.Stats)
	}
}

func TestShipMovementUsesShipStatsValues(t *testing.T) {
	ship := runtime.Ship{
		Stats: runtime.ShipStats{
			RotationSpeed: 3,
			ThrustForce:   10,
			MaxSpeed:      1000,
			Damping:       1,
		},
		Input: runtime.InputState{
			Right:   true,
			Forward: true,
		},
	}

	motion.StepShip(&ship, 0.5)

	expectedRotation := 1.5
	expectedVelocity := physics.Vector2{
		X: math.Sin(expectedRotation) * 5,
		Y: -math.Cos(expectedRotation) * 5,
	}
	assertFloatNear(t, ship.Rotation, expectedRotation)
	assertFloatNear(t, ship.Velocity.X, expectedVelocity.X)
	assertFloatNear(t, ship.Velocity.Y, expectedVelocity.Y)
	assertFloatNear(t, ship.X, expectedVelocity.X*0.5)
	assertFloatNear(t, ship.Y, expectedVelocity.Y*0.5)
}

func TestBasicCannonProfileUsesBulletCooldown(t *testing.T) {
	profile, ok := weapons.Lookup(weapons.BasicCannon)
	if !ok {
		t.Fatal("expected basic cannon profile to exist")
	}
	if profile.CooldownSeconds != constants.BulletCooldown {
		t.Fatalf("expected cooldown %v, got %v", constants.BulletCooldown, profile.CooldownSeconds)
	}
}

func TestSpawnedBulletUsesShipStats(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	ship := scenario.player(playerID)
	ship.X = 100
	ship.Y = 200
	ship.Rotation = 0

	profile, ok := weapons.Lookup(weapons.BasicCannon)
	if !ok {
		t.Fatal("expected basic cannon profile to exist")
	}

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: runtime.InputState{Shoot: true},
	})
	scenario.step(0)

	bullet := scenario.bullet("bullet-1")
	if bullet.Velocity.X != 0 || bullet.Velocity.Y != -profile.Projectile.Speed {
		t.Fatalf("expected bullet velocity (0, %v), got (%v, %v)", -profile.Projectile.Speed, bullet.Velocity.X, bullet.Velocity.Y)
	}
	if bullet.Life != profile.Projectile.Lifetime {
		t.Fatalf("expected bullet lifetime %v, got %v", profile.Projectile.Lifetime, bullet.Life)
	}
	if bullet.X != ship.X || bullet.Y != ship.Y-profile.Projectile.SpawnOffset {
		t.Fatalf("expected bullet position (%v, %v), got (%v, %v)", ship.X, ship.Y-profile.Projectile.SpawnOffset, bullet.X, bullet.Y)
	}
}

func assertFloatNear(t *testing.T, got float64, expected float64) {
	t.Helper()

	if math.Abs(got-expected) > 0.000001 {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
