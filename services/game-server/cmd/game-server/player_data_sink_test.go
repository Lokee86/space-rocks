package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		Context:    protocol.PlayerDataRequestContext{PlayMode: playerDataProfilePlayModeMultiplayer},
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
		Context:  protocol.PlayerDataRequestContext{PlayMode: playerDataProfilePlayModeMultiplayer},
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
		Mode:    serverplayerdata.MatchModeMultiplayer,
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				AccountID:    "acct-1",
				Score:        11,
			},
		},
	}

	if err := reporter.ReportMatchResult(summary); err != nil {
		t.Fatalf("ReportMatchResult returned error: %v", err)
	}
}

func TestHostedPlayerDataSinkGuestWriteFeedsPlayerDataProfileRead(t *testing.T) {
	runtime, err := playerdata.NewInMemoryRuntime()
	if err != nil {
		t.Fatalf("NewInMemoryRuntime returned error: %v", err)
	}

	sink := &hostedPlayerDataSink{runtime: runtime}

	seedPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "guest-result-1",
		MatchID:  "guest-match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: playerdata.IdentityKindGuest,
		},
		Context: protocol.PlayerDataRequestContext{
			PlayMode: playerDataProfilePlayModeSinglePlayer,
		},
		Score:      4160,
		ShipDeaths: 3,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("encode seed payload: %v", err)
	}
	if _, err := sink.HandlePlayerDataCommand(seedPayload); err != nil {
		t.Fatalf("HandlePlayerDataCommand returned error: %v", err)
	}

	handler := newPlayerDataProfileHandler(runtime, nil)
	requestBody := `{"play_mode":"single_player","identity_kind":"guest"}`
	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(requestBody))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body playerDataProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.Stats.GamesPlayed != 1 {
		t.Fatalf("games_played = %d, want %d", body.Profile.Stats.GamesPlayed, 1)
	}
	if body.Profile.Stats.TotalScore != 4160 {
		t.Fatalf("total_score = %d, want %d", body.Profile.Stats.TotalScore, 4160)
	}
	if body.Profile.Stats.HighScore != 4160 {
		t.Fatalf("high_score = %d, want %d", body.Profile.Stats.HighScore, 4160)
	}
	if body.Profile.Stats.ShipDeaths != 3 {
		t.Fatalf("ship_deaths = %d, want %d", body.Profile.Stats.ShipDeaths, 3)
	}
	if body.Profile.Stats.Wins != 0 {
		t.Fatalf("wins = %d, want %d", body.Profile.Stats.Wins, 0)
	}
}
