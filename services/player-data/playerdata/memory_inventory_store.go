package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func (s *MemoryStore) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	identityKey := IdentityKey(identity)
	if identityKey == "" {
		return protocol.HangarInventory{}, false, errors.New("invalid identity")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inventory, found := s.inventoryByIdentityKey[identityKey]
	if !found {
		return protocol.HangarInventory{}, false, nil
	}
	cloned, err := cloneHangarInventory(inventory)
	return cloned, true, err
}

func (s *MemoryStore) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	identityKey := IdentityKey(identity)
	if identityKey == "" {
		return protocol.HangarInventory{}, errors.New("invalid identity")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.inventoryByIdentityKey[identityKey]
	if expectedVersion >= 0 && expectedVersion != current.InventoryVersion {
		return protocol.HangarInventory{}, ErrInventoryConflict
	}
	stored, err := cloneHangarInventory(storeInventoryVersion(inventory, current.InventoryVersion))
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	s.inventoryByIdentityKey[identityKey] = stored
	return cloneHangarInventory(stored)
}
