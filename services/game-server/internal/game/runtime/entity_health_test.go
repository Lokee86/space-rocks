package runtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
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

func TestNewBulletFromWeaponSpawnCopiesWeaponFields(t *testing.T) {
	spawn := weapons.ProjectileSpawn{
		WeaponID:       weapons.BasicCannon,
		ProjectileType: "bullet",
		Position:       physics.Vector2{X: 10, Y: 20},
		Rotation:       0.5,
		Velocity:       physics.Vector2{X: 100, Y: 0},
		Lifetime:       1.25,
		Damage: damage.DamageSpec{
			Amount: 3,
			Type:   damage.DamageTypeKinetic,
			Cause:  damage.DamageCauseProjectile,
		},
	}

	bullet := NewBulletFromWeaponSpawn("bullet-1", "player-1", spawn)

	if bullet.ID != "bullet-1" || bullet.OwnerID != "player-1" {
		t.Fatalf("expected bullet identity to be copied, got %+v", bullet)
	}
	if bullet.WeaponID != spawn.WeaponID {
		t.Fatalf("expected weapon id %q, got %q", spawn.WeaponID, bullet.WeaponID)
	}
	if bullet.ProjectileType != spawn.ProjectileType {
		t.Fatalf("expected projectile type %q, got %q", spawn.ProjectileType, bullet.ProjectileType)
	}
	if bullet.X != spawn.Position.X || bullet.Y != spawn.Position.Y {
		t.Fatalf("expected bullet position %v, got (%v, %v)", spawn.Position, bullet.X, bullet.Y)
	}
	if bullet.Rotation != spawn.Rotation {
		t.Fatalf("expected rotation %v, got %v", spawn.Rotation, bullet.Rotation)
	}
	if bullet.Velocity != spawn.Velocity {
		t.Fatalf("expected velocity %v, got %v", spawn.Velocity, bullet.Velocity)
	}
	if bullet.Life != spawn.Lifetime {
		t.Fatalf("expected life %v, got %v", spawn.Lifetime, bullet.Life)
	}
	if bullet.Damage != spawn.Damage.Amount {
		t.Fatalf("expected damage %d, got %d", spawn.Damage.Amount, bullet.Damage)
	}
	if bullet.DamageSpec.Amount != spawn.Damage.Amount {
		t.Fatalf("expected damage amount %d, got %d", spawn.Damage.Amount, bullet.DamageSpec.Amount)
	}
	if bullet.DamageSpec.Type != spawn.Damage.Type {
		t.Fatalf("expected damage type %q, got %q", spawn.Damage.Type, bullet.DamageSpec.Type)
	}
	if bullet.DamageSpec.Cause != spawn.Damage.Cause {
		t.Fatalf("expected damage cause %q, got %q", spawn.Damage.Cause, bullet.DamageSpec.Cause)
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

