package playerdata

import (
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func guestIdentity() protocol.PlayerDataIdentity {
	return protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest}
}

func TestStarterHangarInventoryIsPlayableAndStable(t *testing.T) {
	first := StarterHangarInventory(guestIdentity())
	second := StarterHangarInventory(guestIdentity())

	if first.SchemaVersion != HangarInventorySchemaVersion {
		t.Fatalf("unexpected schema version %d", first.SchemaVersion)
	}
	if len(first.OwnedShips) != 2 || first.OwnedShips[0].ShipID != StarterShipID || first.OwnedShips[1].ShipID != StarterScoutShipID {
		t.Fatalf("unexpected starter ships: %#v", first.OwnedShips)
	}
	if len(first.OwnedWeapons) != 2 || first.OwnedWeapons[0].WeaponID != StarterPrimaryWeaponID || first.OwnedWeapons[1].WeaponID != StarterSecondaryWeaponID {
		t.Fatalf("unexpected starter weapons: %#v", first.OwnedWeapons)
	}
	if len(first.OwnedModules) != 4 {
		t.Fatalf("unexpected starter modules: %#v", first.OwnedModules)
	}
	if len(first.OwnedShips[0].HardwiredEquipment) != 0 {
		t.Fatalf("starter ship should not contain hardwired equipment")
	}
	if first.OwnedShips[0].OwnedShipID != second.OwnedShips[0].OwnedShipID {
		t.Fatalf("starter ship id changed")
	}
	if first.OwnedWeapons[0].OwnedWeaponID != second.OwnedWeapons[0].OwnedWeaponID {
		t.Fatalf("starter weapon id changed")
	}
	if err := ValidateHangarInventory(guestIdentity(), first); err != nil {
		t.Fatalf("starter inventory invalid: %v", err)
	}
}

func TestInventoryManagerInitializesMissingInventoryDurably(t *testing.T) {
	store := NewMemoryStore()
	manager := NewInventoryManager(store)
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile, LocalProfileID: "local-1"}

	first, err := manager.Load(identity)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Persisted || first.SynthesizedFallback {
		t.Fatalf("unexpected first load flags: %#v", first)
	}
	if first.Inventory.InventoryVersion != 1 {
		t.Fatalf("expected inventory version 1, got %d", first.Inventory.InventoryVersion)
	}

	second, err := manager.Load(identity)
	if err != nil {
		t.Fatal(err)
	}
	if second.Inventory.OwnedShips[0].OwnedShipID != first.Inventory.OwnedShips[0].OwnedShipID {
		t.Fatalf("owned ship id was not stable")
	}
	if second.Inventory.InventoryVersion != first.Inventory.InventoryVersion {
		t.Fatalf("load changed inventory version")
	}
}

