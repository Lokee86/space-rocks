package playerbuild

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"

func testInventory() Inventory {
	return Inventory{
		InventoryVersion: 7,
		OwnedShips: []OwnedShip{
			{
				OwnedShipID: "ship-1",
				ShipID:      ShipVWing,
				State:       "normal",
				HardwiredEquipment: []HardwiredEquipment{
					{HardwiredID: "hardwired-1", EquipmentID: "armor_plate", State: "normal"},
				},
			},
		},
		OwnedWeapons: []OwnedWeapon{
			{OwnedWeaponID: "weapon-1", WeaponID: WeaponPulse, State: "normal"},
			{OwnedWeaponID: "weapon-unknown", WeaponID: "unknown", State: "normal"},
		},
		OwnedModules: []OwnedModule{
			{OwnedModuleID: "module-1", ModuleID: "shield_booster", State: "normal"},
			{OwnedModuleID: "module-active", ModuleID: "active_scanner", State: "normal"},
		},
		DefaultOwnedShipID: "ship-1",
	}
}

func testCatalog() Catalog {
	catalog := DefaultCatalog()
	catalog.Modules = map[string]ModuleProfile{
		"shield_booster": {
			ID:         "shield_booster",
			Slot:       ShieldMod,
			Class:      "defense",
			Activation: ModulePassive,
			Adjustment: ShipStatAdjustment{MaxShieldsDelta: 25},
		},
		"armor_plate": {
			ID:         "armor_plate",
			Slot:       ArmorMod,
			Class:      "defense",
			Activation: ModulePassive,
			Adjustment: ShipStatAdjustment{MaxHealthDelta: 10},
		},
		"active_scanner": {
			ID:         "active_scanner",
			Slot:       UtilityMod,
			Class:      "utility",
			Activation: ModuleActive,
			BehaviorID: "scanner_pulse",
		},
	}
	catalog.Weapons["torpedo_owned"] = WeaponProfile{
		ID:              "torpedo_owned",
		RuntimeID:       weapons.Torpedo,
		Slot:            weapons.Secondary,
		Size:            WeaponStandard,
		DeliveryClass:   "missile",
		TargetingPolicy: "skill_shot",
		EffectFlags:     []EffectFlag{"direct", "area"},
		AmmoPolicy:      weapons.LimitedAmmo,
		StartingAmmo:    3,
	}
	return catalog
}
