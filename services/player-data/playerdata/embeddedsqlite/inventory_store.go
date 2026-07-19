//go:build !noembeddedsqlite

package embeddedsqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func (s *Store) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	if s == nil || s.db == nil {
		return protocol.HangarInventory{}, false, errors.New("sqlite store is not open")
	}
	if identity.IdentityKind != playerdata.IdentityKindLocalProfile {
		return protocol.HangarInventory{}, false, errors.New("identity_kind must be local_profile")
	}
	if identity.LocalProfileID == "" {
		return protocol.HangarInventory{}, false, errors.New("local_profile_id is required")
	}
	if err := s.ensureLocalProfile(identity.LocalProfileID); err != nil {
		return protocol.HangarInventory{}, false, err
	}

	var inventoryJSON string
	var inventoryVersion int
	err := s.db.QueryRow(
		`SELECT inventory_json, inventory_version FROM local_hangar_inventories WHERE local_profile_id = ?`,
		identity.LocalProfileID,
	).Scan(&inventoryJSON, &inventoryVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return protocol.HangarInventory{}, false, nil
	}
	if err != nil {
		return protocol.HangarInventory{}, false, err
	}

	var inventory protocol.HangarInventory
	if err := json.Unmarshal([]byte(inventoryJSON), &inventory); err != nil {
		return protocol.HangarInventory{}, false, fmt.Errorf("%w: %v", playerdata.ErrInventoryCorrupt, err)
	}
	inventory.InventoryVersion = inventoryVersion
	return inventory, true, nil
}

func (s *Store) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	if s == nil || s.db == nil {
		return protocol.HangarInventory{}, errors.New("sqlite store is not open")
	}
	if identity.IdentityKind != playerdata.IdentityKindLocalProfile {
		return protocol.HangarInventory{}, errors.New("identity_kind must be local_profile")
	}
	if identity.LocalProfileID == "" {
		return protocol.HangarInventory{}, errors.New("local_profile_id is required")
	}
	if err := s.ensureLocalProfile(identity.LocalProfileID); err != nil {
		return protocol.HangarInventory{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	defer func() { _ = tx.Rollback() }()

	currentVersion := 0
	err = tx.QueryRow(`SELECT inventory_version FROM local_hangar_inventories WHERE local_profile_id = ?`, identity.LocalProfileID).Scan(&currentVersion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return protocol.HangarInventory{}, err
	}
	if expectedVersion >= 0 && expectedVersion != currentVersion {
		return protocol.HangarInventory{}, playerdata.ErrInventoryConflict
	}

	inventory.InventoryVersion = currentVersion + 1
	payload, err := json.Marshal(inventory)
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	if currentVersion == 0 {
		_, err = tx.Exec(
			`INSERT INTO local_hangar_inventories (local_profile_id, inventory_version, inventory_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			identity.LocalProfileID, inventory.InventoryVersion, string(payload), now, now,
		)
	} else {
		var result sql.Result
		result, err = tx.Exec(
			`UPDATE local_hangar_inventories SET inventory_version = ?, inventory_json = ?, updated_at = ? WHERE local_profile_id = ? AND inventory_version = ?`,
			inventory.InventoryVersion, string(payload), now, identity.LocalProfileID, currentVersion,
		)
		if err == nil {
			rows, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return protocol.HangarInventory{}, rowsErr
			}
			if rows != 1 {
				return protocol.HangarInventory{}, playerdata.ErrInventoryConflict
			}
		}
	}
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	if err := tx.Commit(); err != nil {
		return protocol.HangarInventory{}, err
	}
	return inventory, nil
}
