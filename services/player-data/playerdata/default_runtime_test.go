package playerdata

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestNewInMemoryRuntimeRoutesAccountMatchResult(t *testing.T) {
	runtime, err := NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}

	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      14,
		ShipDeaths: 3,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	if _, err := runtime.Handle(payload); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	loadPayload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode load payload: %v", err)
	}

	response, err := runtime.Handle(loadPayload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !packet.Found {
		t.Fatal("Found = false, want true")
	}
	if packet.Stats.TotalScore != 14 || packet.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want account stats", packet.Stats)
	}
}

func TestNewInMemoryRuntimeRoutesLocalMatchResult(t *testing.T) {
	runtime, err := NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}

	identity := protocol.PlayerDataIdentity{
		IdentityKind:   IdentityKindLocalProfile,
		LocalProfileID: "local-123",
	}

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      7,
		ShipDeaths: 2,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	if _, err := runtime.Handle(payload); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	loadPayload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode load payload: %v", err)
	}

	response, err := runtime.Handle(loadPayload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var packet protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(response, &packet); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !packet.Found {
		t.Fatal("Found = false, want true")
	}
	if packet.Stats.TotalScore != 7 || packet.Stats.ShipDeaths != 2 {
		t.Fatalf("Stats = %+v, want local stats", packet.Stats)
	}
}

func TestNewInMemoryRuntimeRoutesGuestMatchResult(t *testing.T) {
	runtime, err := NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}

	identity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindGuest,
	}

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "guest-result",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      5,
		ShipDeaths: 1,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := runtime.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var recordPacket protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(response, &recordPacket); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !recordPacket.Accepted {
		t.Fatal("Accepted = false, want true")
	}
	if recordPacket.Duplicate {
		t.Fatal("Duplicate = true, want false")
	}
	if recordPacket.Stats.TotalScore != 5 || recordPacket.Stats.HighScore != 5 || recordPacket.Stats.ShipDeaths != 1 || recordPacket.Stats.GamesPlayed != 1 || recordPacket.Stats.Wins != 1 {
		t.Fatalf("Record stats = %+v, want guest stats", recordPacket.Stats)
	}

	loadPayload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode load payload: %v", err)
	}

	loadResponse, err := runtime.Handle(loadPayload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var loadPacket protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(loadResponse, &loadPacket); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !loadPacket.Found {
		t.Fatal("Found = false, want true")
	}
	if loadPacket.Stats.TotalScore != 5 || loadPacket.Stats.HighScore != 5 || loadPacket.Stats.ShipDeaths != 1 || loadPacket.Stats.GamesPlayed != 1 || loadPacket.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want guest stats", loadPacket.Stats)
	}
}
