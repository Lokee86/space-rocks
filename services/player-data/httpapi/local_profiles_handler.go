package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type LocalProfilesHandler struct {
	runtime *playerdata.Runtime
}

type playerDataLocalProfilesRequest struct {
	DisplayName        string `json:"display_name"`
	SeedFromGuestStats bool   `json:"seed_from_guest_stats"`
}

type playerDataLocalProfilesResponse struct {
	Profiles []playerDataLocalProfile `json:"profiles"`
}

type playerDataLocalProfileResponse struct {
	Profile playerDataLocalProfile `json:"profile"`
}

type playerDataLocalProfileDefaultRequest struct {
	IdentityKind   string `json:"identity_kind"`
	LocalProfileID string `json:"local_profile_id"`
}

type playerDataLocalProfileDefaultResponse struct {
	DefaultProfile playerDataLocalProfileDefault `json:"default_profile"`
}

type playerDataLocalProfile struct {
	LocalProfileID string `json:"local_profile_id"`
	DisplayName    string `json:"display_name"`
}

type playerDataLocalProfileDefault struct {
	IdentityKind   string `json:"identity_kind"`
	LocalProfileID string `json:"local_profile_id"`
	DisplayName    string `json:"display_name"`
}

var localProfileDisplayNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func NewLocalProfilesHandler(runtime *playerdata.Runtime) http.Handler {
	return &LocalProfilesHandler{runtime: runtime}
}

func (h *LocalProfilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/api/player-data/local-profiles/default" {
			h.serveGetDefault(w)
			return
		}
		h.serveList(w)
	case http.MethodPost:
		h.serveCreate(w, r)
	case http.MethodPut:
		if r.URL.Path == "/api/player-data/local-profiles/default" {
			h.serveSetDefault(w, r)
			return
		}
		h.serveUpdate(w, r)
	case http.MethodDelete:
		h.serveDelete(w, r)
	default:
		writePlayerDataLocalProfilesError(w, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

func writePlayerDataLocalProfilesError(w http.ResponseWriter, statusCode int, message string) {
	writePlayerDataLocalProfilesJSON(w, statusCode, map[string]string{"error": message})
}

func writePlayerDataLocalProfilesJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *LocalProfilesHandler) serveList(w http.ResponseWriter) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	profiles, err := h.runtime.ListLocalProfiles()
	if err != nil {
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	response := playerDataLocalProfilesResponse{
		Profiles: make([]playerDataLocalProfile, 0, len(profiles)),
	}
	for _, profile := range profiles {
		response.Profiles = append(response.Profiles, playerDataLocalProfile{
			LocalProfileID: profile.LocalProfileID,
			DisplayName:    profile.DisplayName,
		})
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusOK, response)
}

