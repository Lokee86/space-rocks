package playerdata

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func newTestMemoryLocalStoreFactory(t *testing.T) LocalStoreFactory {
	t.Helper()

	return func(path string) (Store, error) {
		return NewMemoryStore(), nil
	}
}

func TestNewConfiguredRuntimeDefaultsToNoopLocalStore(t *testing.T) {
	runtime, err := NewConfiguredRuntime(RuntimeConfig{})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("NewConfiguredRuntime returned nil runtime")
	}

	runtimeStore, ok := runtime.dispatcher.store.(*StoreRouter)
	if !ok {
		t.Fatalf("dispatcher.store type = %T, want *StoreRouter", runtime.dispatcher.store)
	}

	if _, ok := runtimeStore.accountStore.(*MemoryStore); !ok {
		t.Fatalf("accountStore type = %T, want *MemoryStore", runtimeStore.accountStore)
	}
	if _, ok := runtimeStore.localStore.(*NoopStore); !ok {
		t.Fatalf("localStore type = %T, want *NoopStore", runtimeStore.localStore)
	}
	if _, ok := runtimeStore.guestStore.(*GuestMemoryStore); !ok {
		t.Fatalf("guestStore type = %T, want *GuestMemoryStore", runtimeStore.guestStore)
	}
}

func TestNewRuntimeFromEnvUsesRailsEnvAndNoopLocalStore(t *testing.T) {
	gotKeys := make([]string, 0, 2)
	getenv := func(key string) string {
		gotKeys = append(gotKeys, key)
		switch key {
		case "PLAYER_DATA_RAILS_BASE_URL":
			return "https://example.test"
		case "PLAYER_DATA_RAILS_INTERNAL_TOKEN":
			return "test-internal-token"
		default:
			return ""
		}
	}

	runtime, err := NewRuntimeFromEnv(getenv)
	if err != nil {
		t.Fatalf("NewRuntimeFromEnv returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("NewRuntimeFromEnv returned nil runtime")
	}

	if len(gotKeys) != 2 {
		t.Fatalf("getenv called for %d keys, want 2", len(gotKeys))
	}
	if gotKeys[0] != "PLAYER_DATA_RAILS_BASE_URL" || gotKeys[1] != "PLAYER_DATA_RAILS_INTERNAL_TOKEN" {
		t.Fatalf("getenv keys = %v, want rails keys only", gotKeys)
	}

	runtimeStore, ok := runtime.dispatcher.store.(*StoreRouter)
	if !ok {
		t.Fatalf("dispatcher.store type = %T, want *StoreRouter", runtime.dispatcher.store)
	}
	if _, ok := runtimeStore.localStore.(*NoopStore); !ok {
		t.Fatalf("localStore type = %T, want *NoopStore", runtimeStore.localStore)
	}
}

func TestNewConfiguredRuntimeReturnsErrorWithoutLocalStoreFactory(t *testing.T) {
	runtime, err := NewConfiguredRuntime(RuntimeConfig{
		SQLitePath: ":memory:",
	})
	if err == nil {
		t.Fatal("NewConfiguredRuntime returned nil error, want missing local store factory error")
	}
	if runtime != nil {
		t.Fatalf("NewConfiguredRuntime returned runtime = %v, want nil", runtime)
	}
}

func TestNewConfiguredRuntimeUsesInjectedLocalStoreForSQLitePath(t *testing.T) {
	runtime, err := NewConfiguredRuntime(RuntimeConfig{
		SQLitePath:        ":memory:",
		LocalStoreFactory: newTestMemoryLocalStoreFactory(t),
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("NewConfiguredRuntime returned nil runtime")
	}

	runtimeStore, ok := runtime.dispatcher.store.(*StoreRouter)
	if !ok {
		t.Fatalf("dispatcher.store type = %T, want *StoreRouter", runtime.dispatcher.store)
	}

	if _, ok := runtimeStore.localStore.(*MemoryStore); !ok {
		t.Fatalf("localStore type = %T, want *MemoryStore", runtimeStore.localStore)
	}
	if _, ok := runtimeStore.guestStore.(*GuestMemoryStore); !ok {
		t.Fatalf("guestStore type = %T, want *GuestMemoryStore", runtimeStore.guestStore)
	}
}

func TestNewConfiguredRuntimeRoutesAuthenticatedAccountMatchResultThroughRails(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody struct {
		ResultID   string `json:"result_id"`
		MatchID    string `json:"match_id"`
		AccountID  string `json:"account_id"`
		Score      int    `json:"score"`
		ShipDeaths int    `json:"ship_deaths"`
		Won        bool   `json:"won"`
	}

	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accepted":true,"duplicate":false,"stats":{"total_score":14,"high_score":14,"ship_deaths":3,"games_played":1,"wins":1}}`))
	}))

	runtime, err := NewConfiguredRuntime(RuntimeConfig{
		RailsBaseURL:       server.URL,
		RailsInternalToken: "internal-token",
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}

	runtimeStore := runtime.dispatcher.store.(*StoreRouter)
	railsStore, ok := runtimeStore.accountStore.(*RailsStore)
	if !ok {
		t.Fatalf("accountStore type = %T, want *RailsStore", runtimeStore.accountStore)
	}
	railsStore.httpClient = server.Client()

	payload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		},
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer},
		Score:      14,
		ShipDeaths: 3,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	response, err := runtime.Handle(payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("Method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/internal/player-data/match-results" {
		t.Fatalf("Path = %q, want %q", gotPath, "/internal/player-data/match-results")
	}
	if gotBody.ResultID != "result-1" || gotBody.MatchID != "match-1" || gotBody.AccountID != "acct-123" || gotBody.Score != 14 || gotBody.ShipDeaths != 3 || !gotBody.Won {
		t.Fatalf("request body = %+v, want authenticated account payload", gotBody)
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
	if recordPacket.Stats.TotalScore != 14 || recordPacket.Stats.HighScore != 14 || recordPacket.Stats.ShipDeaths != 3 || recordPacket.Stats.GamesPlayed != 1 || recordPacket.Stats.Wins != 1 {
		t.Fatalf("Stats = %+v, want Rails stats", recordPacket.Stats)
	}
}

func TestNewConfiguredRuntimeRoutesLocalProfileMatchResultThroughInjectedStore(t *testing.T) {
	runtimeOne, err := NewConfiguredRuntime(RuntimeConfig{
		SQLitePath:         ":memory:",
		LocalStoreFactory:   newTestMemoryLocalStoreFactory(t),
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}

	recordPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   IdentityKindLocalProfile,
			LocalProfileID: "local-123",
		},
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
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
	if recordPacket.Stats.TotalScore != 8 || recordPacket.Stats.HighScore != 8 || recordPacket.Stats.ShipDeaths != 2 || recordPacket.Stats.GamesPlayed != 1 {
		t.Fatalf("Stats = %+v, want injected store stats", recordPacket.Stats)
	}
	if recordPacket.Stats.Wins != 1 {
		t.Fatalf("Stats.Wins = %d, want 1", recordPacket.Stats.Wins)
	}
}

func TestNewConfiguredRuntimeKeepsAccountLocalAndGuestStatsSeparate(t *testing.T) {
	var gotAccountBody struct {
		ResultID   string `json:"result_id"`
		MatchID    string `json:"match_id"`
		AccountID  string `json:"account_id"`
		Score      int    `json:"score"`
		ShipDeaths int    `json:"ship_deaths"`
		Won        bool   `json:"won"`
	}
	var gotLoadBody struct {
		AccountID string `json:"account_id"`
	}
	var gotLoadAuth string

	server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/player-data/match-results":
			if err := json.NewDecoder(r.Body).Decode(&gotAccountBody); err != nil {
				t.Fatalf("Decode account request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accepted":true,"duplicate":false,"stats":{"total_score":11,"high_score":11,"ship_deaths":1,"games_played":1,"wins":1}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/internal/player-data/stats":
			gotLoadAuth = r.Header.Get("Authorization")
			if err := json.NewDecoder(r.Body).Decode(&gotLoadBody); err != nil {
				t.Fatalf("Decode load request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"stats":{"total_score":11,"high_score":11,"ship_deaths":1,"games_played":1,"wins":1}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))

	runtime, err := NewConfiguredRuntime(RuntimeConfig{
		RailsBaseURL:       server.URL,
		RailsInternalToken: "internal-token",
		SQLitePath:         ":memory:",
		LocalStoreFactory:   newTestMemoryLocalStoreFactory(t),
	})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}

	runtimeStore := runtime.dispatcher.store.(*StoreRouter)
	railsStore, ok := runtimeStore.accountStore.(*RailsStore)
	if !ok {
		t.Fatalf("accountStore type = %T, want *RailsStore", runtimeStore.accountStore)
	}
	railsStore.httpClient = server.Client()

	accountIdentity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindAuthenticatedAccount,
		AccountID:    "acct-123",
	}
	localIdentity := protocol.PlayerDataIdentity{
		IdentityKind:   IdentityKindLocalProfile,
		LocalProfileID: "local-123",
	}
	guestIdentity := protocol.PlayerDataIdentity{
		IdentityKind: IdentityKindGuest,
	}

	accountPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "account-result",
		MatchID:    "match-1",
		Identity:   accountIdentity,
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer},
		Score:      11,
		ShipDeaths: 1,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode account payload: %v", err)
	}
	if _, err := runtime.Handle(accountPayload); err != nil {
		t.Fatalf("account Handle returned error: %v", err)
	}

	localPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "local-result",
		MatchID:    "match-2",
		Identity:   localIdentity,
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
		Score:      7,
		ShipDeaths: 2,
		Won:        false,
	})
	if err != nil {
		t.Fatalf("encode local payload: %v", err)
	}
	if _, err := runtime.Handle(localPayload); err != nil {
		t.Fatalf("local Handle returned error: %v", err)
	}

	guestPayload, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
		Type:       protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID:   "guest-result",
		MatchID:    "match-3",
		Identity:   guestIdentity,
		Context:    protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
		Score:      5,
		ShipDeaths: 1,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("encode guest payload: %v", err)
	}
	if _, err := runtime.Handle(guestPayload); err != nil {
		t.Fatalf("guest Handle returned error: %v", err)
	}

	accountLoad, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: accountIdentity,
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer},
	})
	if err != nil {
		t.Fatalf("encode account load payload: %v", err)
	}
	accountResponse, err := runtime.Handle(accountLoad)
	if err != nil {
		t.Fatalf("account load returned error: %v", err)
	}
	var accountPacket protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(accountResponse, &accountPacket); err != nil {
		t.Fatalf("unmarshal account response: %v", err)
	}
	if !accountPacket.Found {
		t.Fatal("account Found = false, want true")
	}
	if accountPacket.Stats.TotalScore != 11 || accountPacket.Stats.HighScore != 11 || accountPacket.Stats.ShipDeaths != 1 || accountPacket.Stats.GamesPlayed != 1 || accountPacket.Stats.Wins != 1 {
		t.Fatalf("account stats = %+v, want rails stats", accountPacket.Stats)
	}
	if gotLoadAuth != "Bearer internal-token" {
		t.Fatalf("load Authorization = %q, want %q", gotLoadAuth, "Bearer internal-token")
	}
	if gotLoadBody.AccountID != "acct-123" {
		t.Fatalf("load AccountID = %q, want %q", gotLoadBody.AccountID, "acct-123")
	}

	localLoad, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: localIdentity,
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
	})
	if err != nil {
		t.Fatalf("encode local load payload: %v", err)
	}
	localResponse, err := runtime.Handle(localLoad)
	if err != nil {
		t.Fatalf("local load returned error: %v", err)
	}
	var localPacket protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(localResponse, &localPacket); err != nil {
		t.Fatalf("unmarshal local response: %v", err)
	}
	if !localPacket.Found {
		t.Fatal("local Found = false, want true")
	}
	if localPacket.Stats.TotalScore != 7 || localPacket.Stats.HighScore != 7 || localPacket.Stats.ShipDeaths != 2 || localPacket.Stats.GamesPlayed != 1 {
		t.Fatalf("local stats = %+v, want sqlite stats", localPacket.Stats)
	}
	if localPacket.Stats.Wins != 0 {
		t.Fatalf("local wins = %d, want 0", localPacket.Stats.Wins)
	}

	guestLoad, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type:     protocol.PacketTypePlayerDataLoadStats,
		Identity: guestIdentity,
		Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeSinglePlayer},
	})
	if err != nil {
		t.Fatalf("encode guest load payload: %v", err)
	}
	guestResponse, err := runtime.Handle(guestLoad)
	if err != nil {
		t.Fatalf("guest load returned error: %v", err)
	}
	var guestPacket protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(guestResponse, &guestPacket); err != nil {
		t.Fatalf("unmarshal guest response: %v", err)
	}
	if !guestPacket.Found {
		t.Fatal("guest Found = false, want true")
	}
	if guestPacket.Stats.TotalScore != 5 || guestPacket.Stats.HighScore != 5 || guestPacket.Stats.ShipDeaths != 1 || guestPacket.Stats.GamesPlayed != 1 || guestPacket.Stats.Wins != 1 {
		t.Fatalf("guest stats = %+v, want guest memory stats", guestPacket.Stats)
	}

	localResponseAfterGuest, err := runtime.Handle(localLoad)
	if err != nil {
		t.Fatalf("local load after guest returned error: %v", err)
	}
	var localPacketAfterGuest protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(localResponseAfterGuest, &localPacketAfterGuest); err != nil {
		t.Fatalf("unmarshal local response after guest: %v", err)
	}
	if localPacketAfterGuest.Stats.TotalScore != 7 || localPacketAfterGuest.Stats.HighScore != 7 || localPacketAfterGuest.Stats.ShipDeaths != 2 || localPacketAfterGuest.Stats.GamesPlayed != 1 || localPacketAfterGuest.Stats.Wins != 0 {
		t.Fatalf("local stats changed after guest result: %+v", localPacketAfterGuest.Stats)
	}
}
