package playerbuild

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func TestResolveCompilesInventorySelectionAndHardwiredEffects(t *testing.T) {
	inventory := testInventory()
	catalog := testCatalog()
	rules := Rules{ModeID: "arcade_survival", HardwiredPolicy: HardwiredAllowed}
	options := ComputeEligibility("player-1", inventory, catalog, rules)
	selection := options.FallbackLoadout
	selection.SelectedModulesBySlot[ShieldMod] = "module-1"

	build, err := Resolve(selection, inventory, catalog, rules)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if err := ValidateResolvedBuild(build); err != nil {
		t.Fatalf("resolved build should validate: %v", err)
	}
	if build.InventoryVersion != inventory.InventoryVersion {
		t.Fatalf("expected inventory version %d, got %d", inventory.InventoryVersion, build.InventoryVersion)
	}
	if build.ShipStats.MaxShields != 25 {
		t.Fatalf("expected shield module to add 25 shields, got %d", build.ShipStats.MaxShields)
	}
	baseHealth := catalog.Ships[ShipVWing].Stats.MaxHealth
	if build.ShipStats.MaxHealth != baseHealth+10 {
		t.Fatalf("expected hardwired armor to add 10 health, got %d", build.ShipStats.MaxHealth)
	}
	if len(build.HardwiredEquipment) != 1 || !build.HardwiredEquipment[0].EffectsApplied {
		t.Fatalf("expected applied hardwired equipment, got %#v", build.HardwiredEquipment)
	}
	if build.ShieldPolicy.MaxShields != 25 || !build.ShieldPolicy.StartsFull {
		t.Fatalf("unexpected shield policy: %#v", build.ShieldPolicy)
	}
}

func TestResolveNormalizesHardwiredWithoutApplyingPower(t *testing.T) {
	inventory := testInventory()
	catalog := testCatalog()
	rules := Rules{HardwiredPolicy: HardwiredNormalized}
	options := ComputeEligibility("player-1", inventory, catalog, rules)
	build, err := Resolve(options.FallbackLoadout, inventory, catalog, rules)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if build.ShipStats.MaxHealth != catalog.Ships[ShipVWing].Stats.MaxHealth {
		t.Fatalf("normalized hardwired module should not change health, got %d", build.ShipStats.MaxHealth)
	}
	if len(build.HardwiredEquipment) != 1 || build.HardwiredEquipment[0].EffectsApplied {
		t.Fatalf("expected normalized declaration without effects, got %#v", build.HardwiredEquipment)
	}
}

func TestResolveCompilesLimitedSecondaryAmmoAndActiveModule(t *testing.T) {
	inventory := testInventory()
	inventory.OwnedWeapons = append(inventory.OwnedWeapons, OwnedWeapon{
		OwnedWeaponID: "weapon-secondary",
		WeaponID:      "torpedo_owned",
		State:         "normal",
	})
	catalog := testCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{})
	selection := options.FallbackLoadout
	selection.SelectedWeaponsByPoint[Secondary1] = "weapon-secondary"
	selection.StartingAmmoByPoint[Secondary1] = 3
	selection.SelectedModulesBySlot[UtilityMod] = "module-active"

	build, err := Resolve(selection, inventory, catalog, Rules{})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if build.PlayerArmory.Secondary.ID != weapons.Torpedo {
		t.Fatalf("expected torpedo runtime weapon, got %q", build.PlayerArmory.Secondary.ID)
	}
	if build.StartingEquipmentState.SecondaryAmmo != 3 {
		t.Fatalf("expected catalog starting ammo 3, got %d", build.StartingEquipmentState.SecondaryAmmo)
	}
	if len(build.ActiveModuleDeclarations) != 1 || build.ActiveModuleDeclarations[0] != "scanner_pulse" {
		t.Fatalf("expected active module declaration, got %#v", build.ActiveModuleDeclarations)
	}
}

func TestResolveRejectsClientStartingAmmoOverride(t *testing.T) {
	inventory := testInventory()
	inventory.OwnedWeapons = append(inventory.OwnedWeapons, OwnedWeapon{
		OwnedWeaponID: "weapon-secondary",
		WeaponID:      "torpedo_owned",
		State:         "normal",
	})
	catalog := testCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{})
	selection := options.FallbackLoadout
	selection.SelectedWeaponsByPoint[Secondary1] = "weapon-secondary"
	selection.StartingAmmoByPoint[Secondary1] = 999

	if _, err := Resolve(selection, inventory, catalog, Rules{}); err == nil {
		t.Fatal("expected client starting-ammo override rejection")
	}
}

func TestDefaultCatalogResolvesSelectableScoutTorpedoAndModules(t *testing.T) {
	inventory := Inventory{
		InventoryVersion: 2,
		OwnedShips: []OwnedShip{
			{OwnedShipID: "ship-standard", ShipID: ShipVWing, State: "normal"},
			{OwnedShipID: "ship-scout", ShipID: ShipVWingScout, State: "normal"},
		},
		OwnedWeapons: []OwnedWeapon{
			{OwnedWeaponID: "weapon-pulse", WeaponID: WeaponPulse, State: "normal"},
			{OwnedWeaponID: "weapon-torpedo", WeaponID: WeaponTorpedo, State: "normal"},
		},
		OwnedModules: []OwnedModule{
			{OwnedModuleID: "module-shield", ModuleID: ModuleShieldCapacitor, State: "normal"},
			{OwnedModuleID: "module-engine", ModuleID: ModuleEngineOverdrive, State: "normal"},
		},
		DefaultOwnedShipID: "ship-standard",
	}
	catalog := DefaultCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{})
	selection := options.FallbackLoadout
	selection.SelectedOwnedShipID = "ship-scout"
	selection.SelectedWeaponsByPoint[Secondary1] = "weapon-torpedo"
	selection.SelectedModulesBySlot[ShieldMod] = "module-shield"
	selection.SelectedModulesBySlot[EngineMod] = "module-engine"

	build, err := Resolve(selection, inventory, catalog, Rules{})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if build.ShipID != ShipVWingScout || build.PlayerArmory.Secondary.ID != weapons.Torpedo {
		t.Fatalf("unexpected selected runtime build: %#v", build)
	}
	if build.StartingEquipmentState.SecondaryAmmo != 3 {
		t.Fatalf("expected three starting torpedoes, got %d", build.StartingEquipmentState.SecondaryAmmo)
	}
	if build.ShipStats.MaxShields != 50 || build.ShieldPolicy.MaxShields != 50 {
		t.Fatalf("shield module did not reach runtime shield policy: %#v", build.ShieldPolicy)
	}
	if build.ShipStats.MaxSpeed <= catalog.Ships[ShipVWingScout].Stats.MaxSpeed {
		t.Fatal("engine module did not increase scout max speed")
	}
}

func TestResolvedBuildCloneOwnsCollections(t *testing.T) {
	build := DefaultResolvedBuild("player-1")
	clone := build.Clone()
	clone.WeaponPointLayout[Primary1] = PointNone
	delete(clone.EquippedWeapons, Primary1)
	if build.WeaponPointLayout[Primary1] != PointHardpoint {
		t.Fatal("clone mutated source weapon point layout")
	}
	if _, ok := build.EquippedWeapons[Primary1]; !ok {
		t.Fatal("clone mutated source equipped weapons")
	}
}