func TestInventoryManagerUpgradesLegacyStarterCatalogOnce(t *testing.T) {
	store := NewMemoryStore()
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile, LocalProfileID: "legacy-1"}
	legacy := protocol.HangarInventory{
		SchemaVersion: HangarInventorySchemaVersion,
		PlayerRef:     IdentityKey(identity),
		OwnedShips: []protocol.OwnedShip{{
			OwnedShipID: stableOwnedID("s", IdentityKey(identity), "starter", StarterShipID),
			ShipID:      StarterShipID, HardwiredEquipment: []protocol.HardwiredEquipment{}, State: InventoryStateNormal,
		}},
		OwnedWeapons: []protocol.OwnedWeapon{{
			OwnedWeaponID: stableOwnedID("w", IdentityKey(identity), "starter", StarterPrimaryWeaponID),
			WeaponID:      StarterPrimaryWeaponID, State: InventoryStateNormal,
		}},
		OwnedModules:       []protocol.OwnedModule{},
		UnlockedContent:    []string{StarterShipID, StarterPrimaryWeaponID},
		StackableItems:     []protocol.StackableInventoryItem{},
		DefaultOwnedShipID: stableOwnedID("s", IdentityKey(identity), "starter", StarterShipID),
		AppliedGrantIds:    []string{},
	}
	stored, err := store.StoreHangarInventory(identity, legacy, 0)
	if err != nil {
		t.Fatal(err)
	}

	first, err := NewInventoryManager(store).Load(identity)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Persisted || first.Inventory.InventoryVersion != stored.InventoryVersion+1 {
		t.Fatalf("legacy catalog upgrade was not persisted once: %#v", first)
	}
	if len(first.Inventory.OwnedShips) != 2 || len(first.Inventory.OwnedWeapons) != 2 || len(first.Inventory.OwnedModules) != 4 {
		t.Fatalf("legacy catalog was not expanded: %#v", first.Inventory)
	}
	if first.Inventory.DefaultOwnedShipID != legacy.DefaultOwnedShipID {
		t.Fatalf("catalog upgrade changed the default ship")
	}

	second, err := NewInventoryManager(store).Load(identity)
	if err != nil {
		t.Fatal(err)
	}
	if second.Inventory.InventoryVersion != first.Inventory.InventoryVersion {
		t.Fatalf("catalog upgrade repeated: first=%d second=%d", first.Inventory.InventoryVersion, second.Inventory.InventoryVersion)
	}
	if len(second.Inventory.OwnedModules) != 4 {
		t.Fatalf("catalog upgrade duplicated modules: %#v", second.Inventory.OwnedModules)
	}
}

func TestInventoryManagerRepairsInvalidPersistedInventory(t *testing.T) {
	store := NewMemoryStore()
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindLocalProfile, LocalProfileID: "local-1"}
	invalid := StarterHangarInventory(identity)
	invalid.SchemaVersion = 99
	invalid.PlayerRef = "wrong-player"
	stored, err := store.StoreHangarInventory(identity, invalid, 0)
	if err != nil {
		t.Fatal(err)
	}

	load, err := NewInventoryManager(store).Load(identity)
	if err != nil {
		t.Fatal(err)
	}
	if !load.SynthesizedFallback || !load.RepairAttempted || !load.Persisted {
		t.Fatalf("unexpected repair flags: %#v", load)
	}
	if load.Inventory.SchemaVersion != HangarInventorySchemaVersion || load.Inventory.PlayerRef != IdentityKey(identity) {
		t.Fatalf("invalid inventory was not repaired: %#v", load.Inventory)
	}
	if load.Inventory.InventoryVersion != stored.InventoryVersion+1 {
		t.Fatalf("repair did not advance version: got %d want %d", load.Inventory.InventoryVersion, stored.InventoryVersion+1)
	}
}

func TestInventoryManagerGuestStorageIsTransient(t *testing.T) {
	firstManager := NewInventoryManager(NewGuestMemoryStore())
	loaded, err := firstManager.Load(guestIdentity())
	if err != nil {
		t.Fatal(err)
	}
	grant := protocol.PlayerDataApplyInventoryGrant{GrantID: "grant-1", Identity: guestIdentity(), GrantKind: "inventory_item", CatalogRef: "module.overcharger"}
	updated, _, err := firstManager.ApplyGrant(grant)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.OwnedModules) != 5 {
		t.Fatalf("expected starter modules plus transient grant")
	}

	secondManager := NewInventoryManager(NewGuestMemoryStore())
	fresh, err := secondManager.Load(guestIdentity())
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh.Inventory.OwnedModules) != 4 {
		t.Fatalf("guest inventory leaked across stores or lost starter modules")
	}
	if fresh.Inventory.OwnedShips[0].OwnedShipID != loaded.Inventory.OwnedShips[0].OwnedShipID {
		t.Fatalf("starter identity should remain deterministic")
	}
}

