package playerdata

import (
	"encoding/json"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func cloneHangarInventory(inventory protocol.HangarInventory) (protocol.HangarInventory, error) {
	payload, err := json.Marshal(inventory)
	if err != nil {
		return protocol.HangarInventory{}, err
	}
	var cloned protocol.HangarInventory
	if err := json.Unmarshal(payload, &cloned); err != nil {
		return protocol.HangarInventory{}, err
	}
	return cloned, nil
}

func storeInventoryVersion(inventory protocol.HangarInventory, currentVersion int) protocol.HangarInventory {
	inventory.InventoryVersion = currentVersion + 1
	return inventory
}
