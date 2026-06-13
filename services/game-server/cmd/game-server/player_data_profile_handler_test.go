package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/server/internal/authclient"
)

type fakePlayerDataRuntime struct {
	payloads [][]byte
	response []byte
	err      error
}

func (r *fakePlayerDataRuntime) Handle(payload []byte) ([]byte, error) {
	r.payloads = append(r.payloads, append([]byte(nil), payload...))
	if r.err != nil {
		return nil, r.err
	}
	return append([]byte(nil), r.response...), nil
}

type fakeTokenVerifier struct {
	result authclient.VerifyResult
	err    error
	calls  []string
}

func (v *fakeTokenVerifier) VerifyToken(_ context.Context, rawToken string) (authclient.VerifyResult, error) {
	v.calls = append(v.calls, rawToken)
	if v.err != nil {
		return authclient.VerifyResult{}, v.err
	}
	return v.result, nil
}

func TestPlayerDataProfileHandlerGuestSuccess(t *testing.T) {
	responsePayload, err := codec.Encode(protocol.PlayerDataLoadStatsResult{
		Type:  protocol.PacketTypePlayerDataLoadStatsResult,
		Found: true,
		Stats: protocol.PlayerDataStats{
			TotalScore:  12,
			HighScore:   12,
			ShipDeaths:  3,
			GamesPlayed: 4,
			Wins:        0,
		},
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	runtime := &fakePlayerDataRuntime{response: responsePayload}
	handler := newPlayerDataProfileHandler(runtime, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"single_player","identity_kind":"guest"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(runtime.payloads) != 1 {
		t.Fatalf("runtime payload count = %d, want 1", len(runtime.payloads))
	}

	var command protocol.PlayerDataLoadStats
	if err := json.Unmarshal(runtime.payloads[0], &command); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if command.Type != protocol.PacketTypePlayerDataLoadStats {
		t.Fatalf("command.Type = %q, want %q", command.Type, protocol.PacketTypePlayerDataLoadStats)
	}
	if command.Context.PlayMode != playerDataProfilePlayModeSinglePlayer {
		t.Fatalf("command.Context.PlayMode = %q, want %q", command.Context.PlayMode, playerDataProfilePlayModeSinglePlayer)
	}
	if command.Identity.IdentityKind != playerdata.IdentityKindGuest {
		t.Fatalf("command.Identity.IdentityKind = %q, want %q", command.Identity.IdentityKind, playerdata.IdentityKindGuest)
	}
	if command.Identity.AccountID != "" {
		t.Fatalf("command.Identity.AccountID = %q, want empty", command.Identity.AccountID)
	}
	if command.Identity.LocalProfileID != "" {
		t.Fatalf("command.Identity.LocalProfileID = %q, want empty", command.Identity.LocalProfileID)
	}

	var body playerDataProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.Callsign != "Guest" {
		t.Fatalf("callsign = %q, want %q", body.Profile.Callsign, "Guest")
	}
	if body.Profile.ActivityStatus != playerDataProfileActivityStatusOffline {
		t.Fatalf("activity_status = %q, want %q", body.Profile.ActivityStatus, playerDataProfileActivityStatusOffline)
	}
	if body.Profile.IdentityKind != playerdata.IdentityKindGuest {
		t.Fatalf("identity_kind = %q, want %q", body.Profile.IdentityKind, playerdata.IdentityKindGuest)
	}
	if body.Profile.Stats.TotalScore != 12 || body.Profile.Stats.HighScore != 12 || body.Profile.Stats.ShipDeaths != 3 || body.Profile.Stats.GamesPlayed != 4 {
		t.Fatalf("stats = %+v, want runtime stats", body.Profile.Stats)
	}
}

func TestPlayerDataProfileHandlerLocalProfileSuccess(t *testing.T) {
	responsePayload, err := codec.Encode(protocol.PlayerDataLoadStatsResult{
		Type:  protocol.PacketTypePlayerDataLoadStatsResult,
		Found: true,
		Stats: protocol.PlayerDataStats{
			TotalScore:  88,
			HighScore:   77,
			ShipDeaths:  5,
			GamesPlayed: 9,
			Wins:        3,
		},
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	localProfileID := "local-profile-123"
	runtime := &fakePlayerDataRuntime{response: responsePayload}
	handler := newPlayerDataProfileHandler(runtime, nil)

	requestBody := `{"play_mode":"single_player","identity_kind":"local_profile","local_profile_id":"` + localProfileID + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(requestBody))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(runtime.payloads) != 1 {
		t.Fatalf("runtime payload count = %d, want 1", len(runtime.payloads))
	}

	var command protocol.PlayerDataLoadStats
	if err := json.Unmarshal(runtime.payloads[0], &command); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if command.Identity.IdentityKind != playerdata.IdentityKindLocalProfile {
		t.Fatalf("command.Identity.IdentityKind = %q, want %q", command.Identity.IdentityKind, playerdata.IdentityKindLocalProfile)
	}
	if command.Identity.LocalProfileID != localProfileID {
		t.Fatalf("command.Identity.LocalProfileID = %q, want %q", command.Identity.LocalProfileID, localProfileID)
	}
	if command.Identity.AccountID != "" {
		t.Fatalf("command.Identity.AccountID = %q, want empty", command.Identity.AccountID)
	}
	if command.Context.PlayMode != playerDataProfilePlayModeSinglePlayer {
		t.Fatalf("command.Context.PlayMode = %q, want %q", command.Context.PlayMode, playerDataProfilePlayModeSinglePlayer)
	}

	var body playerDataProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.Callsign != "Local Pilot" {
		t.Fatalf("callsign = %q, want %q", body.Profile.Callsign, "Local Pilot")
	}
	if body.Profile.ActivityStatus != playerDataProfileActivityStatusLocal {
		t.Fatalf("activity_status = %q, want %q", body.Profile.ActivityStatus, playerDataProfileActivityStatusLocal)
	}
	if body.Profile.IdentityKind != playerdata.IdentityKindLocalProfile {
		t.Fatalf("identity_kind = %q, want %q", body.Profile.IdentityKind, playerdata.IdentityKindLocalProfile)
	}
}

func TestPlayerDataProfileHandlerLocalProfileMissingLocalProfileID(t *testing.T) {
	runtime := &fakePlayerDataRuntime{}
	handler := newPlayerDataProfileHandler(runtime, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"single_player","identity_kind":"local_profile"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if len(runtime.payloads) != 0 {
		t.Fatalf("runtime payload count = %d, want 0", len(runtime.payloads))
	}

	var body playerDataErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "invalid_request" {
		t.Fatalf("error = %q, want %q", body.Error, "invalid_request")
	}
}

func TestPlayerDataProfileHandlerAuthenticatedAccountSuccess(t *testing.T) {
	responsePayload, err := codec.Encode(protocol.PlayerDataLoadStatsResult{
		Type:  protocol.PacketTypePlayerDataLoadStatsResult,
		Found: true,
		Stats: protocol.PlayerDataStats{
			TotalScore:  44,
			HighScore:   44,
			ShipDeaths:  2,
			GamesPlayed: 6,
			Wins:        1,
		},
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	runtime := &fakePlayerDataRuntime{response: responsePayload}
	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{
			Valid: true,
			Identity: authclient.Identity{
				AccountID:   "11111111-2222-3333-4444-555555555555",
				DisplayName: "Ada",
			},
		},
	}
	handler := newPlayerDataProfileHandler(runtime, verifier)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"multiplayer","identity_kind":"authenticated_account","account_id":"ignored-by-handler"}`))
	request.Header.Set("Authorization", "Bearer submitted-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(verifier.calls) != 1 {
		t.Fatalf("verifier call count = %d, want 1", len(verifier.calls))
	}
	if verifier.calls[0] != "submitted-token" {
		t.Fatalf("verifier token = %q, want %q", verifier.calls[0], "submitted-token")
	}
	if len(runtime.payloads) != 1 {
		t.Fatalf("runtime payload count = %d, want 1", len(runtime.payloads))
	}

	var command protocol.PlayerDataLoadStats
	if err := json.Unmarshal(runtime.payloads[0], &command); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if command.Context.PlayMode != playerDataProfilePlayModeMultiplayer {
		t.Fatalf("command.Context.PlayMode = %q, want %q", command.Context.PlayMode, playerDataProfilePlayModeMultiplayer)
	}
	if command.Identity.IdentityKind != playerdata.IdentityKindAuthenticatedAccount {
		t.Fatalf("command.Identity.IdentityKind = %q, want %q", command.Identity.IdentityKind, playerdata.IdentityKindAuthenticatedAccount)
	}
	if command.Identity.AccountID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("command.Identity.AccountID = %q, want verified account id", command.Identity.AccountID)
	}
	if command.Identity.LocalProfileID != "" {
		t.Fatalf("command.Identity.LocalProfileID = %q, want empty", command.Identity.LocalProfileID)
	}

	var body playerDataProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.Callsign != "Ada" {
		t.Fatalf("callsign = %q, want %q", body.Profile.Callsign, "Ada")
	}
	if body.Profile.ActivityStatus != playerDataProfileActivityStatusActive {
		t.Fatalf("activity_status = %q, want %q", body.Profile.ActivityStatus, playerDataProfileActivityStatusActive)
	}
	if body.Profile.IdentityKind != playerdata.IdentityKindAuthenticatedAccount {
		t.Fatalf("identity_kind = %q, want %q", body.Profile.IdentityKind, playerdata.IdentityKindAuthenticatedAccount)
	}
}

func TestPlayerDataProfileHandlerAuthenticatedAccountDerivesIdentityAndUsesLoadStats(t *testing.T) {
	responsePayload, err := codec.Encode(protocol.PlayerDataLoadStatsResult{
		Type:  protocol.PacketTypePlayerDataLoadStatsResult,
		Found: true,
		Stats: protocol.PlayerDataStats{
			TotalScore:  9,
			HighScore:   9,
			ShipDeaths:  1,
			GamesPlayed: 1,
			Wins:        1,
		},
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	runtime := &fakePlayerDataRuntime{response: responsePayload}
	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{
			Valid: true,
			Identity: authclient.Identity{
				AccountID:   "11111111-2222-3333-4444-555555555555",
				DisplayName: "Ada",
			},
		},
	}
	handler := newPlayerDataProfileHandler(runtime, verifier)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"multiplayer","identity_kind":"authenticated_account","account_id":"client-supplied-account-id"}`))
	request.Header.Set("Authorization", "Bearer submitted-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(runtime.payloads) != 1 {
		t.Fatalf("runtime payload count = %d, want 1", len(runtime.payloads))
	}

	var command protocol.PlayerDataLoadStats
	if err := json.Unmarshal(runtime.payloads[0], &command); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	if command.Type != protocol.PacketTypePlayerDataLoadStats {
		t.Fatalf("command.Type = %q, want %q", command.Type, protocol.PacketTypePlayerDataLoadStats)
	}
	if command.Context.PlayMode != playerDataProfilePlayModeMultiplayer {
		t.Fatalf("command.Context.PlayMode = %q, want %q", command.Context.PlayMode, playerDataProfilePlayModeMultiplayer)
	}
	if command.Identity.IdentityKind != playerdata.IdentityKindAuthenticatedAccount {
		t.Fatalf("command.Identity.IdentityKind = %q, want %q", command.Identity.IdentityKind, playerdata.IdentityKindAuthenticatedAccount)
	}
	if command.Identity.AccountID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("command.Identity.AccountID = %q, want verified account id", command.Identity.AccountID)
	}
	if command.Identity.LocalProfileID != "" {
		t.Fatalf("command.Identity.LocalProfileID = %q, want empty", command.Identity.LocalProfileID)
	}

	var body playerDataProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.Callsign != "Ada" {
		t.Fatalf("callsign = %q, want %q", body.Profile.Callsign, "Ada")
	}
	if body.Profile.IdentityKind != playerdata.IdentityKindAuthenticatedAccount {
		t.Fatalf("identity_kind = %q, want %q", body.Profile.IdentityKind, playerdata.IdentityKindAuthenticatedAccount)
	}
	if body.Profile.Stats.TotalScore != 9 || body.Profile.Stats.HighScore != 9 || body.Profile.Stats.ShipDeaths != 1 || body.Profile.Stats.GamesPlayed != 1 || body.Profile.Stats.Wins != 1 {
		t.Fatalf("stats = %+v, want runtime stats", body.Profile.Stats)
	}
}

func TestPlayerDataProfileHandlerAuthenticatedAccountMissingAuthorization(t *testing.T) {
	runtime := &fakePlayerDataRuntime{}
	handler := newPlayerDataProfileHandler(runtime, &fakeTokenVerifier{})

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"multiplayer","identity_kind":"authenticated_account"}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if len(runtime.payloads) != 0 {
		t.Fatalf("runtime payload count = %d, want 0", len(runtime.payloads))
	}
}

func TestPlayerDataProfileHandlerAuthenticatedAccountInvalidToken(t *testing.T) {
	runtime := &fakePlayerDataRuntime{}
	verifier := &fakeTokenVerifier{result: authclient.VerifyResult{Valid: false}}
	handler := newPlayerDataProfileHandler(runtime, verifier)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":"multiplayer","identity_kind":"authenticated_account","account_id":"ignored-by-handler"}`))
	request.Header.Set("Authorization", "Bearer invalid-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if len(verifier.calls) != 1 {
		t.Fatalf("verifier call count = %d, want 1", len(verifier.calls))
	}
	if verifier.calls[0] != "invalid-token" {
		t.Fatalf("verifier token = %q, want %q", verifier.calls[0], "invalid-token")
	}
	if len(runtime.payloads) != 0 {
		t.Fatalf("runtime payload count = %d, want 0", len(runtime.payloads))
	}
}

func TestPlayerDataProfileHandlerMalformedRequest(t *testing.T) {
	runtime := &fakePlayerDataRuntime{}
	handler := newPlayerDataProfileHandler(runtime, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/player-data/profile", strings.NewReader(`{"play_mode":`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if len(runtime.payloads) != 0 {
		t.Fatalf("runtime payload count = %d, want 0", len(runtime.payloads))
	}
}
