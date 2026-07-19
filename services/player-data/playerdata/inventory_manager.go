package playerdata

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type InventoryManager struct {
	store InventoryStore
}

func NewInventoryManager(store InventoryStore) *InventoryManager {
	return &InventoryManager{store: store}
}

func (m *InventoryManager) Load(identity protocol.PlayerDataIdentity) (InventoryLoad, error) {
	if m == nil || m.store == nil {
		return InventoryLoad{}, errors.New("inventory store is required")
	}
	if IdentityKey(identity) == "" {
		return InventoryLoad{}, errors.New("invalid identity")
	}

	inventory, found, err := m.store.LoadHangarInventory(identity)
	if err != nil {
		fallback := StarterHangarInventory(identity)
		load := InventoryLoad{Inventory: fallback, Found: true, SynthesizedFallback: true, Message: err.Error()}
		if errors.Is(err, ErrInventoryCorrupt) {
			load.RepairAttempted = true
			stored, storeErr := m.store.StoreHangarInventory(identity, fallback, -1)
			if storeErr == nil {
				load.Inventory, load.Persisted = stored, true
			}
		}
		return load, nil
	}

	if !found {
		starter := StarterHangarInventory(identity)
		stored, storeErr := m.store.StoreHangarInventory(identity, starter, 0)
		if storeErr == nil {
			return InventoryLoad{Inventory: stored, Found: true, Persisted: true}, nil
		}
		if errors.Is(storeErr, ErrInventoryConflict) {
			return m.Load(identity)
		}
		return InventoryLoad{Inventory: starter, Found: true, SynthesizedFallback: true, RepairAttempted: true, Message: storeErr.Error()}, nil
	}

	inventory = NormalizeHangarInventory(identity, inventory)
	if err := ValidateHangarInventory(identity, inventory); err != nil {
		fallback := StarterHangarInventory(identity)
		stored, storeErr := m.store.StoreHangarInventory(identity, fallback, inventory.InventoryVersion)
		if storeErr == nil {
			return InventoryLoad{Inventory: stored, Found: true, Persisted: true, SynthesizedFallback: true, RepairAttempted: true, Message: err.Error()}, nil
		}
		return InventoryLoad{Inventory: fallback, Found: true, SynthesizedFallback: true, RepairAttempted: true, Message: err.Error()}, nil
	}
	return InventoryLoad{Inventory: inventory, Found: true, Persisted: true}, nil
}

func (m *InventoryManager) ApplyGrant(command protocol.PlayerDataApplyInventoryGrant) (protocol.HangarInventory, bool, error) {
	if m == nil || m.store == nil {
		return protocol.HangarInventory{}, false, errors.New("inventory store is required")
	}
	if err := validateInventoryGrant(command); err != nil {
		return protocol.HangarInventory{}, false, err
	}

	for attempt := 0; attempt < 4; attempt++ {
		inventory, found, err := m.store.LoadHangarInventory(command.Identity)
		if err != nil {
			return protocol.HangarInventory{}, false, err
		}
		if !found {
			inventory = StarterHangarInventory(command.Identity)
		}
		inventory = NormalizeHangarInventory(command.Identity, inventory)
		if found {
			if err := ValidateHangarInventory(command.Identity, inventory); err != nil {
				return protocol.HangarInventory{}, false, ErrInventoryCorrupt
			}
		}
		if containsString(inventory.AppliedGrantIds, command.GrantID) {
			return inventory, true, nil
		}

		applyInventoryGrant(&inventory, command)
		inventory.AppliedGrantIds = append(inventory.AppliedGrantIds, command.GrantID)
		inventory.AppliedGrantIds = uniqueSorted(inventory.AppliedGrantIds)
		expectedVersion := inventory.InventoryVersion
		if !found {
			expectedVersion = 0
		}
		stored, err := m.store.StoreHangarInventory(command.Identity, inventory, expectedVersion)
		if err == nil {
			return stored, false, nil
		}
		if !errors.Is(err, ErrInventoryConflict) {
			return protocol.HangarInventory{}, false, err
		}
	}
	return protocol.HangarInventory{}, false, ErrInventoryConflict
}

func validateInventoryGrant(command protocol.PlayerDataApplyInventoryGrant) error {
	if IdentityKey(command.Identity) == "" {
		return errors.New("invalid identity")
	}
	if strings.TrimSpace(command.GrantID) == "" {
		return errors.New("grant_id is required")
	}
	if strings.TrimSpace(command.GrantKind) == "" {
		return errors.New("grant_kind is required")
	}
	switch command.GrantKind {
	case "unlock", "entitlement", "inventory_item", "rare_drop", "ship_part", "stackable_item", "reversal":
	default:
		return errors.New("unsupported inventory grant kind")
	}
	if command.GrantKind != "reversal" && strings.TrimSpace(command.CatalogRef) == "" {
		return errors.New("catalog_ref is required")
	}
	if command.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	return nil
}

