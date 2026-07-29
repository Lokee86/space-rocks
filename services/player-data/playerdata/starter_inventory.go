package playerdata

import (
	"time"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

const (
	StarterShipID            = "v_wing"
	StarterScoutShipID       = "v_wing_scout"
	StarterPrimaryWeaponID   = "pulse"
	StarterSecondaryWeaponID = "torpedo"
	StarterShieldModuleID    = "shield_capacitor"
	StarterArmorModuleID     = "reinforced_hull"
	StarterEngineModuleID    = "engine_overdrive"
	StarterUtilityModuleID   = "flight_stabilizer"
)

var starterCatalogIDs = []string{
	StarterShipID,
	StarterScoutShipID,
	StarterPrimaryWeaponID,
	StarterSecondaryWeaponID,
	StarterShieldModuleID,
	StarterArmorModuleID,
	StarterEngineModuleID,
	StarterUtilityModuleID,
}

func StarterHangarInventory(identity protocol.PlayerDataIdentity) protocol.HangarInventory {
	playerRef := IdentityKey(identity)
	acquiredAt := time.Now().UTC().Format(time.RFC3339)
	inventory := protocol.HangarInventory{
		SchemaVersion:   HangarInventorySchemaVersion,
		PlayerRef:       playerRef,
		OwnedShips:      []protocol.OwnedShip{},
		OwnedWeapons:    []protocol.OwnedWeapon{},
		OwnedModules:    []protocol.OwnedModule{},
		UnlockedContent: uniqueSorted(append([]string(nil), starterCatalogIDs...)),
		StackableItems:  []protocol.StackableInventoryItem{},
		AppliedGrantIds: []string{},
	}
	addStarterShip(&inventory, StarterShipID, acquiredAt)
	addStarterShip(&inventory, StarterScoutShipID, acquiredAt)
	addStarterWeapon(&inventory, StarterPrimaryWeaponID, acquiredAt)
	addStarterWeapon(&inventory, StarterSecondaryWeaponID, acquiredAt)
	addStarterModule(&inventory, StarterShieldModuleID, acquiredAt)
	addStarterModule(&inventory, StarterArmorModuleID, acquiredAt)
	addStarterModule(&inventory, StarterEngineModuleID, acquiredAt)
	addStarterModule(&inventory, StarterUtilityModuleID, acquiredAt)
	inventory.DefaultOwnedShipID = stableOwnedID("s", playerRef, "starter", StarterShipID)
	return inventory
}

func NormalizeHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory) protocol.HangarInventory {
	normalized, _ := normalizeHangarInventory(identity, inventory)
	return normalized
}

func normalizeHangarInventory(_ protocol.PlayerDataIdentity, inventory protocol.HangarInventory) (protocol.HangarInventory, bool) {
	changed := false
	if inventory.OwnedShips == nil {
		inventory.OwnedShips = []protocol.OwnedShip{}
		changed = true
	}
	if inventory.OwnedWeapons == nil {
		inventory.OwnedWeapons = []protocol.OwnedWeapon{}
		changed = true
	}
	if inventory.OwnedModules == nil {
		inventory.OwnedModules = []protocol.OwnedModule{}
		changed = true
	}
	if inventory.UnlockedContent == nil {
		inventory.UnlockedContent = []string{}
		changed = true
	}
	if inventory.StackableItems == nil {
		inventory.StackableItems = []protocol.StackableInventoryItem{}
		changed = true
	}
	if inventory.AppliedGrantIds == nil {
		inventory.AppliedGrantIds = []string{}
		changed = true
	}
	for i := range inventory.OwnedShips {
		if inventory.OwnedShips[i].HardwiredEquipment == nil {
			inventory.OwnedShips[i].HardwiredEquipment = []protocol.HardwiredEquipment{}
			changed = true
		}
	}

	acquiredAt := time.Now().UTC().Format(time.RFC3339)
	for _, shipID := range []string{StarterShipID, StarterScoutShipID} {
		if !hasShip(inventory, shipID) {
			addStarterShip(&inventory, shipID, acquiredAt)
			changed = true
		}
	}
	for _, weaponID := range []string{StarterPrimaryWeaponID, StarterSecondaryWeaponID} {
		if !hasWeapon(inventory, weaponID) {
			addStarterWeapon(&inventory, weaponID, acquiredAt)
			changed = true
		}
	}
	for _, moduleID := range []string{StarterShieldModuleID, StarterArmorModuleID, StarterEngineModuleID, StarterUtilityModuleID} {
		if !hasModule(inventory, moduleID) {
			addStarterModule(&inventory, moduleID, acquiredAt)
			changed = true
		}
	}
	if inventory.DefaultOwnedShipID == "" {
		inventory.DefaultOwnedShipID = stableOwnedID("s", inventory.PlayerRef, "starter", StarterShipID)
		changed = true
	}

	beforeUnlocked := append([]string(nil), inventory.UnlockedContent...)
	inventory.UnlockedContent = uniqueSorted(append(inventory.UnlockedContent, starterCatalogIDs...))
	changed = changed || !equalStrings(beforeUnlocked, inventory.UnlockedContent)
	beforeGrants := append([]string(nil), inventory.AppliedGrantIds...)
	inventory.AppliedGrantIds = uniqueSorted(inventory.AppliedGrantIds)
	changed = changed || !equalStrings(beforeGrants, inventory.AppliedGrantIds)
	return inventory, changed
}

func addStarterShip(inventory *protocol.HangarInventory, shipID, acquiredAt string) {
	inventory.OwnedShips = append(inventory.OwnedShips, protocol.OwnedShip{
		OwnedShipID: stableOwnedID("s", inventory.PlayerRef, "starter", shipID), ShipID: shipID,
		AcquiredAt: acquiredAt, AcquisitionRef: "starter", HardwiredEquipment: []protocol.HardwiredEquipment{}, State: InventoryStateNormal,
	})
}

func addStarterWeapon(inventory *protocol.HangarInventory, weaponID, acquiredAt string) {
	inventory.OwnedWeapons = append(inventory.OwnedWeapons, protocol.OwnedWeapon{
		OwnedWeaponID: stableOwnedID("w", inventory.PlayerRef, "starter", weaponID), WeaponID: weaponID,
		AcquiredAt: acquiredAt, AcquisitionRef: "starter", State: InventoryStateNormal,
	})
}

func addStarterModule(inventory *protocol.HangarInventory, moduleID, acquiredAt string) {
	inventory.OwnedModules = append(inventory.OwnedModules, protocol.OwnedModule{
		OwnedModuleID: stableOwnedID("m", inventory.PlayerRef, "starter", moduleID), ModuleID: moduleID,
		AcquiredAt: acquiredAt, AcquisitionRef: "starter", State: InventoryStateNormal,
	})
}

func hasShip(inventory protocol.HangarInventory, shipID string) bool {
	for _, owned := range inventory.OwnedShips {
		if owned.ShipID == shipID {
			return true
		}
	}
	return false
}

func hasWeapon(inventory protocol.HangarInventory, weaponID string) bool {
	for _, owned := range inventory.OwnedWeapons {
		if owned.WeaponID == weaponID {
			return true
		}
	}
	return false
}

func hasModule(inventory protocol.HangarInventory, moduleID string) bool {
	for _, owned := range inventory.OwnedModules {
		if owned.ModuleID == moduleID {
			return true
		}
	}
	return false
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
