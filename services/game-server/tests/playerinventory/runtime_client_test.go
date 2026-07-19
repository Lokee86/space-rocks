package playerinventory_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
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
		identity protocol.PlayerDataIdentity
		playMode string
	}{
		{name: "guest", identity: protocol.PlayerDataIdentity{IdentityKind: "guest"}, playMode: "single_player"},
		{name: "local profile", identity: protocol.PlayerDataIdentity{IdentityKind: "local_profile", LocalProfileID: "local-1"}, playMode: "single_player"},
		{name: "authenticated account", identity: protocol.PlayerDataIdentity{IdentityKind: "authenticated_account", AccountID: "account-1"}, playMode: "multiplayer"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := codec.Encode(protocol.PlayerDataLoadHangarInventoryResult{
				Type:      protocol.PacketTypePlayerDataLoadHangarInventoryResult,
				Found:     true,
				Persisted: true,
				Inventory: protocol.HangarInventory{SchemaVersion: 1, PlayerRef: test.identity.IdentityKind},
			})
			if err != nil {
				t.Fatal(err)
			}
			sink := &recordingSink{response: response}
			client, err := playerinventory.NewRuntimeClient(sink)
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.Load(test.identity, protocol.PlayerDataRequestContext{PlayMode: test.playMode, TraceID: "trace-1"})
			if err != nil {
				t.Fatal(err)
			}
			if !result.Found || !result.Persisted {
				t.Fatalf("unexpected result: %#v", result)
			}
			var command protocol.PlayerDataLoadHangarInventory
			if err := json.Unmarshal(sink.payload, &command); err != nil {
				t.Fatal(err)
			}
			if command.Type != protocol.PacketTypePlayerDataLoadHangarInventory || command.Identity != test.identity || command.Context.PlayMode != test.playMode {
				t.Fatalf("unexpected command: %#v", command)
			}
		})
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
	result, err := client.Load(protocol.PlayerDataIdentity{IdentityKind: "guest"}, protocol.PlayerDataRequestContext{PlayMode: "single_player"})
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
	if _, err := client.Load(protocol.PlayerDataIdentity{IdentityKind: "guest"}, protocol.PlayerDataRequestContext{PlayMode: "single_player"}); err == nil {
		t.Fatal("expected rejected load")
	}
}

func TestRuntimeClientPropagatesSinkFailure(t *testing.T) {
	client, err := playerinventory.NewRuntimeClient(&recordingSink{err: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Load(protocol.PlayerDataIdentity{IdentityKind: "guest"}, protocol.PlayerDataRequestContext{PlayMode: "single_player"}); err == nil {
		t.Fatal("expected sink failure")
	}
}
