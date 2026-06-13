package playerdata

import (
	"encoding/json"
	"fmt"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type Dispatcher struct {
	store Store
}

func NewDispatcher(store Store) *Dispatcher {
	return &Dispatcher{store: store}
}

func (d *Dispatcher) Handle(payload []byte) ([]byte, error) {
	packetType, err := codec.DecodeType(payload)
	if err != nil {
		return nil, err
	}

	switch packetType {
	case protocol.PacketTypePlayerDataLoadStats:
		var packet protocol.PlayerDataLoadStats
		if err := json.Unmarshal(payload, &packet); err != nil {
			return nil, err
		}
		if err := ValidateModeIdentity(packet.Context.PlayMode, packet.Identity); err != nil {
			return codec.Encode(protocol.PlayerDataLoadStatsResult{
				Type:      protocol.PacketTypePlayerDataLoadStatsResult,
				Found:     false,
				Stats:     protocol.PlayerDataStats{},
				ErrorCode: "invalid_mode_identity",
				Message:   err.Error(),
			})
		}
		stats, found, storeErr := d.store.LoadStats(packet.Identity)
		if storeErr != nil {
			return codec.Encode(protocol.PlayerDataLoadStatsResult{
				Type:      protocol.PacketTypePlayerDataLoadStatsResult,
				Found:     false,
				Stats:     protocol.PlayerDataStats{},
				ErrorCode: "store_error",
				Message:   storeErr.Error(),
			})
		}
		return codec.Encode(protocol.PlayerDataLoadStatsResult{
			Type:  protocol.PacketTypePlayerDataLoadStatsResult,
			Found: found,
			Stats: stats,
		})
	case protocol.PacketTypePlayerDataRecordMatchResult:
		var packet protocol.PlayerDataRecordMatchResult
		if err := json.Unmarshal(payload, &packet); err != nil {
			return nil, err
		}
		if err := ValidateModeIdentity(packet.Context.PlayMode, packet.Identity); err != nil {
			return codec.Encode(protocol.PlayerDataRecordMatchResultResult{
				Type:      protocol.PacketTypePlayerDataRecordMatchResultResult,
				Accepted:  false,
				Duplicate: false,
				Stats:     protocol.PlayerDataStats{},
				ErrorCode: "invalid_mode_identity",
				Message:   err.Error(),
			})
		}
		stats, duplicate, storeErr := d.store.RecordMatchResult(packet)
		if storeErr != nil {
			return codec.Encode(protocol.PlayerDataRecordMatchResultResult{
				Type:      protocol.PacketTypePlayerDataRecordMatchResultResult,
				Accepted:  false,
				Duplicate: false,
				Stats:     protocol.PlayerDataStats{},
				ErrorCode: "store_error",
				Message:   storeErr.Error(),
			})
		}
		return codec.Encode(protocol.PlayerDataRecordMatchResultResult{
			Type:      protocol.PacketTypePlayerDataRecordMatchResultResult,
			Accepted:  true,
			Duplicate: duplicate,
			Stats:     stats,
		})
	default:
		return nil, fmt.Errorf("unknown packet type %q", packetType)
	}
}
