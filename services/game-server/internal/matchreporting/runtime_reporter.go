package matchreporting

import (
	"encoding/json"
	"errors"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	serverplayerdata "github.com/Lokee86/space-rocks/server/internal/playerdata"
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
