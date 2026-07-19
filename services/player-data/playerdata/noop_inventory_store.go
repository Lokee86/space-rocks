package playerdata

import "github.com/Lokee86/space-rocks/player-data/protocol"

func (s *NoopStore) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error) {
	return protocol.HangarInventory{}, false, nil
}

func (s *NoopStore) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error) {
	inventory.InventoryVersion++
	return inventory, nil
}
