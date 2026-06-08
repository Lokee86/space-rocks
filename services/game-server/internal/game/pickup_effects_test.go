package game

import (
	"testing"

	pickuprules "github.com/Lokee86/space-rocks/server/internal/game/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func TestApplyPickupEffectIntentLockedAddsAmmoToEquippedWeapon(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	if player == nil {
		t.Fatal("expected active player ship to exist")
	}

	player.ShipWeapons.Secondary = weapons.Equipped{
		ID:         weapons.Torpedo,
		AmmoPolicy: weapons.LimitedAmmo,
	}
	player.WeaponState.Secondary.AmmoRemaining = 2

	intent := pickuprules.EffectIntent{
		PlayerID:   playerID,
		PickupID:   "pickup-1",
		PickupType: "torpedo",
		EffectType: pickuprules.EffectTypeEquipWeapon,
		WeaponID:   weapons.Torpedo,
		Slot:       weapons.Secondary,
		Ammo:       1,
	}

	if !game.applyPickupEffectIntentLocked(intent) {
		t.Fatal("expected pickup effect application to succeed")
	}

	if player.ShipWeapons.Secondary.ID != weapons.Torpedo {
		t.Fatalf("Secondary.ID = %q, want %q", player.ShipWeapons.Secondary.ID, weapons.Torpedo)
	}

	if player.WeaponState.Secondary.AmmoRemaining != 3 {
		t.Fatalf("Secondary.AmmoRemaining = %d, want 3", player.WeaponState.Secondary.AmmoRemaining)
	}
}

func TestApplyPickupEffectIntentLockedAddsAmmoToEmptySecondaryWeapon(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	if player == nil {
		t.Fatal("expected active player ship to exist")
	}

	intent := pickuprules.EffectIntent{
		PlayerID:   playerID,
		PickupID:   "pickup-1",
		PickupType: "torpedo",
		EffectType: pickuprules.EffectTypeEquipWeapon,
		WeaponID:   weapons.Torpedo,
		Slot:       weapons.Secondary,
		Ammo:       1,
	}

	if !game.applyPickupEffectIntentLocked(intent) {
		t.Fatal("expected pickup effect application to succeed")
	}

	if player.ShipWeapons.Secondary.ID != weapons.Torpedo {
		t.Fatalf("Secondary.ID = %q, want %q", player.ShipWeapons.Secondary.ID, weapons.Torpedo)
	}

	if player.WeaponState.Secondary.AmmoRemaining != 1 {
		t.Fatalf("Secondary.AmmoRemaining = %d, want 1", player.WeaponState.Secondary.AmmoRemaining)
	}
}
