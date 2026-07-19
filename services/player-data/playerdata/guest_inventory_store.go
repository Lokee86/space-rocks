package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func (s *GuestMemoryStore) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	if identity.IdentityKind != IdentityKindGuest {
		return protocol.HangarInventory{}, false, errors.New("identity_kind must be guest")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inventory == nil {
		return protocol.HangarInventory{}, false, nil
	}
	inventory, err := cloneHangarInventory(*s.inventory)
	return inventory, true, err
}

func (s *GuestMemoryStore) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	if identity.IdentityKind != IdentityKindGuest {
		return protocol.HangarInventory{}, errors.New("identity_kind must be guest")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	currentVersion := 0
	if s.inventory != nil {
		currentVersion = s.inventory.InventoryVersion
	}
	if expectedVersion >= 0 && expectedVersion != currentVersion {
		return protocol.HangarInventory{}, ErrInventoryConflict
	}
	stored, err := cloneHangarInventory(storeInventoryVersion(inventory, currentVersion))
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	s.inventory = &stored
	return cloneHangarInventory(stored)
}
