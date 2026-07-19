package playerdata

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

const (
	HangarInventorySchemaVersion = 1
	StarterShipID                = "v_wing"
	StarterPrimaryWeaponID       = "pulse"
	InventoryStateNormal         = "normal"
	InventoryStateReversed       = "reversed"
)

var (
	ErrInventoryConflict = errors.New("inventory version conflict")
	ErrInventoryCorrupt  = errors.New("inventory is corrupt")
)

type InventoryStore interface {
	LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error)
	StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error)
}

type InventoryLoad struct {
	Inventory           protocol.HangarInventory
	Found               bool
	Persisted           bool
	SynthesizedFallback bool
	RepairAttempted     bool
	Message             string
}

func StarterHangarInventory(identity protocol.PlayerDataIdentity) protocol.HangarInventory {
	playerRef := IdentityKey(identity)
	acquiredAt := time.Now().UTC().Format(time.RFC3339)
	shipID := stableOwnedID("s", playerRef, "starter", StarterShipID)
	weaponID := stableOwnedID("w", playerRef, "starter", StarterPrimaryWeaponID)

	return protocol.HangarInventory{
		SchemaVersion:      HangarInventorySchemaVersion,
		PlayerRef:          playerRef,
		OwnedShips:         []protocol.OwnedShip{{OwnedShipID: shipID, ShipID: StarterShipID, AcquiredAt: acquiredAt, AcquisitionRef: "starter", HardwiredEquipment: []protocol.HardwiredEquipment{}, State: InventoryStateNormal}},
		OwnedWeapons:       []protocol.OwnedWeapon{{OwnedWeaponID: weaponID, WeaponID: StarterPrimaryWeaponID, AcquiredAt: acquiredAt, AcquisitionRef: "starter", State: InventoryStateNormal}},
		OwnedModules:       []protocol.OwnedModule{},
		UnlockedContent:    []string{StarterShipID, StarterPrimaryWeaponID},
		StackableItems:     []protocol.StackableInventoryItem{},
		DefaultOwnedShipID: shipID,
		AppliedGrantIds:    []string{},
	}
}

func NormalizeHangarInventory(_ protocol.PlayerDataIdentity, inventory protocol.HangarInventory) protocol.HangarInventory {
	if inventory.OwnedShips == nil {
		inventory.OwnedShips = []protocol.OwnedShip{}
	}
	if inventory.OwnedWeapons == nil {
		inventory.OwnedWeapons = []protocol.OwnedWeapon{}
	}
	if inventory.OwnedModules == nil {
		inventory.OwnedModules = []protocol.OwnedModule{}
	}
	if inventory.UnlockedContent == nil {
		inventory.UnlockedContent = []string{}
	}
	if inventory.StackableItems == nil {
		inventory.StackableItems = []protocol.StackableInventoryItem{}
	}
	if inventory.AppliedGrantIds == nil {
		inventory.AppliedGrantIds = []string{}
	}
	inventory.UnlockedContent = uniqueSorted(inventory.UnlockedContent)
	inventory.AppliedGrantIds = uniqueSorted(inventory.AppliedGrantIds)
	for i := range inventory.OwnedShips {
		if inventory.OwnedShips[i].HardwiredEquipment == nil {
			inventory.OwnedShips[i].HardwiredEquipment = []protocol.HardwiredEquipment{}
		}
	}
	return inventory
}

func ValidateHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory) error {
	if IdentityKey(identity) == "" {
		return errors.New("invalid identity")
	}
	if inventory.SchemaVersion != HangarInventorySchemaVersion {
		return fmt.Errorf("unsupported inventory schema version %d", inventory.SchemaVersion)
	}
	if inventory.PlayerRef != IdentityKey(identity) {
		return errors.New("inventory player_ref does not match identity")
	}
	if len(inventory.OwnedShips) == 0 || len(inventory.OwnedWeapons) == 0 {
		return errors.New("inventory is not playable")
	}
	shipIDs := map[string]struct{}{}
	for _, ship := range inventory.OwnedShips {
		if ship.OwnedShipID == "" || ship.ShipID == "" {
			return errors.New("owned ship identity is required")
		}
		if _, exists := shipIDs[ship.OwnedShipID]; exists {
			return errors.New("duplicate owned ship id")
		}
		shipIDs[ship.OwnedShipID] = struct{}{}
	}
	if _, exists := shipIDs[inventory.DefaultOwnedShipID]; !exists {
		return errors.New("default owned ship is missing")
	}
	weaponIDs := map[string]struct{}{}
	for _, weapon := range inventory.OwnedWeapons {
		if weapon.OwnedWeaponID == "" || weapon.WeaponID == "" {
			return errors.New("owned weapon identity is required")
		}
		if _, exists := weaponIDs[weapon.OwnedWeaponID]; exists {
			return errors.New("duplicate owned weapon id")
		}
		weaponIDs[weapon.OwnedWeaponID] = struct{}{}
	}
	moduleIDs := map[string]struct{}{}
	for _, module := range inventory.OwnedModules {
		if module.OwnedModuleID == "" || module.ModuleID == "" {
			return errors.New("owned module identity is required")
		}
		if _, exists := moduleIDs[module.OwnedModuleID]; exists {
			return errors.New("duplicate owned module id")
		}
		moduleIDs[module.OwnedModuleID] = struct{}{}
	}
	for _, item := range inventory.StackableItems {
		if item.ItemRef == "" || item.Quantity < 0 {
			return errors.New("invalid stackable inventory item")
		}
	}
	return nil
}

func stableOwnedID(prefix string, parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%s_%x", prefix, sum[:8])
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
