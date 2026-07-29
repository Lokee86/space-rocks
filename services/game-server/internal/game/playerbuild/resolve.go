package playerbuild

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"

func Resolve(selection LoadoutSelection, inventory Inventory, catalog Catalog, rawRules Rules) (ResolvedPlayerBuild, error) {
	if err := catalog.Validate(); err != nil {
		return ResolvedPlayerBuild{}, err
	}
	if err := ValidateRules(rawRules); err != nil {
		return ResolvedPlayerBuild{}, err
	}
	rules := NormalizeRules(rawRules)
	options := ComputeEligibility(selection.PlayerID, inventory, catalog, rules)
	if err := ValidateSelection(selection, inventory, catalog, options); err != nil {
		return ResolvedPlayerBuild{}, err
	}

	shipOption, _ := eligibleShip(options, selection.SelectedOwnedShipID)
	ownedShip, _ := findOwnedShip(inventory, selection.SelectedOwnedShipID)
	variant := catalog.Ships[shipOption.ShipID]
	build := ResolvedPlayerBuild{
		PlayerID:            selection.PlayerID,
		ModeID:              selection.ModeID,
		InventoryVersion:    inventory.InventoryVersion,
		SelectedOwnedShipID: selection.SelectedOwnedShipID,
		ShipID:              variant.ID,
		WeightClass:         variant.WeightClass,
		ShipStats:           variant.Stats,
		WeaponPointLayout:   clonePointLayout(variant.WeaponPoints),
		EquippedWeapons:     map[WeaponPoint]ResolvedWeapon{},
		EquippedModules:     map[ModuleSlot]ResolvedModule{},
		PlayerArmory:        weapons.PlayerArmory{},
	}

	for point, ownedID := range selection.SelectedWeaponsByPoint {
		option, _ := eligibleWeapon(options, point, ownedID)
		profile := catalog.Weapons[option.WeaponID]
		resolved := ResolvedWeapon{
			OwnedWeaponID: ownedID,
			CatalogID:     profile.ID,
			RuntimeID:     profile.RuntimeID,
			Point:         point,
			AmmoPolicy:    profile.AmmoPolicy,
			StartingAmmo:  profile.StartingAmmo,
		}
		build.EquippedWeapons[point] = resolved
		applyWeaponToRuntime(&build, resolved)
	}

	for slot, ownedID := range selection.SelectedModulesBySlot {
		option, _ := eligibleModule(options, slot, ownedID)
		profile := catalog.Modules[option.ModuleID]
		resolved := ResolvedModule{
			OwnedModuleID:  ownedID,
			CatalogID:      profile.ID,
			Slot:           profile.Slot,
			Activation:     profile.Activation,
			BehaviorID:     profile.BehaviorID,
			EffectsApplied: true,
		}
		build.EquippedModules[slot] = resolved
		applyModule(&build, profile, false)
	}

	resolveHardwired(&build, ownedShip.HardwiredEquipment, catalog, rules.HardwiredPolicy)
	build.ShieldPolicy = ShieldPolicy{MaxShields: build.ShipStats.MaxShields, StartsFull: true}
	return build.Clone(), nil
}

func DefaultResolvedBuild(playerID string) ResolvedPlayerBuild {
	catalog := DefaultCatalog()
	inventory := Inventory{
		InventoryVersion:   0,
		OwnedShips:         []OwnedShip{{OwnedShipID: "runtime_default_ship", ShipID: ShipVWing, State: "normal"}},
		OwnedWeapons:       []OwnedWeapon{{OwnedWeaponID: "runtime_default_weapon", WeaponID: WeaponPulse, State: "normal"}},
		DefaultOwnedShipID: "runtime_default_ship",
	}
	options := ComputeEligibility(playerID, inventory, catalog, Rules{})
	build, err := Resolve(options.FallbackLoadout, inventory, catalog, Rules{})
	if err != nil {
		panic(err)
	}
	return build
}

func NewRuntimeEquipmentState(build ResolvedPlayerBuild) weapons.State {
	return weapons.State{
		Primary:   weapons.SlotState{AmmoRemaining: build.StartingEquipmentState.PrimaryAmmo},
		Secondary: weapons.SlotState{AmmoRemaining: build.StartingEquipmentState.SecondaryAmmo},
	}
}
