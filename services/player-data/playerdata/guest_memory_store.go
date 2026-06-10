package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type GuestMemoryStore struct {
	stats              protocol.PlayerDataStats
	processedResultIDs map[string]struct{}
}

func NewGuestMemoryStore() *GuestMemoryStore {
	return &GuestMemoryStore{
		processedResultIDs: make(map[string]struct{}),
	}
}

func (s *GuestMemoryStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	if identity.IdentityKind != IdentityKindGuest {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be guest")
	}

	return s.stats, true, nil
}

func (s *GuestMemoryStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	if command.Identity.IdentityKind != IdentityKindGuest {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be guest")
	}
	if command.ResultID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("result_id is required")
	}

	if _, duplicate := s.processedResultIDs[command.ResultID]; duplicate {
		return s.stats, true, nil
	}

	s.stats.GamesPlayed += 1
	s.stats.TotalScore += command.Score
	if command.Score > s.stats.HighScore {
		s.stats.HighScore = command.Score
	}
	s.stats.ShipDeaths += command.ShipDeaths
	if command.Won {
		s.stats.Wins += 1
	}

	s.processedResultIDs[command.ResultID] = struct{}{}

	return s.stats, false, nil
}
