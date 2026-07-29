package playerinventory_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerinventory"
)

type recordingSink struct {
	payload  []byte
	response []byte
	err      error
}

func (s *recordingSink) HandlePlayerDataCommand(payload []byte) ([]byte, error) {
	s.payload = payload
	return s.response, s.err
}

func TestRuntimeClientLoadsInventoryForAllIdentityRoutes(t *testing.T) {
	tests := []struct {
		name     string
		identity playerbuild.InventoryIdentity
		expected protocol.PlayerDataIdentity
		playMode string
	}{
		{
			name:     "guest",
			identity: playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityGuest},
			expected: protocol.PlayerDataIdentity{IdentityKind: playerbuild.InventoryIdentityGuest},
			playMode: "single_player",
		},
		{
			name:     "local profile",
			identity: playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityLocalProfile, LocalProfileID: "local-1"},
			expected: protocol.PlayerDataIdentity{IdentityKind: playerbuild.InventoryIdentityLocalProfile, LocalProfileID: "local-1"},
			playMode: "single_player",
		},
		{
			name:     "authenticated account",
			identity: playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityAuthenticatedAccount, AccountID: "account-1"},
			expected: protocol.PlayerDataIdentity{IdentityKind: playerbuild.InventoryIdentityAuthenticatedAccount, AccountID: "account-1"},
			playMode: "multiplayer",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := codec.Encode(protocol.PlayerDataLoadHangarInventoryResult{
				Type:      protocol.PacketTypePlayerDataLoadHangarInventoryResult,
				Found:     true,
				Persisted: true,
				Inventory: protocol.HangarInventory{
					InventoryVersion:   7,
					DefaultOwnedShipID: "ship-1",
					OwnedShips: []protocol.OwnedShip{{
						OwnedShipID: "ship-1",
						ShipID:      "v_wing",
						State:       "normal",
						HardwiredEquipment: []protocol.HardwiredEquipment{{
							HardwiredID: "hardwired-1", EquipmentID: "reinforced_hull", State: "normal",
						}},
					}},
					OwnedWeapons: []protocol.OwnedWeapon{{OwnedWeaponID: "weapon-1", WeaponID: "pulse", State: "normal"}},
					OwnedModules: []protocol.OwnedModule{{OwnedModuleID: "module-1", ModuleID: "shield_capacitor", State: "normal"}},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			sink := &recordingSink{response: response}
			client, err := playerinventory.NewRuntimeClient(sink)
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.Load(test.identity, playerbuild.InventoryLoadRequest{PlayMode: test.playMode, TraceID: "trace-1"})
			if err != nil {
				t.Fatal(err)
			}
			if !result.Found || !result.Persisted || result.Inventory.InventoryVersion != 7 {
				t.Fatalf("unexpected result: %#v", result)
			}
			if len(result.Inventory.OwnedShips) != 1 || len(result.Inventory.OwnedShips[0].HardwiredEquipment) != 1 ||
				len(result.Inventory.OwnedWeapons) != 1 || len(result.Inventory.OwnedModules) != 1 {
				t.Fatalf("inventory projection lost owned equipment: %#v", result.Inventory)
			}
			var command protocol.PlayerDataLoadHangarInventory
			if err := json.Unmarshal(sink.payload, &command); err != nil {
				t.Fatal(err)
			}
			if command.Type != protocol.PacketTypePlayerDataLoadHangarInventory || command.Identity != test.expected || command.Context.PlayMode != test.playMode {
				t.Fatalf("unexpected command: %#v", command)
			}
		})
	}
}

