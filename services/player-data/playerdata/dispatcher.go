package playerdata

import (
	"encoding/json"
	"fmt"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
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
			emitPacketEvent(observability.EventNamePlayerDataReadFailed, packetType, packet.Context, packet.Identity, "", "", observability.Fields{"failure_mode": "invalid_mode_identity", "error_code": "invalid_mode_identity"})
			return codec.Encode(protocol.PlayerDataLoadStatsResult{
				Type: protocol.PacketTypePlayerDataLoadStatsResult, Found: false,
				Stats: protocol.PlayerDataStats{}, ErrorCode: "invalid_mode_identity", Message: err.Error(),
			})
		}
		stats, found, storeErr := d.store.LoadStats(packet.Identity)
		if storeErr != nil {
			emitPacketEvent(observability.EventNamePlayerDataReadFailed, packetType, packet.Context, packet.Identity, "", "", observability.Fields{"failure_mode": "store_failure", "error_code": failureCode(storeErr)})
			return codec.Encode(protocol.PlayerDataLoadStatsResult{
				Type: protocol.PacketTypePlayerDataLoadStatsResult, Found: false,
				Stats: protocol.PlayerDataStats{}, ErrorCode: "store_error", Message: storeErr.Error(),
			})
		}
		return codec.Encode(protocol.PlayerDataLoadStatsResult{Type: protocol.PacketTypePlayerDataLoadStatsResult, Found: found, Stats: stats})
	case protocol.PacketTypePlayerDataRecordMatchResult:
		var packet protocol.PlayerDataRecordMatchResult
		if err := json.Unmarshal(payload, &packet); err != nil {

			return nil, err
		}
		if err := ValidateModeIdentity(packet.Context.PlayMode, packet.Identity); err != nil {
			emitPacketEvent(observability.EventNamePlayerDataWriteFailed, packetType, packet.Context, packet.Identity, packet.MatchID, packet.ResultID, observability.Fields{"failure_mode": "invalid_mode_identity", "error_code": "invalid_mode_identity"})
			return codec.Encode(protocol.PlayerDataRecordMatchResultResult{
				Type: protocol.PacketTypePlayerDataRecordMatchResultResult, Accepted: false,
				Duplicate: false, Stats: protocol.PlayerDataStats{}, ErrorCode: "invalid_mode_identity", Message: err.Error(),
			})
		}
		stats, duplicate, storeErr := d.store.RecordMatchResult(packet)
		if storeErr != nil {
			emitPacketEvent(observability.EventNamePlayerDataWriteFailed, packetType, packet.Context, packet.Identity, packet.MatchID, packet.ResultID, observability.Fields{"failure_mode": "store_failure", "error_code": failureCode(storeErr)})
			return codec.Encode(protocol.PlayerDataRecordMatchResultResult{
				Type: protocol.PacketTypePlayerDataRecordMatchResultResult, Accepted: false,
				Duplicate: false, Stats: protocol.PlayerDataStats{}, ErrorCode: "store_error", Message: storeErr.Error(),
			})
		}
		if duplicate {
			emitPacketEvent(observability.EventNameMatchResultDuplicateSuppressed, packetType, packet.Context, packet.Identity, packet.MatchID, packet.ResultID, observability.Fields{"duplicate": true})
		} else {
			emitPacketEvent(observability.EventNameMatchResultReportSucceeded, packetType, packet.Context, packet.Identity, packet.MatchID, packet.ResultID, nil)
		}
		return codec.Encode(protocol.PlayerDataRecordMatchResultResult{Type: protocol.PacketTypePlayerDataRecordMatchResultResult, Accepted: true, Duplicate: duplicate, Stats: stats})
	default:
		return nil, fmt.Errorf("unknown packet type %q", packetType)
	}
}

func emitPacketEvent(event observability.EventName, packetType string, requestContext protocol.PlayerDataRequestContext, identity protocol.PlayerDataIdentity, matchID, resultID string, fields observability.Fields) {
	context := observability.Context{TraceID: requestContext.TraceID, MatchID: matchID, AccountID: identity.AccountID, PacketType: packetType}
	if resultID != "" {
		if fields == nil {
			fields = observability.Fields{}
		}
		fields["result_id"] = resultID
	}
	logging.Emit(observability.Request{Event: event, Context: context, Fields: fields})
}

func failureCode(err error) string {
	if class := FailureClassOf(err); class != "" {
		return string(class)
	}
	return "operation_failed"
}
