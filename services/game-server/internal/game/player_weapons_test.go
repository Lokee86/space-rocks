package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func TestFirePlayerPrimaryWeaponCreatesWeaponBackedBullet(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	if player == nil {
		t.Fatal("expected active player ship to exist")
	}

	player.ShipWeapons.Primary = weapons.Equipped{
		ID:         weapons.BasicCannon,
		AmmoPolicy: weapons.InfiniteAmmo,
	}

	before := len(game.entities.Projectiles)
	if !game.firePlayerPrimaryWeapon(playerID, player) {
		t.Fatal("expected primary weapon fire to succeed")
	}
	if got := len(game.entities.Projectiles); got != before+1 {
		t.Fatalf("expected 1 projectile to be added, got %d", got-before)
	}
	if player.WeaponState.Primary.CooldownRemaining <= 0 {
		t.Fatalf("expected primary cooldown to become positive, got %v", player.WeaponState.Primary.CooldownRemaining)
	}

	var projectile *runtime.Bullet
	for _, bullet := range game.entities.Projectiles {
		projectile = bullet
		break
	}
	if projectile == nil {
		t.Fatal("expected projectile to be stored")
	}
	if projectile.WeaponID != weapons.BasicCannon {
		t.Fatalf("WeaponID = %q, want %q", projectile.WeaponID, weapons.BasicCannon)
	}
	if projectile.ProjectileType != "bullet" {
		t.Fatalf("ProjectileType = %q, want %q", projectile.ProjectileType, "bullet")
	}
	if projectile.DamageSpec.Amount != constants.BasicCannonDamage {
		t.Fatalf("DamageSpec.Amount = %d, want %d", projectile.DamageSpec.Amount, constants.BasicCannonDamage)
	}
}

func TestFirePlayerPrimaryWeaponLeavesInfiniteAmmoBasicCannonIntact(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	if player == nil {
		t.Fatal("expected active player ship to exist")
	}

	player.ShipWeapons.Primary = weapons.Equipped{
		ID:         weapons.BasicCannon,
		AmmoPolicy: weapons.InfiniteAmmo,
	}
	beforePrimary := player.ShipWeapons.Primary

	if !game.firePlayerPrimaryWeapon(playerID, player) {
		t.Fatal("expected primary weapon fire to succeed")
	}

	if player.ShipWeapons.Primary != beforePrimary {
		t.Fatalf("expected primary weapon to remain %v, got %v", beforePrimary, player.ShipWeapons.Primary)
	}
	if player.WeaponState.Primary.AmmoRemaining != 0 {
		t.Fatalf("expected ammo remaining to stay zero for infinite ammo, got %d", player.WeaponState.Primary.AmmoRemaining)
	}
}

func TestStepPlayersFiresMountedSecondaryTorpedoWeapon(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	if player == nil {
		t.Fatal("expected active player ship to exist")
	}

	player.ShipWeapons.Secondary = weapons.Equipped{
		ID:         weapons.Torpedo,
		AmmoPolicy: weapons.InfiniteAmmo,
	}

	player.SetInput(runtime.InputState{
		PrimaryFire:   false,
		SecondaryFire: true,
	})

	game.Step(1.0 / float64(constants.ServerTickRate))

	if got := len(game.entities.Projectiles); got != 1 {
		t.Fatalf("expected exactly 1 projectile to be created, got %d", got)
	}

	var projectile *runtime.Bullet
	for _, bullet := range game.entities.Projectiles {
		projectile = bullet
		break
	}
	if projectile == nil {
		t.Fatal("expected projectile to be stored")
	}
	if projectile.WeaponID != weapons.Torpedo {
		t.Fatalf("WeaponID = %q, want %q", projectile.WeaponID, weapons.Torpedo)
	}
	if projectile.ProjectileType != "torpedo" {
		t.Fatalf("ProjectileType = %q, want %q", projectile.ProjectileType, "torpedo")
	}
	if player.WeaponState.Secondary.CooldownRemaining <= 0 {
		t.Fatalf("expected secondary cooldown to become positive, got %v", player.WeaponState.Secondary.CooldownRemaining)
	}
}