func applyInventoryGrant(inventory *protocol.HangarInventory, command protocol.PlayerDataApplyInventoryGrant) {
	quantity := command.Quantity
	if quantity == 0 {
		quantity = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	acquisitionRef := command.AcquisitionRef
	if acquisitionRef == "" {
		acquisitionRef = command.GrantID
	}

	switch command.GrantKind {
	case "unlock", "entitlement":
		inventory.UnlockedContent = append(inventory.UnlockedContent, command.CatalogRef)
		inventory.UnlockedContent = uniqueSorted(inventory.UnlockedContent)
	case "inventory_item", "rare_drop":
		for i := 0; i < quantity; i++ {
			addOwnedInventoryItem(inventory, command, acquisitionRef, now, i)
		}
	case "ship_part", "stackable_item":
		addStackableItem(inventory, command.CatalogRef, quantity, now)
	case "reversal":
		reverseOwnedItem(inventory, command.TargetOwnedInstanceID)
	}
}

func addOwnedInventoryItem(inventory *protocol.HangarInventory, command protocol.PlayerDataApplyInventoryGrant, acquisitionRef, now string, index int) {
	idSeed := fmt.Sprintf("%s|%d", command.GrantID, index)
	switch {
	case strings.HasPrefix(command.CatalogRef, "ship.") || command.CatalogRef == StarterShipID:
		ownedID := stableOwnedID("s", inventory.PlayerRef, idSeed, command.CatalogRef)
		inventory.OwnedShips = append(inventory.OwnedShips, protocol.OwnedShip{OwnedShipID: ownedID, ShipID: strings.TrimPrefix(command.CatalogRef, "ship."), AcquiredAt: now, AcquisitionRef: acquisitionRef, HardwiredEquipment: []protocol.HardwiredEquipment{}, State: InventoryStateNormal})
	case strings.HasPrefix(command.CatalogRef, "weapon.") || command.CatalogRef == StarterPrimaryWeaponID:
		ownedID := stableOwnedID("w", inventory.PlayerRef, idSeed, command.CatalogRef)
		inventory.OwnedWeapons = append(inventory.OwnedWeapons, protocol.OwnedWeapon{OwnedWeaponID: ownedID, WeaponID: strings.TrimPrefix(command.CatalogRef, "weapon."), AcquiredAt: now, AcquisitionRef: acquisitionRef, State: InventoryStateNormal})
	case strings.HasPrefix(command.CatalogRef, "module."):
		ownedID := stableOwnedID("m", inventory.PlayerRef, idSeed, command.CatalogRef)
		inventory.OwnedModules = append(inventory.OwnedModules, protocol.OwnedModule{OwnedModuleID: ownedID, ModuleID: strings.TrimPrefix(command.CatalogRef, "module."), AcquiredAt: now, AcquisitionRef: acquisitionRef, State: InventoryStateNormal})
	default:
		addStackableItem(inventory, command.CatalogRef, 1, now)
	}
}

func addStackableItem(inventory *protocol.HangarInventory, itemRef string, quantity int, now string) {
	for i := range inventory.StackableItems {
		if inventory.StackableItems[i].ItemRef == itemRef {
			inventory.StackableItems[i].Quantity += quantity
			inventory.StackableItems[i].UpdatedAt = now
			return
		}
	}
	inventory.StackableItems = append(inventory.StackableItems, protocol.StackableInventoryItem{ItemRef: itemRef, Quantity: quantity, UpdatedAt: now})
}

func reverseOwnedItem(inventory *protocol.HangarInventory, ownedID string) {
	for i := range inventory.OwnedShips {
		if inventory.OwnedShips[i].OwnedShipID == ownedID {
			inventory.OwnedShips[i].State = InventoryStateReversed
		}
	}
	for i := range inventory.OwnedWeapons {
		if inventory.OwnedWeapons[i].OwnedWeaponID == ownedID {
			inventory.OwnedWeapons[i].State = InventoryStateReversed
		}
	}
	for i := range inventory.OwnedModules {
		if inventory.OwnedModules[i].OwnedModuleID == ownedID {
			inventory.OwnedModules[i].State = InventoryStateReversed
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
