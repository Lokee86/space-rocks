package main

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/server/internal/matchreporting"
	serverplayerdata "github.com/Lokee86/space-rocks/server/internal/playerdata"
)

func TestHostedPlayerDataSinkHandlePlayerDataCommandNilRuntime(t *testing.T) {
	sink := &hostedPlayerDataSink{}

	if _, err := sink.HandlePlayerDataCommand([]byte(`{"type":"player_data_load_stats"}`)); err == nil {
		t.Fatal("HandlePlayerDataCommand returned nil error for nil runtime")
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
	if _, err := sink.HandlePlayerDataCommand(seedPayload); err != nil {
		t.Fatalf("seed HandlePlayerDataCommand returned error: %v", err)
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: identity,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := sink.HandlePlayerDataCommand(payload)
	if err != nil {
		t.Fatalf("HandlePlayerDataCommand returned error: %v", err)
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

func TestHostedPlayerDataSinkSupportsRuntimeReporter(t *testing.T) {
	runtime, err := playerdata.NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}

	sink := &hostedPlayerDataSink{runtime: runtime}
	reporter, err := matchreporting.NewRuntimeReporter(sink)
	if err != nil {
		t.Fatalf("NewRuntimeReporter returned error: %v", err)
	}

	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				Score:        11,
			},
		},
	}

	if err := reporter.ReportMatchResult(summary); err != nil {
		t.Fatalf("ReportMatchResult returned error: %v", err)
	}
}
