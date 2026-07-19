package playerdata

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestRailsStoreLoadsHangarInventory(t *testing.T) {
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"}
	inventory := StarterHangarInventory(identity)
	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/player-data/inventory/load" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer internal-token" {
			t.Fatalf("missing internal auth")
		}
		var body struct {
			AccountID string `json:"account_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.AccountID != identity.AccountID {
			t.Fatalf("unexpected account id %q", body.AccountID)
		}
		inventory.InventoryVersion = 3
		_ = json.NewEncoder(w).Encode(map[string]any{"found": true, "inventory": inventory, "inventory_version": 3})
	}))
	store := &RailsStore{BaseURL: server.URL, internalToken: "internal-token", httpClient: server.Client()}
	loaded, found, err := store.LoadHangarInventory(identity)
	if err != nil {
		t.Fatal(err)
	}
	if !found || loaded.InventoryVersion != 3 {
		t.Fatalf("unexpected load: found=%v inventory=%#v", found, loaded)
	}
}

func TestRailsStoreMissingHangarInventory(t *testing.T) {
	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	store := &RailsStore{BaseURL: server.URL, internalToken: "internal-token", httpClient: server.Client()}
	_, found, err := store.LoadHangarInventory(protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"})
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatalf("missing inventory reported found")
	}
}

func TestRailsStoreStoresHangarInventoryWithExpectedVersion(t *testing.T) {
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"}
	inventory := StarterHangarInventory(identity)
	inventory.InventoryVersion = 2
	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/player-data/inventory/store" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		var body struct {
			AccountID       string                   `json:"account_id"`
			Inventory       protocol.HangarInventory `json:"inventory"`
			ExpectedVersion *int                     `json:"expected_version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.ExpectedVersion == nil || *body.ExpectedVersion != 2 {
			t.Fatalf("unexpected expected_version %#v", body.ExpectedVersion)
		}
		body.Inventory.InventoryVersion = 3
		_ = json.NewEncoder(w).Encode(map[string]any{"inventory": body.Inventory, "inventory_version": 3})
	}))
	store := &RailsStore{BaseURL: server.URL, internalToken: "internal-token", httpClient: server.Client()}
	stored, err := store.StoreHangarInventory(identity, inventory, 2)
	if err != nil {
		t.Fatal(err)
	}
	if stored.InventoryVersion != 3 {
		t.Fatalf("expected version 3, got %d", stored.InventoryVersion)
	}
}

func TestRailsStoreMapsInventoryConflict(t *testing.T) {
	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusConflict) }))
	store := &RailsStore{BaseURL: server.URL, internalToken: "internal-token", httpClient: server.Client()}
	identity := protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"}
	_, err := store.StoreHangarInventory(identity, StarterHangarInventory(identity), 1)
	if !errors.Is(err, ErrInventoryConflict) {
		t.Fatalf("expected inventory conflict, got %v", err)
	}
}
