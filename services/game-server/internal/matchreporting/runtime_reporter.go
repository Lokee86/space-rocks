package matchreporting

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	serverplayerdata "github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type PlayerDataSink interface {
	HandlePlayerDataCommand(payload []byte) ([]byte, error)
}

type RuntimeReporter struct {
	sink PlayerDataSink
}

func NewRuntimeReporter(sink PlayerDataSink) (*RuntimeReporter, error) {
	if sink == nil {
		return nil, errors.New("player-data sink is required")
	}

	return &RuntimeReporter{sink: sink}, nil
}

func (r *RuntimeReporter) ReportMatchResult(summary serverplayerdata.MatchResultSummary) error {
	commands := BuildRecordMatchResultCommands(summary)
	for _, command := range commands {
		payload, err := codec.Encode(command)
		if err != nil {
			return err
		}

		response, err := r.sink.HandlePlayerDataCommand(payload)
		if err != nil {
			emitPlayerDataPacketBoundaryEvent(command, err)
			return err
		}

		var recordResult protocol.PlayerDataRecordMatchResultResult
		if err := json.Unmarshal(response, &recordResult); err != nil {
			return err
		}
		if !recordResult.Accepted {
			return errors.New("player-data runtime rejected match result")
		}
	}

	return nil
}

func emitPlayerDataPacketBoundaryEvent(command protocol.PlayerDataRecordMatchResult, err error) {
	event := observability.EventNamePacketDecodeFailed
	failureMode := "packet_decode"
	if strings.Contains(err.Error(), "unknown packet type") {
		event = observability.EventNamePacketRouteUnknown
		failureMode = "unknown_packet_type"
	}
	logging.Emit(observability.Request{
		Event: event,
		Context: observability.Context{
			TraceID:    command.Context.TraceID,
			MatchID:    command.MatchID,
			AccountID:  command.Identity.AccountID,
			PacketType: command.Type,
		},
		Fields: observability.Fields{
			"failure_mode": failureMode,
			"error_code":   failureMode,
			"result_id":    command.ResultID,
		},
	})
}