func (h *LocalProfilesHandler) serveCreate(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	var request playerDataLocalProfilesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	displayName := strings.TrimSpace(request.DisplayName)
	if !isValidLocalProfileDisplayName(displayName) {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	stats, err := h.localProfileSeedStats(request.SeedFromGuestStats)
	if err != nil {
		logging.HTTP.Error("local profile create guest stat seeding failed", err,
			logging.FieldOperation, "create_local_profile",
		)
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	localProfileID, err := generateLocalProfileID()
	if err != nil {
		logging.HTTP.Error("local profile create id generation failed", err,
			logging.FieldOperation, "create_local_profile",
		)
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	profile, err := h.runtime.CreateLocalProfile(localProfileID, displayName, stats)
	if err != nil {
		logging.HTTP.Error("local profile create failed", err,
			logging.FieldOperation, "create_local_profile",
			logging.FieldLocalProfileID, localProfileID,
		)
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusCreated, playerDataLocalProfileResponse{
		Profile: playerDataLocalProfile{
			LocalProfileID: profile.LocalProfileID,
			DisplayName:    profile.DisplayName,
		},
	})
}

func (h *LocalProfilesHandler) serveGetDefault(w http.ResponseWriter) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	defaultProfile, err := h.runtime.GetDefaultLocalProfile()
	if err != nil {
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusOK, playerDataLocalProfileDefaultResponse{
		DefaultProfile: playerDataLocalProfileDefault{
			IdentityKind:   defaultProfile.IdentityKind,
			LocalProfileID: defaultProfile.LocalProfileID,
			DisplayName:    defaultProfile.DisplayName,
		},
	})
}

func (h *LocalProfilesHandler) serveSetDefault(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	var request playerDataLocalProfileDefaultRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	identityKind := strings.TrimSpace(request.IdentityKind)
	localProfileID := strings.TrimSpace(request.LocalProfileID)
	switch identityKind {
	case playerdata.IdentityKindGuest:
		if localProfileID != "" {
			writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
			return
		}
	case playerdata.IdentityKindLocalProfile:
		if localProfileID == "" {
			writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
			return
		}
	default:
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	defaultProfile, err := h.runtime.SetDefaultLocalProfile(identityKind, localProfileID)
	if err != nil {
		if err.Error() == "local profile not found" {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusOK, playerDataLocalProfileDefaultResponse{
		DefaultProfile: playerDataLocalProfileDefault{
			IdentityKind:   defaultProfile.IdentityKind,
			LocalProfileID: defaultProfile.LocalProfileID,
			DisplayName:    defaultProfile.DisplayName,
		},
	})
}

func (h *LocalProfilesHandler) serveUpdate(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	localProfileID := strings.TrimSpace(r.PathValue("local_profile_id"))
	if localProfileID == "" {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	var request playerDataLocalProfilesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	displayName := strings.TrimSpace(request.DisplayName)
	if !isValidLocalProfileDisplayName(displayName) {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	profile, err := h.runtime.UpdateLocalProfileDisplayName(localProfileID, displayName)
	if err != nil {
		if err.Error() == "local profile not found" {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusOK, playerDataLocalProfileResponse{
		Profile: playerDataLocalProfile{
			LocalProfileID: profile.LocalProfileID,
			DisplayName:    profile.DisplayName,
		},
	})
}

func (h *LocalProfilesHandler) serveDelete(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	localProfileID := strings.TrimSpace(r.PathValue("local_profile_id"))
	if localProfileID == "" {
		writePlayerDataLocalProfilesError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	err := h.runtime.DeleteLocalProfile(localProfileID)
	if err != nil {
		if err.Error() == "local profile not found" {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		writePlayerDataLocalProfilesError(w, http.StatusInternalServerError, "local_profiles_unavailable")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LocalProfilesHandler) localProfileSeedStats(seedFromGuestStats bool) (protocol.PlayerDataStats, error) {
	if !seedFromGuestStats {
		return protocol.PlayerDataStats{}, nil
	}
	if h == nil || h.runtime == nil {
		return protocol.PlayerDataStats{}, fmt.Errorf("player-data runtime is required")
	}

	payload, err := codec.Encode(protocol.PlayerDataLoadStats{
		Type: protocol.PacketTypePlayerDataLoadStats,
		Context: protocol.PlayerDataRequestContext{
			PlayMode: playerDataProfilePlayModeSinglePlayer,
		},
		Identity: protocol.PlayerDataIdentity{
			IdentityKind: playerdata.IdentityKindGuest,
		},
	})
	if err != nil {
		return protocol.PlayerDataStats{}, err
	}

	responsePayload, err := h.runtime.Handle(payload)
	if err != nil {
		return protocol.PlayerDataStats{}, err
	}

	var result protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(responsePayload, &result); err != nil {
		return protocol.PlayerDataStats{}, err
	}
	if result.ErrorCode != "" || !result.Found {
		return protocol.PlayerDataStats{}, fmt.Errorf("guest stats unavailable")
	}

	return result.Stats, nil
}

func isValidLocalProfileDisplayName(displayName string) bool {
	if displayName == "" {
		return false
	}
	return localProfileDisplayNamePattern.MatchString(displayName)
}

func generateLocalProfileID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	return fmt.Sprintf("local-profile-%x", bytes[:]), nil
}
