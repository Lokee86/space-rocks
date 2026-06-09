package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type MemoryStore struct {
	statsByIdentityKey map[string]protocol.PlayerDataStats
	processedResultIDs map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		statsByIdentityKey: make(map[string]protocol.PlayerDataStats),
		processedResultIDs: make(map[string]string),
	}
}

func (s *MemoryStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	identityKey := IdentityKey(identity)
	if identityKey == "" {
		return protocol.PlayerDataStats{}, false, errors.New("invalid identity")
	}

	stats, found := s.statsByIdentityKey[identityKey]
	if !found {
		return protocol.PlayerDataStats{}, false, nil
	}

	return stats, true, nil
}

func (s *MemoryStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	identityKey := IdentityKey(command.Identity)
	if identityKey == "" {
		return protocol.PlayerDataStats{}, false, errors.New("invalid identity")
	}
	if command.ResultID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("result_id is required")
	}

	if storedIdentityKey, duplicate := s.processedResultIDs[command.ResultID]; duplicate {
		stats, _ := s.statsByIdentityKey[storedIdentityKey]
		return stats, true, nil
	}

	stats := s.statsByIdentityKey[identityKey]
	stats.GamesPlayed += 1
	stats.TotalScore += command.Score
	if command.Score > stats.HighScore {
		stats.HighScore = command.Score
	}
	stats.ShipDeaths += command.ShipDeaths
	if command.Won {
		stats.Wins += 1
	}

	s.statsByIdentityKey[identityKey] = stats
	s.processedResultIDs[command.ResultID] = identityKey

	return stats, false, nil
}