func TestRuntimeClientProjectsStarterInventoryIntoBuildEligibility(t *testing.T) {
	identity := protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest}
	response, err := codec.Encode(protocol.PlayerDataLoadHangarInventoryResult{
		Type:      protocol.PacketTypePlayerDataLoadHangarInventoryResult,
		Found:     true,
		Persisted: true,
		Inventory: playerdata.StarterHangarInventory(identity),
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := playerinventory.NewRuntimeClient(&recordingSink{response: response})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := client.Load(
		playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityGuest},
		playerbuild.InventoryLoadRequest{PlayMode: playerdata.PlayModeSinglePlayer},
	)
	if err != nil {
		t.Fatal(err)
	}
	options := playerbuild.ComputeEligibility("player-1", loaded.Inventory, playerbuild.DefaultCatalog(), playerbuild.Rules{})
	if len(options.BlockedOptions) != 0 || len(options.EligibleShips) != 2 {
		t.Fatalf("starter inventory projection is incompatible with build catalog: %+v", options)
	}
}

func TestRuntimeClientPreservesFallbackLoadState(t *testing.T) {
	response, err := codec.Encode(protocol.PlayerDataLoadHangarInventoryResult{
		Type:                protocol.PacketTypePlayerDataLoadHangarInventoryResult,
		Found:               true,
		SynthesizedFallback: true,
		RepairAttempted:     true,
		Inventory:           protocol.HangarInventory{SchemaVersion: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := playerinventory.NewRuntimeClient(&recordingSink{response: response})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Load(
		playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityGuest},
		playerbuild.InventoryLoadRequest{PlayMode: "single_player"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.SynthesizedFallback || !result.RepairAttempted {
		t.Fatalf("fallback state lost: %#v", result)
	}
}

func TestRuntimeClientAppliesDurableGrant(t *testing.T) {
	response, err := codec.Encode(protocol.PlayerDataApplyInventoryGrantResult{
		Type:      protocol.PacketTypePlayerDataApplyInventoryGrantResult,
		Accepted:  true,
		Duplicate: false,
		Inventory: protocol.HangarInventory{InventoryVersion: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	sink := &recordingSink{response: response}
	client, err := playerinventory.NewRuntimeClient(sink)
	if err != nil {
		t.Fatal(err)
	}
	command := protocol.PlayerDataApplyInventoryGrant{
		GrantID:    "grant-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: "authenticated_account", AccountID: "account-1"},
		Context:    protocol.PlayerDataRequestContext{PlayMode: "multiplayer"},
		GrantKind:  "inventory_item",
		CatalogRef: "weapon.railgun",
	}
	result, err := client.ApplyGrant(command)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Accepted || result.Inventory.InventoryVersion != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
	var encoded protocol.PlayerDataApplyInventoryGrant
	if err := json.Unmarshal(sink.payload, &encoded); err != nil {
		t.Fatal(err)
	}
	if encoded.Type != protocol.PacketTypePlayerDataApplyInventoryGrant || encoded.GrantID != command.GrantID {
		t.Fatalf("unexpected encoded grant: %#v", encoded)
	}
}

func TestRuntimeClientRejectsPlayerDataFailure(t *testing.T) {
	response, err := codec.Encode(protocol.PlayerDataLoadHangarInventoryResult{
		Type:      protocol.PacketTypePlayerDataLoadHangarInventoryResult,
		ErrorCode: "inventory_load_failed",
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := playerinventory.NewRuntimeClient(&recordingSink{response: response})
	if err != nil {
		t.Fatal(err)
	}
	result, loadErr := client.Load(
		playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityGuest},
		playerbuild.InventoryLoadRequest{PlayMode: "single_player"},
	)
	if loadErr == nil {
		t.Fatal("expected rejected load")
	}
	if result.ErrorCode != "inventory_load_failed" {
		t.Fatalf("error code was not preserved: %+v", result)
	}
}

func TestRuntimeClientPropagatesSinkFailure(t *testing.T) {
	client, err := playerinventory.NewRuntimeClient(&recordingSink{err: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Load(
		playerbuild.InventoryIdentity{Kind: playerbuild.InventoryIdentityGuest},
		playerbuild.InventoryLoadRequest{PlayMode: "single_player"},
	); err == nil {
		t.Fatal("expected sink failure")
	}
}
