//go:build !noembeddedsqlite

package embeddedsqlite

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestLocalHangarInventoryPersistsAcrossReopen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "player-data.sqlite")
	identity := protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: "profile-1"}

	store, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(); err != nil {
		t.Fatal(err)
	}
	inventory := playerdata.StarterHangarInventory(identity)
	stored, err := store.StoreHangarInventory(identity, inventory, 0)
	if err != nil {
		t.Fatal(err)
	}
	if stored.InventoryVersion != 1 {
		t.Fatalf("expected version 1, got %d", stored.InventoryVersion)
	}
	ownedShipID := stored.OwnedShips[0].OwnedShipID
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.InitSchema(); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := reopened.LoadHangarInventory(identity)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatalf("inventory missing after reopen")
	}
	if loaded.OwnedShips[0].OwnedShipID != ownedShipID {
		t.Fatalf("owned instance id changed after persistence")
	}
	if loaded.InventoryVersion != 1 {
		t.Fatalf("unexpected version %d", loaded.InventoryVersion)
	}
}

func TestLocalHangarInventoryEnforcesExpectedVersion(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InitSchema(); err != nil {
		t.Fatal(err)
	}
	identity := protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: "profile-1"}
	inventory := playerdata.StarterHangarInventory(identity)
	stored, err := store.StoreHangarInventory(identity, inventory, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.StoreHangarInventory(identity, stored, 0); !errors.Is(err, playerdata.ErrInventoryConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	updated, err := store.StoreHangarInventory(identity, stored, 1)
	if err != nil {
		t.Fatal(err)
	}
	if updated.InventoryVersion != 2 {
		t.Fatalf("expected version 2, got %d", updated.InventoryVersion)
	}
}

func TestLocalHangarInventoryCorruptDocumentIsClassified(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InitSchema(); err != nil {
		t.Fatal(err)
	}
	identity := protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: "profile-1"}
	if err := store.ensureLocalProfile(identity.LocalProfileID); err != nil {
		t.Fatal(err)
	}
	_, err = store.db.Exec(`INSERT INTO local_hangar_inventories (local_profile_id, inventory_version, inventory_json, created_at, updated_at) VALUES (?, 1, ?, 'now', 'now')`, identity.LocalProfileID, "{broken")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.LoadHangarInventory(identity); !errors.Is(err, playerdata.ErrInventoryCorrupt) {
		t.Fatalf("expected corrupt classification, got %v", err)
	}
}

func TestDeleteLocalProfileDeletesHangarInventory(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InitSchema(); err != nil {
		t.Fatal(err)
	}
	identity := protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: "profile-1"}
	if _, err := store.StoreHangarInventory(identity, playerdata.StarterHangarInventory(identity), 0); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteLocalProfile(identity.LocalProfileID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM local_hangar_inventories WHERE local_profile_id = ?`, identity.LocalProfileID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("hangar inventory survived profile deletion")
	}
}
