package playerdata

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func testRequestContext(playMode string) protocol.PlayerDataRequestContext {
	return protocol.PlayerDataRequestContext{PlayMode: playMode}
}

type countingStore struct {
	loadStatsCalls          int
	recordMatchResultCalls  int
	loadStatsResult         protocol.PlayerDataStats
	loadStatsFound          bool
	recordMatchResultResult protocol.PlayerDataStats
	recordMatchResultDup    bool
}

func (s *countingStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	s.loadStatsCalls++
	return s.loadStatsResult, s.loadStatsFound, nil
}

func (s *countingStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	s.recordMatchResultCalls++
	return s.recordMatchResultResult, s.recordMatchResultDup, nil
}

func TestDispatcherHandle(t *testing.T) {
	dispatcher := NewDispatcher(NewNoopStore())

	t.Run("malformed json", func(t *testing.T) {
		if _, err := dispatcher.Handle([]byte(`{"type":`)); err == nil {
			t.Fatal("Handle returned nil error for malformed JSON")
		}
	})

	t.Run("unknown type", func(t *testing.T) {
		if _, err := dispatcher.Handle([]byte(`{"type":"unknown_packet"}`)); err == nil {
			t.Fatal("Handle returned nil error for unknown packet type")
		}
	})
}

func TestDispatcherHandleLoadStats(t *testing.T) {
	store := NewMemoryStore()
	dispatcher := NewDispatcher(store)

	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}
	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      9,
		ShipDeaths: 3,
		Won:        true,
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
		Context:  testRequestContext(PlayModeMultiplayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Type != protocol.PacketTypePlayerDataLoadStatsResult {
		t.Fatalf("Type = %q, want %q", packet.Type, protocol.PacketTypePlayerDataLoadStatsResult)
	}
	if !packet.Found {
		t.Fatal("Found = false, want true")
	}
	if packet.Stats.TotalScore != 9 || packet.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want seeded stats", packet.Stats)
	}
}

func TestDispatcherHandleLoadStatsInvalidIdentity(t *testing.T) {
	dispatcher := NewDispatcher(NewMemoryStore())

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type: protocol.PacketTypePlayerDataLoadStats,
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
		},
		Context: testRequestContext(PlayModeMultiplayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Type != protocol.PacketTypePlayerDataLoadStatsResult {
		t.Fatalf("Type = %q, want %q", packet.Type, protocol.PacketTypePlayerDataLoadStatsResult)
	}
	if packet.Found {
		t.Fatal("Found = true, want false")
	}
	if packet.ErrorCode == "" {
		t.Fatal("ErrorCode is empty, want error result packet")
	}
}

func TestDispatcherHandleLoadStatsInvalidModeIdentityRejectsBeforeStore(t *testing.T) {
	store := &countingStore{}
	dispatcher := NewDispatcher(store)

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type: protocol.PacketTypePlayerDataLoadStats,
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		},
		Context: testRequestContext(PlayModeSinglePlayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Found {
		t.Fatal("Found = true, want false")
	}
	if packet.ErrorCode != "invalid_mode_identity" {
		t.Fatalf("ErrorCode = %q, want %q", packet.ErrorCode, "invalid_mode_identity")
	}
	if store.loadStatsCalls != 0 {
		t.Fatalf("LoadStats calls = %d, want 0", store.loadStatsCalls)
	}
	if store.recordMatchResultCalls != 0 {
		t.Fatalf("RecordMatchResult calls = %d, want 0", store.recordMatchResultCalls)
	}
}

func TestDispatcherHandleRecordMatchResult(t *testing.T) {
	store := NewMemoryStore()
	dispatcher := NewDispatcher(store)

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"},
		Context:    testRequestContext(PlayModeMultiplayer),
		Score:      11,
		ShipDeaths: 2,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Type != protocol.PacketTypePlayerDataRecordMatchResultResult {
		t.Fatalf("Type = %q, want %q", packet.Type, protocol.PacketTypePlayerDataRecordMatchResultResult)
	}
	if !packet.Accepted {
		t.Fatal("Accepted = false, want true")
	}
	if packet.Duplicate {
		t.Fatal("Duplicate = true, want false")
	}
	if packet.Stats.TotalScore != 11 || packet.Stats.HighScore != 11 || packet.Stats.ShipDeaths != 2 || packet.Stats.GamesPlayed != 1 || packet.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want seeded stats", packet.Stats)
	}
}

func TestDispatcherHandleRecordMatchResultDuplicate(t *testing.T) {
	dispatcher := NewDispatcher(NewMemoryStore())

	firstPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"},
		Context:    testRequestContext(PlayModeMultiplayer),
		Score:      11,
		ShipDeaths: 2,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	if _, err := dispatcher.Handle(firstPayload); err != nil {
		t.Fatalf("seed first result: %v", err)
	}

	duplicatePayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"},
		Context:    testRequestContext(PlayModeMultiplayer),
		Score:      99,
		ShipDeaths: 9,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("encode duplicate payload: %v", err)
	}

	response, err := dispatcher.Handle(duplicatePayload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !packet.Accepted {
		t.Fatal("Accepted = false, want true")
	}
	if !packet.Duplicate {
		t.Fatal("Duplicate = false, want true")
	}
	if packet.Stats.TotalScore != 11 || packet.Stats.GamesPlayed != 1 || packet.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want unchanged stats", packet.Stats)
	}
}

func TestDispatcherHandleRecordMatchResultInvalidIdentity(t *testing.T) {
	dispatcher := NewDispatcher(NewMemoryStore())

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
		},
		Context: testRequestContext(PlayModeMultiplayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Accepted {
		t.Fatal("Accepted = true, want false")
	}
	if packet.ErrorCode == "" {
		t.Fatal("ErrorCode is empty, want error result packet")
	}
}

func TestDispatcherHandleRecordMatchResultInvalidModeIdentityRejectsBeforeStore(t *testing.T) {
	store := &countingStore{}
	dispatcher := NewDispatcher(store)

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindGuest,
		},
		Context: testRequestContext(PlayModeMultiplayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Accepted {
		t.Fatal("Accepted = true, want false")
	}
	if packet.Duplicate {
		t.Fatal("Duplicate = true, want false")
	}
	if packet.ErrorCode != "invalid_mode_identity" {
		t.Fatalf("ErrorCode = %q, want %q", packet.ErrorCode, "invalid_mode_identity")
	}
	if store.loadStatsCalls != 0 {
		t.Fatalf("LoadStats calls = %d, want 0", store.loadStatsCalls)
	}
	if store.recordMatchResultCalls != 0 {
		t.Fatalf("RecordMatchResult calls = %d, want 0", store.recordMatchResultCalls)
	}
}

func TestDispatcherHandleRecordMatchResultMissingResultID(t *testing.T) {
	dispatcher := NewDispatcher(NewMemoryStore())

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type: protocol.PacketTypePlayerDataRecordMatchResult,
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		},
		Context: testRequestContext(PlayModeMultiplayer),
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := dispatcher.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if packet.Accepted {
		t.Fatal("Accepted = true, want false")
	}
	if packet.ErrorCode == "" {
		t.Fatal("ErrorCode is empty, want error result packet")
	}
}
