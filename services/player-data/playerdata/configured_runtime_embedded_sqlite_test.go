//go:build embedded_sqlite

package playerdata_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/playerdata/embeddedsqlite"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func newTestEmbeddedSQLiteLocalStoreFactory(t *testing.T) playerdata.LocalStoreFactory {
	t.Helper()

	return func(path string) (playerdata.Store, error) {
		store, err := embeddedsqlite.New(embeddedsqlite.Config{Path: path})
		if err != nil {
			return nil, err
		}
		if err := store.InitSchema(); err != nil {
			_ = store.Close()
			return nil, err
		}
		return store, nil
	}
}

func TestNewConfiguredRuntimePersistsLocalProfileStatsWithEmbeddedSQLite(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "player-data.sqlite")

	runtimeOne, err := playerdata.NewConfiguredRuntime(playerdata.RuntimeConfig{
		SQLitePath:        dbPath,
		LocalStoreFactory: newTestEmbeddedSQLiteLocalStoreFactory(t),
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}

	recordPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "local-123",
		},
		Context:    protocol.PlayerDataRequestContext{PlayMode: playerdata.PlayModeSinglePlayer},
		Score:      8,
		ShipDeaths: 2,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode record payload: %v", err)
	}

	recordResponse, err := runtimeOne.Handle(recordPayload)
	if err != nil {
		t.Fatalf("runtimeOne Handle returned error: %v", err)
	}

	var recordPacket protocol.PlayerDataRecordMatchResultResult
	if err := json.Unmarshal(recordResponse, &recordPacket); err != nil {
		t.Fatalf("unmarshal record response: %v", err)
	}
	if !recordPacket.Accepted {
		t.Fatal("Accepted = false, want true")
	}
	if recordPacket.Duplicate {
		t.Fatal("Duplicate = true, want false")
	}
	if recordPacket.Stats.TotalScore != 8 || recordPacket.Stats.HighScore != 8 || recordPacket.Stats.ShipDeaths != 2 || recordPacket.Stats.GamesPlayed != 1 || recordPacket.Stats.Wins != 0 {
		t.Fatalf("Stats = %+v, want persisted sqlite stats", recordPacket.Stats)
	}

	runtimeTwo, err := playerdata.NewConfiguredRuntime(playerdata.RuntimeConfig{
		SQLitePath:        dbPath,
		LocalStoreFactory: newTestEmbeddedSQLiteLocalStoreFactory(t),
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error for reopen: %v", err)
	}

	loadPayload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type: protocol.PacketTypePlayerDataLoadStats,
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "local-123",
		},
		Context: protocol.PlayerDataRequestContext{PlayMode: playerdata.PlayModeSinglePlayer},
	})
	if err != nil {
		t.Fatalf("encode load payload: %v", err)
	}

	loadResponse, err := runtimeTwo.Handle(loadPayload)
	if err != nil {
		t.Fatalf("runtimeTwo Handle returned error: %v", err)
	}

	var loadPacket protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(loadResponse, &loadPacket); err != nil {
		t.Fatalf("unmarshal load response: %v", err)
	}
	if !loadPacket.Found {
		t.Fatal("Found = false, want true")
	}
	if loadPacket.Stats.TotalScore != 8 || loadPacket.Stats.HighScore != 8 || loadPacket.Stats.ShipDeaths != 2 || loadPacket.Stats.GamesPlayed != 1 || loadPacket.Stats.Wins != 0 {
		t.Fatalf("Stats = %+v, want persisted sqlite stats", loadPacket.Stats)
	}
}
