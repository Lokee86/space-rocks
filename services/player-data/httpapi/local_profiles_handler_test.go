package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestLocalProfilesHandlerReturnsUnavailableWhenLocalProfileStoreMissing(t *testing.T) {
	runtime, err := playerdata.NewConfiguredRuntime(playerdata.RuntimeConfig{})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}
	if err := runtime.DeleteLocalProfile("missing-profile"); !errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
		t.Fatalf("DeleteLocalProfile error = %v, want ErrLocalProfileUnavailable", err)
	}

	handler := NewLocalProfilesHandler(runtime)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{
			name:       "list",
			method:     http.MethodGet,
			target:     "/api/player-data/local-profiles",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "default get",
			method:     http.MethodGet,
			target:     "/api/player-data/local-profiles/default",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "create",
			method:     http.MethodPost,
			target:     "/api/player-data/local-profiles",
			body:       `{"display_name":"Pilot_1"}`,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "default put",
			method:     http.MethodPut,
			target:     "/api/player-data/local-profiles/default",
			body:       `{"identity_kind":"guest","local_profile_id":""}`,
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			var payload map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if payload["error"] != "local_profiles_unavailable" {
				t.Fatalf("error = %q, want %q", payload["error"], "local_profiles_unavailable")
			}
		})
	}
}

type localProfilesHandlerTestStore struct {
	guestStats protocol.PlayerDataStats
	profiles   map[string]protocol.PlayerDataStats
}

func newLocalProfilesHandlerTestStore() *localProfilesHandlerTestStore {
	return &localProfilesHandlerTestStore{
		profiles: make(map[string]protocol.PlayerDataStats),
	}
}

func (s *localProfilesHandlerTestStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	switch identity.IdentityKind {
	case playerdata.IdentityKindGuest:
		return s.guestStats, true, nil
	case playerdata.IdentityKindLocalProfile:
		stats, found := s.profiles[identity.LocalProfileID]
		return stats, found, nil
	default:
		return protocol.PlayerDataStats{}, false, errors.New("unsupported identity")
	}
}

func (s *localProfilesHandlerTestStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, errors.New("unsupported")
}

func (s *localProfilesHandlerTestStore) ListLocalProfiles() ([]playerdata.LocalProfileSummary, error) {
	return nil, playerdata.ErrLocalProfileUnavailable
}

func (s *localProfilesHandlerTestStore) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (playerdata.LocalProfileSummary, error) {
	s.profiles[localProfileID] = stats
	return playerdata.LocalProfileSummary{
		LocalProfileID: localProfileID,
		DisplayName:    displayName,
	}, nil
}

func (s *localProfilesHandlerTestStore) DeleteLocalProfile(localProfileID string) error {
	return playerdata.ErrLocalProfileUnavailable
}

func (s *localProfilesHandlerTestStore) UpdateLocalProfileDisplayName(localProfileID string, displayName string) (playerdata.LocalProfileSummary, error) {
	return playerdata.LocalProfileSummary{}, playerdata.ErrLocalProfileUnavailable
}

func (s *localProfilesHandlerTestStore) GetDefaultLocalProfile() (playerdata.LocalProfileDefault, error) {
	return playerdata.LocalProfileDefault{}, playerdata.ErrLocalProfileUnavailable
}

func (s *localProfilesHandlerTestStore) SetDefaultLocalProfile(identityKind string, localProfileID string) (playerdata.LocalProfileDefault, error) {
	return playerdata.LocalProfileDefault{}, playerdata.ErrLocalProfileUnavailable
}

func TestLocalProfilesHandlerCreateUsesZeroSeedStatsWhenDisabled(t *testing.T) {
	runtime, err := playerdata.NewRuntime(playerdata.Config{Store: newLocalProfilesHandlerTestStore()})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	handler := NewLocalProfilesHandler(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/player-data/local-profiles", strings.NewReader(`{"display_name":"Pilot_1","seed_from_guest_stats":false}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var response playerDataLocalProfileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Profile.DisplayName != "Pilot_1" {
		t.Fatalf("display_name = %q, want %q", response.Profile.DisplayName, "Pilot_1")
	}

	stats, found, err := runtime.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind:   playerdata.IdentityKindLocalProfile,
		LocalProfileID: response.Profile.LocalProfileID,
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if !found {
		t.Fatal("LoadStats found = false, want true")
	}
	if stats != (protocol.PlayerDataStats{}) {
		t.Fatalf("stats = %+v, want zero stats", stats)
	}
}

func TestLocalProfilesHandlerCreateSeedsGuestStatsWhenEnabled(t *testing.T) {
	store := newLocalProfilesHandlerTestStore()
	store.guestStats = protocol.PlayerDataStats{
		GamesPlayed: 3,
		TotalScore:  17,
		HighScore:   9,
		ShipDeaths:  4,
		Wins:        2,
	}

	runtime, err := playerdata.NewRuntime(playerdata.Config{Store: store})
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}

	handler := NewLocalProfilesHandler(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/player-data/local-profiles", strings.NewReader(`{"display_name":"Pilot_2","seed_from_guest_stats":true}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var response playerDataLocalProfileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Profile.DisplayName != "Pilot_2" {
		t.Fatalf("display_name = %q, want %q", response.Profile.DisplayName, "Pilot_2")
	}

	stats, found, err := runtime.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind:   playerdata.IdentityKindLocalProfile,
		LocalProfileID: response.Profile.LocalProfileID,
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if !found {
		t.Fatal("LoadStats found = false, want true")
	}
	want := store.guestStats
	if stats != want {
		t.Fatalf("stats = %+v, want %+v", stats, want)
	}
}
