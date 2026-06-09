package main

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestHostedPlayerDataSinkHandlePlayerDataCommandNilRuntime(t *testing.T) {
	sink := &hostedPlayerDataSink{}

	if _, err := sink.handlePlayerDataCommand([]byte(`{"type":"player_data_load_stats"}`)); err == nil {
		t.Fatal("handlePlayerDataCommand returned nil error for nil runtime")
	}
}

func TestHostedPlayerDataSinkHandlePlayerDataCommandLoadStats(t *testing.T) {
	runtime, err := playerdata.NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}
	sink := &hostedPlayerDataSink{runtime: runtime}

	identity := protocol.PlayerDataIdentity{
		IdentityKind: playerdata.IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}

	seedPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "result-1",
		MatchID:    "match-1",
		Identity:   identity,
		Score:      4,
		ShipDeaths: 1,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode seed payload: %v", err)
	}
	if _, err := sink.handlePlayerDataCommand(seedPayload); err != nil {
		t.Fatalf("seed handlePlayerDataCommand returned error: %v", err)
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := sink.handlePlayerDataCommand(payload)
	if err != nil {
		t.Fatalf("handlePlayerDataCommand returned error: %v", err)
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
	if packet.Stats.TotalScore != 4 || packet.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want seeded stats", packet.Stats)
	}
}
