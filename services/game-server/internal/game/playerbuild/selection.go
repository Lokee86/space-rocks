package playerbuild

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func FallbackSelection(inventory Inventory, options EligibleBuildOptions) LoadoutSelection {
	selection := LoadoutSelection{
		PlayerID:               options.PlayerID,
		ModeID:                 options.ModeID,
		SelectedWeaponsByPoint: map[WeaponPoint]string{},
		SelectedModulesBySlot:  map[ModuleSlot]string{},
		StartingAmmoByPoint:    map[WeaponPoint]int{},
	}

	for _, ship := range options.EligibleShips {
		if ship.OwnedShipID == inventory.DefaultOwnedShipID {
			selection.SelectedOwnedShipID = ship.OwnedShipID
			break
		}
	}
	if selection.SelectedOwnedShipID == "" && len(options.EligibleShips) > 0 {
		selection.SelectedOwnedShipID = options.EligibleShips[0].OwnedShipID
	}
	primaryOptions := options.WeaponsByPoint[Primary1]
	if selectedShip, ok := eligibleShip(options, selection.SelectedOwnedShipID); ok {
		for _, weapon := range primaryOptions {
			if weapon.WeaponID == selectedShip.DefaultPrimaryWeaponID {
				selection.SelectedWeaponsByPoint[Primary1] = weapon.OwnedWeaponID
				break
			}
		}
	}
	if selection.SelectedWeaponsByPoint[Primary1] == "" && len(primaryOptions) > 0 {
		selection.SelectedWeaponsByPoint[Primary1] = primaryOptions[0].OwnedWeaponID
	}
	return selection
}

func ValidateSelection(selection LoadoutSelection, inventory Inventory, catalog Catalog, options EligibleBuildOptions) error {
	if selection.PlayerID == "" || selection.PlayerID != options.PlayerID {
		return fmt.Errorf("loadout player ID does not match eligible options")
	}
	if selection.ModeID != options.ModeID {
		return fmt.Errorf("loadout mode ID does not match eligible options")
	}
	ship, ok := eligibleShip(options, selection.SelectedOwnedShipID)
	if !ok {
		return fmt.Errorf("selected ship is not eligible")
	}
	variant := catalog.Ships[ship.ShipID]
	if selection.SelectedWeaponsByPoint[Primary1] == "" {
		return fmt.Errorf("primary_1 weapon is required")
	}

	usedWeapons := map[string]struct{}{}
	for point, ownedID := range selection.SelectedWeaponsByPoint {
		if variant.WeaponPoints[point] != PointHardpoint {
			return fmt.Errorf("weapon point %q is not a selectable hardpoint", point)
		}
		option, ok := eligibleWeapon(options, point, ownedID)
		if !ok {
			return fmt.Errorf("selected weapon for %q is not eligible", point)
		}
		if _, exists := usedWeapons[ownedID]; exists {
			return fmt.Errorf("owned weapon %q cannot fill multiple points", ownedID)
		}
		usedWeapons[ownedID] = struct{}{}
		profile := catalog.Weapons[option.WeaponID]
		if pointSlot(point) != profile.Slot {
			return fmt.Errorf("weapon %q is incompatible with point %q", option.WeaponID, point)
		}
		if ammo, supplied := selection.StartingAmmoByPoint[point]; supplied && ammo != profile.StartingAmmo {
			return fmt.Errorf("starting ammo for %q is server controlled", point)
		}
	}
	for point := range selection.StartingAmmoByPoint {
		if _, selected := selection.SelectedWeaponsByPoint[point]; !selected {
			return fmt.Errorf("starting ammo references unselected weapon point %q", point)
		}
	}

	usedModules := map[string]struct{}{}
	for slot, ownedID := range selection.SelectedModulesBySlot {
		if !containsComparable(variant.ModuleSlots, slot) {
			return fmt.Errorf("module slot %q is unavailable on selected ship", slot)
		}
		if _, ok := eligibleModule(options, slot, ownedID); !ok {
			return fmt.Errorf("selected module for %q is not eligible", slot)
		}
		if _, exists := usedModules[ownedID]; exists {
			return fmt.Errorf("owned module %q cannot fill multiple slots", ownedID)
		}
		usedModules[ownedID] = struct{}{}
	}

	if _, ok := findOwnedShip(inventory, selection.SelectedOwnedShipID); !ok {
		return fmt.Errorf("selected ship does not exist in inventory")
	}
	return nil
}

func eligibleShip(options EligibleBuildOptions, ownedID string) (EligibleShipOption, bool) {
	for _, option := range options.EligibleShips {
		if option.OwnedShipID == ownedID {
			return option, true
		}
	}
	return EligibleShipOption{}, false
}

func eligibleWeapon(options EligibleBuildOptions, point WeaponPoint, ownedID string) (EligibleWeaponOption, bool) {
	for _, option := range options.WeaponsByPoint[point] {
		if option.OwnedWeaponID == ownedID {
			return option, true
		}
	}
	return EligibleWeaponOption{}, false
}

func eligibleModule(options EligibleBuildOptions, slot ModuleSlot, ownedID string) (EligibleModuleOption, bool) {
	for _, option := range options.ModulesBySlot[slot] {
		if option.OwnedModuleID == ownedID {
			return option, true
		}
	}
	return EligibleModuleOption{}, false
}

func findOwnedShip(inventory Inventory, ownedID string) (OwnedShip, bool) {
	for _, owned := range inventory.OwnedShips {
		if owned.OwnedShipID == ownedID {
			return owned, true
		}
	}
	return OwnedShip{}, false
}

func pointSlot(point WeaponPoint) weapons.Slot {
	switch point {
	case Primary1, Primary2:
		return weapons.Primary
	case Secondary1, Secondary2:
		return weapons.Secondary
	default:
		return ""
	}
}
