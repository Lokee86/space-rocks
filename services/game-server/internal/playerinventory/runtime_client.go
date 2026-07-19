package playerinventory

import (
	"encoding/json"
	"errors"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
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

func (c *RuntimeClient) Load(identity protocol.PlayerDataIdentity, context protocol.PlayerDataRequestContext) (protocol.PlayerDataLoadHangarInventoryResult, error) {
	command := protocol.PlayerDataLoadHangarInventory{
		Type:     protocol.PacketTypePlayerDataLoadHangarInventory,
		Identity: identity,
		Context:  context,
	}
	response, err := c.handle(command)
	if err != nil {
		return protocol.PlayerDataLoadHangarInventoryResult{}, err
	}
	var result protocol.PlayerDataLoadHangarInventoryResult
	if err := json.Unmarshal(response, &result); err != nil {
		return protocol.PlayerDataLoadHangarInventoryResult{}, err
	}
	if result.ErrorCode != "" {
		return result, errors.New("player-data runtime rejected inventory load")
	}
	return result, nil
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

func (c *RuntimeClient) handle(command any) ([]byte, error) {
	payload, err := codec.Encode(command)
	if err != nil {
		return nil, err
	}
	return c.sink.HandlePlayerDataCommand(payload)
}
