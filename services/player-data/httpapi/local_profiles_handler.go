package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
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
	r = withRequestContext(w, r)
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/api/player-data/local-profiles/default" {
			h.serveGetDefault(w, r)
			return
		}
		h.serveList(w, r)
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

func (h *LocalProfilesHandler) serveList(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "list_local_profiles", playerdata.ErrLocalProfileUnavailable, "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	profiles, err := h.runtime.ListLocalProfiles()
	if err != nil {
		if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
			emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "list_local_profiles", err, "")
			writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
			return
		}
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "list_local_profiles", err, "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
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

	stats, err := h.runtime.LocalProfileSeedStats(request.SeedFromGuestStats)
	if err != nil {
		emitLocalProfileCreateFailure(r, "guest_seed_read", "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	localProfileID, err := generateLocalProfileID()
	if err != nil {
		emitLocalProfileCreateFailure(r, "id_generation", "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	profile, err := h.runtime.CreateLocalProfile(localProfileID, displayName, stats)
	if err != nil {
		emitLocalProfileCreateFailure(r, "store_write", localProfileID)
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	writePlayerDataLocalProfilesJSON(w, http.StatusCreated, playerDataLocalProfileResponse{
		Profile: playerDataLocalProfile{
			LocalProfileID: profile.LocalProfileID,
			DisplayName:    profile.DisplayName,
		},
	})
}

func (h *LocalProfilesHandler) serveGetDefault(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.runtime == nil {
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "get_default_local_profile", playerdata.ErrLocalProfileUnavailable, "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	defaultProfile, err := h.runtime.GetDefaultLocalProfile()
	if err != nil {
		if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
			emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "get_default_local_profile", err, "")
			writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
			return
		}
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataReadFailed, "get_default_local_profile", err, "")
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
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
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "set_default_local_profile", playerdata.ErrLocalProfileUnavailable, "")
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
		if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
			emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "set_default_local_profile", err, localProfileID)
			writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
			return
		}
		if errors.Is(err, playerdata.ErrLocalProfileNotFound) {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "set_default_local_profile", err, localProfileID)
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
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
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "update_local_profile", playerdata.ErrLocalProfileUnavailable, "")
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
		if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
			emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "update_local_profile", err, localProfileID)
			writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
			return
		}
		if errors.Is(err, playerdata.ErrLocalProfileNotFound) {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "update_local_profile", err, localProfileID)
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
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
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "delete_local_profile", playerdata.ErrLocalProfileUnavailable, "")
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
		if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
			emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "delete_local_profile", err, localProfileID)
			writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
			return
		}
		if errors.Is(err, playerdata.ErrLocalProfileNotFound) {
			writePlayerDataLocalProfilesError(w, http.StatusNotFound, "local_profile_not_found")
			return
		}
		emitLocalProfileOperationFailure(r, observability.EventNamePlayerDataWriteFailed, "delete_local_profile", err, localProfileID)
		writePlayerDataLocalProfilesError(w, http.StatusServiceUnavailable, "local_profiles_unavailable")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func emitLocalProfileCreateFailure(r *http.Request, failureMode, localProfileID string) {
	fields := observability.Fields{
		"failure_mode": failureMode,
		"error_code":   "operation_failed",
	}
	if localProfileID != "" {
		fields["local_profile_id"] = localProfileID
	}
	logging.Emit(observability.Request{
		Event:   observability.EventNameLocalProfileCreateFailed,
		Context: observability.Context{TraceID: TraceIDFromContext(r.Context()), RequestID: RequestIDFromContext(r.Context())},
		Fields:  fields,
	})
}

func emitLocalProfileOperationFailure(r *http.Request, event observability.EventName, operation string, err error, localProfileID string) {
	fields := observability.Fields{"operation": operation, "error_code": "operation_failed"}
	if class := playerdata.FailureClassOf(err); class != "" {
		fields["error_code"] = string(class)
	}
	if errors.Is(err, playerdata.ErrLocalProfileUnavailable) {
		fields["failure_mode"] = "store_unavailable"
	}
	if localProfileID != "" {
		fields["local_profile_id"] = localProfileID
	}
	logging.Emit(observability.Request{
		Event:   event,
		Context: observability.Context{TraceID: TraceIDFromContext(r.Context()), RequestID: RequestIDFromContext(r.Context())},
		Fields:  fields,
	})
}
