package playerdata

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

const (
	HangarInventorySchemaVersion = 1
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
