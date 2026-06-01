package entities

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestResolveShipStatsDefaultTypeSetsMaxHealthFromConstants(t *testing.T) {
	stats := ResolveShipStats(DefaultShipTypeID)
	if stats.MaxHealth != constants.PlayerMaxHealth {
		t.Fatalf("expected max health %d, got %d", constants.PlayerMaxHealth, stats.MaxHealth)
	}
}

func TestResolveShipStatsDefaultTypeSetsBulletDamageFromConstants(t *testing.T) {
	stats := ResolveShipStats(DefaultShipTypeID)
	if stats.BulletDamage != constants.BulletDamage {
		t.Fatalf("expected bullet damage %d, got %d", constants.BulletDamage, stats.BulletDamage)
	}
}

func TestNewAsteroidInitializesHealthFromConstants(t *testing.T) {
	asteroid := NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	if asteroid.Health != constants.AsteroidHealth {
		t.Fatalf("expected asteroid health %d, got %d", constants.AsteroidHealth, asteroid.Health)
	}
}

func TestNewAsteroidInitializesCollisionDamageFromConstants(t *testing.T) {
	asteroid := NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	if asteroid.CollisionDamage != constants.AsteroidCollisionDamage {
		t.Fatalf("expected asteroid collision damage %d, got %d", constants.AsteroidCollisionDamage, asteroid.CollisionDamage)
	}
}

func TestNewBulletInitializesDamageFromConstants(t *testing.T) {
	bullet := NewBullet("bullet-1", "player-1", physics.Vector2{}, 0, physics.Vector2{}, 1.0)
	if bullet.Damage != constants.BulletDamage {
		t.Fatalf("expected bullet damage %d, got %d", constants.BulletDamage, bullet.Damage)
	}
}

func TestShipStateIncludesHealthAndShields(t *testing.T) {
	ship := &Ship{
		ID:      "player-1",
		Score:   1200,
		Lives:   2,
		Health:  75,
		Shields: 30,
	}

	state := ship.State()

	if state.Health != ship.Health {
		t.Fatalf("expected health %d, got %d", ship.Health, state.Health)
	}
	if state.Shields != ship.Shields {
		t.Fatalf("expected shields %d, got %d", ship.Shields, state.Shields)
	}
}
