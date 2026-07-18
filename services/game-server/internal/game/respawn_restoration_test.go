package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func TestRespawnRestorationHealthAndShieldOverrides(t *testing.T) {
	session := newPlayerSession("player-1", physics.Vector2{})
	session.Stats.MaxShields = 100
	ship := session.NewShip(physics.Vector2{})
	ship.Health = 37
	ship.Shields = 23
	session.CaptureBetweenLifeState(ship)

	policy := lives.NewBaselineRestorationPolicy()
	policy.Health = lives.RestorationNone
	policy.Shields = lives.RestorationNone
	respawned := session.NewRespawnShip(physics.Vector2{}, policy)
	if respawned.Health != 37 || respawned.Shields != 23 {
		t.Fatalf("none restoration = health %d, shields %d; want 37, 23", respawned.Health, respawned.Shields)
	}

	ship = session.NewShip(physics.Vector2{})
	ship.Health = 11
	ship.Shields = 9
	session.CaptureBetweenLifeState(ship)
	policy.Health = lives.RestorationFull
	policy.Shields = lives.RestorationFull
	respawned = session.NewRespawnShip(physics.Vector2{}, policy)
	if respawned.Health != session.Stats.MaxHealth || respawned.Shields != session.Stats.MaxShields {
		t.Fatalf("full restoration = health %d, shields %d; want %d, %d", respawned.Health, respawned.Shields, session.Stats.MaxHealth, session.Stats.MaxShields)
	}
}

func TestRespawnRestorationCooldownEffectsAndLoadoutPolicies(t *testing.T) {
	session := newPlayerSession("player-1", physics.Vector2{})
	session.PlayerArmory.Secondary = weapons.Equipped{ID: weapons.Torpedo, AmmoPolicy: weapons.InfiniteAmmo}
	ship := session.NewShip(physics.Vector2{})
	ship.WeaponState.Primary = weapons.SlotState{CooldownRemaining: 7, AmmoRemaining: 4}
	ship.WeaponState.Secondary = weapons.SlotState{CooldownRemaining: 8, AmmoRemaining: 2}
	persistent := damage.DamageModifier{Operation: damage.DamageModifierOperationAdd, Value: 1, PersistsThroughDeath: true}
	temporary := damage.DamageModifier{Operation: damage.DamageModifierOperationAdd, Value: 2}
	ship.DamageModifiers = []damage.DamageModifier{persistent, temporary}
	session.CaptureBetweenLifeState(ship)

	policy := lives.NewBaselineRestorationPolicy()
	policy.ShortCooldownThreshold = 0.1
	respawned := session.NewRespawnShip(physics.Vector2{}, policy)
	if respawned.WeaponState.Primary.CooldownRemaining != 7 || respawned.WeaponState.Primary.AmmoRemaining != 4 {
		t.Fatalf("long cooldown/ammo was not preserved: %+v", respawned.WeaponState.Primary)
	}
	if len(respawned.DamageModifiers) != 1 || !respawned.DamageModifiers[0].PersistsThroughDeath {
		t.Fatalf("temporary effect filtering = %+v", respawned.DamageModifiers)
	}

	ship = session.NewShip(physics.Vector2{})
	ship.WeaponState.Primary.CooldownRemaining = 7
	session.CaptureBetweenLifeState(ship)
	policy.ShortCooldownThreshold = 10
	policy.TemporaryEffects = lives.TemporaryEffectsPersist
	policy.Loadout = lives.LoadoutReset
	respawned = session.NewRespawnShip(physics.Vector2{}, policy)
	if respawned.WeaponState.Primary.CooldownRemaining != 0 {
		t.Fatalf("short cooldown was not reset: %+v", respawned.WeaponState.Primary)
	}
	if respawned.ShipWeapons.Secondary != weapons.EmptyEquipped() {
		t.Fatalf("loadout reset retained secondary weapon: %+v", respawned.ShipWeapons.Secondary)
	}
	if session.PlayerArmory.Secondary != weapons.EmptyEquipped() {
		t.Fatalf("loadout reset left durable secondary weapon: %+v", session.PlayerArmory.Secondary)
	}
}
