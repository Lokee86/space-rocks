package playerinventory

import (
	"encoding/json"
	"errors"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
)

type PlayerDataSink interface {
	HandlePlayerDataCommand(payload []byte) ([]byte, error)
}

type RuntimeClient struct {
	sink PlayerDataSink
}

func NewRuntimeClient(sink PlayerDataSink) (*RuntimeClient, error) {
	if sink == nil {
		return nil, errors.New("player-data sink is required")
	}
	return &RuntimeClient{sink: sink}, nil
}

func (c *RuntimeClient) Load(identity playerbuild.InventoryIdentity, context playerbuild.InventoryLoadRequest) (playerbuild.InventoryLoadResult, error) {
	command := protocol.PlayerDataLoadHangarInventory{
		Type: protocol.PacketTypePlayerDataLoadHangarInventory,
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   identity.Kind,
			AccountID:      identity.AccountID,
			LocalProfileID: identity.LocalProfileID,
		},
		Context: protocol.PlayerDataRequestContext{PlayMode: context.PlayMode, TraceID: context.TraceID},
	}
	response, err := c.handle(command)
	if err != nil {
		return playerbuild.InventoryLoadResult{}, err
	}
	var result protocol.PlayerDataLoadHangarInventoryResult
	if err := json.Unmarshal(response, &result); err != nil {
		return playerbuild.InventoryLoadResult{}, err
	}
	converted := inventoryLoadResult(result)
	if result.ErrorCode != "" {
		return converted, errors.New("player-data runtime rejected inventory load")
	}
	return converted, nil
}

func (c *RuntimeClient) ApplyGrant(command protocol.PlayerDataApplyInventoryGrant) (protocol.PlayerDataApplyInventoryGrantResult, error) {
	command.Type = protocol.PacketTypePlayerDataApplyInventoryGrant
	response, err := c.handle(command)
	if err != nil {
		return protocol.PlayerDataApplyInventoryGrantResult{}, err
	}
	var result protocol.PlayerDataApplyInventoryGrantResult
	if err := json.Unmarshal(response, &result); err != nil {
		return protocol.PlayerDataApplyInventoryGrantResult{}, err
	}
	if !result.Accepted {
		return result, errors.New("player-data runtime rejected inventory grant")
	}
	return result, nil
}

func inventoryLoadResult(result protocol.PlayerDataLoadHangarInventoryResult) playerbuild.InventoryLoadResult {
	return playerbuild.InventoryLoadResult{
		Found:               result.Found,
		Persisted:           result.Persisted,
		SynthesizedFallback: result.SynthesizedFallback,
		RepairAttempted:     result.RepairAttempted,
		Inventory:           buildInventory(result.Inventory),
		ErrorCode:           result.ErrorCode,
		Message:             result.Message,
	}
}

func buildInventory(source protocol.HangarInventory) playerbuild.Inventory {
	inventory := playerbuild.Inventory{
		InventoryVersion:   source.InventoryVersion,
		DefaultOwnedShipID: source.DefaultOwnedShipID,
		OwnedShips:         make([]playerbuild.OwnedShip, 0, len(source.OwnedShips)),
		OwnedWeapons:       make([]playerbuild.OwnedWeapon, 0, len(source.OwnedWeapons)),
		OwnedModules:       make([]playerbuild.OwnedModule, 0, len(source.OwnedModules)),
	}
	for _, owned := range source.OwnedShips {
		ship := playerbuild.OwnedShip{
			OwnedShipID:        owned.OwnedShipID,
			ShipID:             owned.ShipID,
			State:              owned.State,
			HardwiredEquipment: make([]playerbuild.HardwiredEquipment, 0, len(owned.HardwiredEquipment)),
		}
		for _, hardwired := range owned.HardwiredEquipment {
			ship.HardwiredEquipment = append(ship.HardwiredEquipment, playerbuild.HardwiredEquipment{
				HardwiredID: hardwired.HardwiredID,
				EquipmentID: hardwired.EquipmentID,
				State:       hardwired.State,
			})
		}
		inventory.OwnedShips = append(inventory.OwnedShips, ship)
	}
	for _, owned := range source.OwnedWeapons {
		inventory.OwnedWeapons = append(inventory.OwnedWeapons, playerbuild.OwnedWeapon{
			OwnedWeaponID: owned.OwnedWeaponID,
			WeaponID:      owned.WeaponID,
			State:         owned.State,
		})
	}
	for _, owned := range source.OwnedModules {
		inventory.OwnedModules = append(inventory.OwnedModules, playerbuild.OwnedModule{
			OwnedModuleID: owned.OwnedModuleID,
			ModuleID:      owned.ModuleID,
			State:         owned.State,
		})
	}
	return inventory
}

func (c *RuntimeClient) handle(command any) ([]byte, error) {
	payload, err := codec.Encode(command)
	if err != nil {
		return nil, err
	}
	return c.sink.HandlePlayerDataCommand(payload)
}