func TestInventoryGrantIsIdempotentAndUsesStableOwnedIDs(t *testing.T) {
	store := NewMemoryStore()
	manager := NewInventoryManager(store)
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "account-1"}
	command := protocol.PlayerDataApplyInventoryGrant{GrantID: "reward-1", Identity: identity, GrantKind: "inventory_item", CatalogRef: "weapon.railgun", Quantity: 1}

	first, duplicate, err := manager.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate {
		t.Fatalf("first grant was duplicate")
	}
	if len(first.OwnedWeapons) != 3 {
		t.Fatalf("expected starter weapons plus granted weapon, got %d", len(first.OwnedWeapons))
	}
	ownedID := first.OwnedWeapons[2].OwnedWeaponID

	second, duplicate, err := manager.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate {
		t.Fatalf("replayed grant was not duplicate")
	}
	if len(second.OwnedWeapons) != 3 || second.OwnedWeapons[2].OwnedWeaponID != ownedID {
		t.Fatalf("duplicate grant changed ownership")
	}
}

func TestUnlockGrantDoesNotCreateOwnership(t *testing.T) {
	manager := NewInventoryManager(NewMemoryStore())
	command := protocol.PlayerDataApplyInventoryGrant{GrantID: "unlock-1", Identity: guestIdentity(), GrantKind: "unlock", CatalogRef: "weapon.railgun"}
	inventory, _, err := manager.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.OwnedWeapons) != 2 {
		t.Fatalf("unlock created ownership")
	}
	if !containsString(inventory.UnlockedContent, "weapon.railgun") {
		t.Fatalf("unlock was not recorded")
	}
}

func TestStackableGrantIsIdempotent(t *testing.T) {
	manager := NewInventoryManager(NewMemoryStore())
	command := protocol.PlayerDataApplyInventoryGrant{GrantID: "parts-1", Identity: guestIdentity(), GrantKind: "ship_part", CatalogRef: "item.coolant_fragment", Quantity: 3}
	first, _, err := manager.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	second, duplicate, err := manager.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate || first.StackableItems[0].Quantity != 3 || second.StackableItems[0].Quantity != 3 {
		t.Fatalf("stackable grant was not idempotent")
	}
}

func TestRuntimePickupGrantKindIsRejected(t *testing.T) {
	manager := NewInventoryManager(NewMemoryStore())
	_, _, err := manager.ApplyGrant(protocol.PlayerDataApplyInventoryGrant{GrantID: "pickup-1", Identity: guestIdentity(), GrantKind: "runtime_pickup", CatalogRef: "weapon.railgun"})
	if err == nil {
		t.Fatalf("runtime pickup entered durable inventory")
	}
}

type failingInventoryStore struct {
	loadErr    error
	storeCalls int
	stored     protocol.HangarInventory
}

func (s *failingInventoryStore) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	return protocol.HangarInventory{}, false, s.loadErr
}

func (s *failingInventoryStore) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	s.storeCalls++
	inventory.InventoryVersion = 1
	s.stored = inventory
	return inventory, nil
}

func TestCorruptInventoryProducesPlayableFallbackAndRepairs(t *testing.T) {
	store := &failingInventoryStore{loadErr: ErrInventoryCorrupt}
	load, err := NewInventoryManager(store).Load(guestIdentity())
	if err != nil {
		t.Fatal(err)
	}
	if !load.SynthesizedFallback || !load.RepairAttempted || !load.Persisted {
		t.Fatalf("unexpected fallback flags: %#v", load)
	}
	if store.storeCalls != 1 {
		t.Fatalf("expected one controlled repair attempt")
	}
	if err := ValidateHangarInventory(guestIdentity(), load.Inventory); err != nil {
		t.Fatalf("fallback is not playable: %v", err)
	}
}

func TestUnavailableInventoryProducesFallbackWithoutUnsafeWrite(t *testing.T) {
	store := &failingInventoryStore{loadErr: errors.New("temporarily unavailable")}
	load, err := NewInventoryManager(store).Load(guestIdentity())
	if err != nil {
		t.Fatal(err)
	}
	if !load.SynthesizedFallback || load.RepairAttempted || load.Persisted {
		t.Fatalf("unexpected unavailable flags: %#v", load)
	}
	if store.storeCalls != 0 {
		t.Fatalf("unavailable store should not receive repair write")
	}
}
