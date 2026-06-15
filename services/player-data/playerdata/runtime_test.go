package playerdata

import (
	"bytes"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type runtimeLoadStatsStore struct {
	loadStatsCalls int
	loadStatsStats  protocol.PlayerDataStats
	loadStatsFound  bool
	loadStatsErr    error
}

func (s *runtimeLoadStatsStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	s.loadStatsCalls++
	return s.loadStatsStats, s.loadStatsFound, s.loadStatsErr
}

func (s *runtimeLoadStatsStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, nil
}

func (s *runtimeLoadStatsStore) ListLocalProfiles() ([]LocalProfileSummary, error) {
	return nil, ErrLocalProfileUnavailable
}

func (s *runtimeLoadStatsStore) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error) {
	return LocalProfileSummary{}, ErrLocalProfileUnavailable
}

func (s *runtimeLoadStatsStore) DeleteLocalProfile(localProfileID string) error {
	return ErrLocalProfileUnavailable
}

func (s *runtimeLoadStatsStore) UpdateLocalProfileDisplayName(localProfileID string, displayName string) (LocalProfileSummary, error) {
	return LocalProfileSummary{}, ErrLocalProfileUnavailable
}

func (s *runtimeLoadStatsStore) GetDefaultLocalProfile() (LocalProfileDefault, error) {
	return LocalProfileDefault{}, ErrLocalProfileUnavailable
}

func (s *runtimeLoadStatsStore) SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error) {
	return LocalProfileDefault{}, ErrLocalProfileUnavailable
}

func TestNewRuntimeRejectsNilStore(t *testing.T) {
	if _, err := NewRuntime(Config{}); err == nil {
		t.Fatal("NewRuntime returned nil error for nil store")
	}
}

func TestRuntimeHandleDelegatesLoadStats(t *testing.T) {
	store := NewMemoryStore()
	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}
	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      8,
		ShipDeaths: 1,
		Won:        true,
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	runtime, err := NewRuntime(Config{Store: store})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer},
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	got, err := runtime.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	want, err := NewDispatcher(store).Handle(payload)
	if err != nil {
		t.Fatalf("dispatcher handle returned error: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("Handle() = %s, want %s", got, want)
	}
}

func TestRuntimeLoadStatsDelegatesToStore(t *testing.T) {
	store := &runtimeLoadStatsStore{
		loadStatsStats: protocol.PlayerDataStats{
			GamesPlayed: 4,
			TotalScore:  12,
			HighScore:   9,
			ShipDeaths:  2,
			Wins:        1,
		},
		loadStatsFound: true,
	}

	runtime, err := NewRuntime(Config{Store: store})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}

	gotStats, gotFound, gotErr := runtime.LoadStats(identity)
	if gotErr != nil {
		t.Fatalf("LoadStats returned error: %v", gotErr)
	}
	if store.loadStatsCalls != 1 {
		t.Fatalf("LoadStats calls = %d, want 1", store.loadStatsCalls)
	}
	if !gotFound {
		t.Fatal("LoadStats found = false, want true")
	}
	if gotStats != store.loadStatsStats {
		t.Fatalf("LoadStats stats = %+v, want %+v", gotStats, store.loadStatsStats)
	}
}

func TestRuntimeLocalProfileSeedStatsFalseReturnsZeroStats(t *testing.T) {
	var runtime *Runtime

	gotStats, gotErr := runtime.LocalProfileSeedStats(false)
	if gotErr != nil {
		t.Fatalf("LocalProfileSeedStats returned error: %v", gotErr)
	}
	if gotStats != (protocol.PlayerDataStats{}) {
		t.Fatalf("LocalProfileSeedStats stats = %+v, want zero stats", gotStats)
	}
}

func TestRuntimeLocalProfileSeedStatsTrueReturnsGuestStats(t *testing.T) {
	store := NewGuestMemoryStore()
	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "guest-result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindGuest,
		},
		Score:      7,
		ShipDeaths: 2,
		Won:        true,
	}); err != nil {
		t.Fatalf("seed guest stats: %v", err)
	}

	runtime, err := NewRuntime(Config{Store: store})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	gotStats, gotErr := runtime.LocalProfileSeedStats(true)
	if gotErr != nil {
		t.Fatalf("LocalProfileSeedStats returned error: %v", gotErr)
	}
	want := protocol.PlayerDataStats{
		GamesPlayed: 1,
		TotalScore:  7,
		HighScore:   7,
		ShipDeaths:  2,
		Wins:        1,
	}
	if gotStats != want {
		t.Fatalf("LocalProfileSeedStats stats = %+v, want %+v", gotStats, want)
	}
}

func TestRuntimeLocalProfileSeedStatsTrueReturnsErrorWhenGuestStatsMissing(t *testing.T) {
	runtime, err := NewRuntime(Config{Store: NewMemoryStore()})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	gotStats, gotErr := runtime.LocalProfileSeedStats(true)
	if gotErr == nil {
		t.Fatal("LocalProfileSeedStats returned nil error, want guest stats unavailable")
	}
	if gotErr.Error() != "guest stats unavailable" {
		t.Fatalf("LocalProfileSeedStats error = %v, want guest stats unavailable", gotErr)
	}
	if gotStats != (protocol.PlayerDataStats{}) {
		t.Fatalf("LocalProfileSeedStats stats = %+v, want zero stats", gotStats)
	}
}
