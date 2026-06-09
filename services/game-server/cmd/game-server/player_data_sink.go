package main

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
)

type hostedPlayerDataSink struct {
	runtime *playerdata.Runtime
}

func (s *hostedPlayerDataSink) handlePlayerDataCommand(payload []byte) ([]byte, error) {
	if s.runtime == nil {
		return nil, errors.New("player-data runtime is required")
	}
	return s.runtime.Handle(payload)
}
