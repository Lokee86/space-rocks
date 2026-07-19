package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type inventoryProfileAuthVerifier struct{}

func (inventoryProfileAuthVerifier) VerifyToken(ctx context.Context, rawToken string) (AuthVerificationResult, error) {
	if rawToken != "account-token" {
		return AuthVerificationResult{}, nil
	}
	return AuthVerificationResult{
		Valid: true,
		Identity: AuthIdentity{
			AccountID:   "account-1",
			DisplayName: "Account Pilot",
		},
	}, nil
}

type inventoryProfileResponse struct {
	Profile struct {
		IdentityKind                 string                   `json:"identity_kind"`
		Inventory                    protocol.HangarInventory `json:"inventory"`
		InventoryPersisted           bool                     `json:"inventory_persisted"`
		InventorySynthesizedFallback bool                     `json:"inventory_synthesized_fallback"`
		InventoryRepairAttempted     bool                     `json:"inventory_repair_attempted"`
	} `json:"profile"`
}

func TestProfileHandlerReturnsDurableStarterInventoryAcrossIdentityRoutes(t *testing.T) {
	router := playerdata.NewStoreRouter(playerdata.NewMemoryStore(), playerdata.NewMemoryStore(), playerdata.NewGuestMemoryStore())
	runtime, err := playerdata.NewRuntime(playerdata.Config{Store: router})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewProfileHandler(runtime, inventoryProfileAuthVerifier{})

	tests := []struct {
		name          string
		body          string
		authorization string
		identity      protocol.PlayerDataIdentity
	}{
		{
			name:     "guest",
			body:     `{"play_mode":"single_player","identity_kind":"guest"}`,
			identity: protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest},
		},
		{
			name:     "local profile",
			body:     `{"play_mode":"single_player","identity_kind":"local_profile","local_profile_id":"local-1"}`,
			identity: protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: "local-1"},
		},
		{
			name:          "authenticated account",
			body:          `{"play_mode":"multiplayer","identity_kind":"authenticated_account"}`,
			authorization: "Bearer account-token",
			identity:      protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindAuthenticatedAccount, AccountID: "account-1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := requestInventoryProfile(t, handler, test.body, test.authorization)
			if first.Profile.IdentityKind != test.identity.IdentityKind {
				t.Fatalf("identity_kind = %q", first.Profile.IdentityKind)
			}
			if !first.Profile.InventoryPersisted || first.Profile.InventorySynthesizedFallback || first.Profile.InventoryRepairAttempted {
				t.Fatalf("unexpected inventory state: %#v", first.Profile)
			}
			if len(first.Profile.Inventory.OwnedShips) != 1 || first.Profile.Inventory.OwnedShips[0].ShipID != playerdata.StarterShipID {
				t.Fatalf("starter ship missing: %#v", first.Profile.Inventory)
			}
			if len(first.Profile.Inventory.OwnedWeapons) != 1 || first.Profile.Inventory.OwnedWeapons[0].WeaponID != playerdata.StarterPrimaryWeaponID {
				t.Fatalf("starter weapon missing: %#v", first.Profile.Inventory)
			}
			if len(first.Profile.Inventory.OwnedModules) != 0 || len(first.Profile.Inventory.OwnedShips[0].HardwiredEquipment) != 0 {
				t.Fatalf("starter inventory contains optional equipment: %#v", first.Profile.Inventory)
			}

			second := requestInventoryProfile(t, handler, test.body, test.authorization)
			if second.Profile.Inventory.OwnedShips[0].OwnedShipID != first.Profile.Inventory.OwnedShips[0].OwnedShipID {
				t.Fatalf("owned ship id changed across profile loads")
			}
			if second.Profile.Inventory.InventoryVersion != first.Profile.Inventory.InventoryVersion {
				t.Fatalf("profile reload changed inventory version")
			}

			persisted, found, err := router.LoadHangarInventory(test.identity)
			if err != nil {
				t.Fatal(err)
			}
			if !found || persisted.OwnedShips[0].OwnedShipID != first.Profile.Inventory.OwnedShips[0].OwnedShipID {
				t.Fatalf("profile inventory was not stored on the selected route")
			}
		})
	}
}

func requestInventoryProfile(t *testing.T, handler http.Handler, body string, authorization string) inventoryProfileResponse {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	var decoded inventoryProfileResponse
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}
