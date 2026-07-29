package playerdata

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestDispatcherLoadsStarterHangarForGuest(t *testing.T) {
	dispatcher := NewDispatcher(NewGuestMemoryStore())
	payload, err := codec.Encode(protocol.PlayerDataLoadHangarInventory{
		Type:     protocol.PacketTypePlayerDataLoadHangarInventory,
		Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatal(err)
	}
	var result protocol.PlayerDataLoadHangarInventoryResult
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatal(err)
	}
	if !result.Found || !result.Persisted || result.ErrorCode != "" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Inventory.OwnedShips[0].ShipID != StarterShipID || result.Inventory.OwnedWeapons[0].WeaponID != StarterPrimaryWeaponID {
		t.Fatalf("starter inventory missing: %#v", result.Inventory)
	}
}

func TestDispatcherAppliesInventoryGrantIdempotently(t *testing.T) {
	dispatcher := NewDispatcher(NewGuestMemoryStore())
	command := protocol.PlayerDataApplyInventoryGrant{
		Type:       protocol.PacketTypePlayerDataApplyInventoryGrant,
		GrantID:    "grant-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
		GrantKind:  "inventory_item",
		CatalogRef: "module.overcharger",
	}
	payload, err := codec.Encode(command)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		response, err := dispatcher.Handle(payload)
		if err != nil {
			t.Fatal(err)
		}
		var result protocol.PlayerDataApplyInventoryGrantResult
		if err := json.Unmarshal(response, &result); err != nil {
			t.Fatal(err)
		}
		if !result.Accepted || len(result.Inventory.OwnedModules) != 5 {
			t.Fatalf("unexpected result: %#v", result)
		}
		if (attempt == 1) != result.Duplicate {
			t.Fatalf("duplicate=%v on attempt %d", result.Duplicate, attempt)
		}
	}
}

func TestDispatcherRejectsInventoryIdentityRouteMismatch(t *testing.T) {
	dispatcher := NewDispatcher(NewGuestMemoryStore())
	payload, err := codec.Encode(protocol.PlayerDataLoadHangarInventory{
		Type:     protocol.PacketTypePlayerDataLoadHangarInventory,
		Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatal(err)
	}
	var result protocol.PlayerDataLoadHangarInventoryResult
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatal(err)
	}
	if result.ErrorCode != "invalid_mode_identity" {
		t.Fatalf("unexpected error code %q", result.ErrorCode)
	}
}
