package playerdata

import "github.com/Lokee86/space-rocks/player-data/protocol"

type NoopStore struct{}

func NewNoopStore() *NoopStore {
	return &NoopStore{}
}

func (s *NoopStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, nil
}

func (s *NoopStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, nil
}
