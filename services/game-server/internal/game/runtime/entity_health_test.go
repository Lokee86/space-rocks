package runtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
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

func TestShipStateIncludesTargetKindAndTargetID(t *testing.T) {
	ship := &Ship{
		ID:         "player-1",
		TargetKind: "player",
		TargetID:   "player-2",
	}

	state := ship.State()

	if state.TargetKind != "player" {
		t.Fatalf("expected target kind %q, got %q", "player", state.TargetKind)
	}
	if state.TargetID != "player-2" {
		t.Fatalf("expected target id %q, got %q", "player-2", state.TargetID)
	}
}

func TestShipConstructionIncludesDamageModifiers(t *testing.T) {
	ship := &Ship{
		ID: "player-1",
		DamageModifiers: []damage.DamageModifier{
			{Type: damage.DamageTypeThermal, Category: damage.DamageModifierCategoryResistance, Operation: damage.DamageModifierOperationMultiply, Value: 0.25},
			{Type: damage.DamageTypeRadioactive, Category: damage.DamageModifierCategoryVulnerability, Operation: damage.DamageModifierOperationMultiply, Value: 1.25},
		},
	}

	if len(ship.DamageModifiers) != 2 {
		t.Fatalf("expected 2 ship damage modifiers, got %d", len(ship.DamageModifiers))
	}
	if ship.DamageModifiers[0].Type != damage.DamageTypeThermal {
		t.Fatalf("expected first modifier type %q, got %q", damage.DamageTypeThermal, ship.DamageModifiers[0].Type)
	}
	if ship.DamageModifiers[1].Type != damage.DamageTypeRadioactive {
		t.Fatalf("expected second modifier type %q, got %q", damage.DamageTypeRadioactive, ship.DamageModifiers[1].Type)
	}
}

func TestAsteroidConstructionIncludesDamageModifiers(t *testing.T) {
	asteroid := &Asteroid{
		ID: "asteroid-1",
		DamageModifiers: []damage.DamageModifier{
			{Type: damage.DamageTypeExplosive, Category: damage.DamageModifierCategoryResistance, Operation: damage.DamageModifierOperationMultiply, Value: 0.5},
			{Type: damage.DamageTypeEnergy, Category: damage.DamageModifierCategoryVulnerability, Operation: damage.DamageModifierOperationMultiply, Value: 1.5},
		},
	}

	if len(asteroid.DamageModifiers) != 2 {
		t.Fatalf("expected 2 asteroid damage modifiers, got %d", len(asteroid.DamageModifiers))
	}
	if asteroid.DamageModifiers[0].Type != damage.DamageTypeExplosive {
		t.Fatalf("expected first modifier type %q, got %q", damage.DamageTypeExplosive, asteroid.DamageModifiers[0].Type)
	}
	if asteroid.DamageModifiers[1].Type != damage.DamageTypeEnergy {
		t.Fatalf("expected second modifier type %q, got %q", damage.DamageTypeEnergy, asteroid.DamageModifiers[1].Type)
	}
}

