package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

const (
	playerDataProfilePlayModeSinglePlayer          = "single_player"
	playerDataProfilePlayModeMultiplayer           = "multiplayer"
	playerDataProfilePlayModeMultiplayerSimulation = "multiplayer_simulation"
	playerDataProfileActivityStatusActive          = "ACTIVE"
	playerDataProfileActivityStatusLocal           = "LOCAL"
	playerDataProfileActivityStatusOffline         = "OFFLINE"
)

type AuthVerifier interface {
	VerifyToken(ctx context.Context, rawToken string) (AuthVerificationResult, error)
}

type AuthVerificationResult struct {
	Valid    bool
	Identity AuthIdentity
}

type AuthIdentity struct {
	AccountID   string
	DisplayName string
}

type ProfileHandler struct {
	runtime      *playerdata.Runtime
	authVerifier AuthVerifier
}

type playerDataProfileRequest struct {
	PlayMode       string  `json:"play_mode"`
	IdentityKind   string  `json:"identity_kind"`
	LocalProfileID *string `json:"local_profile_id"`
}

type playerDataProfileResponse struct {
	Profile playerDataProfile `json:"profile"`
}

type playerDataProfile struct {
	Callsign       string                   `json:"callsign"`
	ActivityStatus string                   `json:"activity_status"`
	IdentityKind   string                   `json:"identity_kind"`
	Stats          protocol.PlayerDataStats `json:"stats"`
}

type playerDataErrorResponse struct {
	Error string `json:"error"`
}

func NewProfileHandler(runtime *playerdata.Runtime, authVerifier AuthVerifier) http.Handler {
	return &ProfileHandler{
		runtime:      runtime,
		authVerifier: authVerifier,
	}
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writePlayerDataProfileError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if h.runtime == nil {
		writePlayerDataProfileError(w, http.StatusInternalServerError, "profile_unavailable")
		return
	}

	var request playerDataProfileRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		writePlayerDataProfileError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if !isSupportedPlayerDataProfilePlayMode(request.PlayMode) {
		writePlayerDataProfileError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if !isSupportedPlayerDataProfileIdentityKind(request.IdentityKind) {
		writePlayerDataProfileError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	identity, callsign, activityStatus, statusCode, err := h.resolveIdentityAndPresentation(r.Context(), r.Header.Get("Authorization"), request)
	if err != nil {
		writePlayerDataProfileError(w, statusCode, err.Error())
		return
	}

	command := protocol.PlayerDataLoadStats{
		Type: protocol.PacketTypePlayerDataLoadStats,
		Context: protocol.PlayerDataRequestContext{
			PlayMode: request.PlayMode,
		},
		Identity: identity,
	}

	payload, err := codec.Encode(command)
	if err != nil {
		writePlayerDataProfileError(w, http.StatusInternalServerError, "profile_unavailable")
		return
	}

	responsePayload, err := h.runtime.Handle(payload)
	if err != nil {
		writePlayerDataProfileError(w, http.StatusInternalServerError, "profile_unavailable")
		return
	}

	var result protocol.PlayerDataLoadStatsResult
	if err := json.Unmarshal(responsePayload, &result); err != nil {
		writePlayerDataProfileError(w, http.StatusInternalServerError, "profile_unavailable")
		return
	}

	if result.ErrorCode != "" {
		if result.ErrorCode == "invalid_mode_identity" {
			writePlayerDataProfileError(w, http.StatusUnprocessableEntity, result.Message)
			return
		}
		writePlayerDataProfileError(w, http.StatusInternalServerError, result.Message)
		return
	}

	writePlayerDataProfileJSON(w, http.StatusOK, playerDataProfileResponse{
		Profile: playerDataProfile{
			Callsign:       callsign,
			ActivityStatus: activityStatus,
			IdentityKind:   request.IdentityKind,
			Stats:          result.Stats,
		},
	})
}

func (h *ProfileHandler) resolveIdentityAndPresentation(ctx context.Context, authorizationHeader string, request playerDataProfileRequest) (protocol.PlayerDataIdentity, string, string, int, error) {
	switch request.IdentityKind {
	case playerdata.IdentityKindGuest:
		return protocol.PlayerDataIdentity{
			IdentityKind: playerdata.IdentityKindGuest,
		}, "Guest", playerDataProfileActivityStatusOffline, 0, nil
	case playerdata.IdentityKindLocalProfile:
		if request.LocalProfileID == nil || strings.TrimSpace(*request.LocalProfileID) == "" {
			return protocol.PlayerDataIdentity{}, "", "", http.StatusBadRequest, errors.New("invalid_request")
		}
		return protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: strings.TrimSpace(*request.LocalProfileID),
		}, "Local Pilot", playerDataProfileActivityStatusLocal, 0, nil
	case playerdata.IdentityKindAuthenticatedAccount:
		token, ok := bearerTokenFromAuthorizationHeader(authorizationHeader)
		if !ok || h.authVerifier == nil {
			return protocol.PlayerDataIdentity{}, "", "", http.StatusUnauthorized, errors.New("unauthorized")
		}
		result, err := h.authVerifier.VerifyToken(ctx, token)
		if err != nil || !result.Valid || strings.TrimSpace(result.Identity.AccountID) == "" {
			return protocol.PlayerDataIdentity{}, "", "", http.StatusUnauthorized, errors.New("unauthorized")
		}
		callsign := strings.TrimSpace(result.Identity.DisplayName)
		if callsign == "" {
			callsign = "Pilot"
		}
		return protocol.PlayerDataIdentity{
			IdentityKind: playerdata.IdentityKindAuthenticatedAccount,
			AccountID:    result.Identity.AccountID,
		}, callsign, playerDataProfileActivityStatusActive, 0, nil
	default:
		return protocol.PlayerDataIdentity{}, "", "", http.StatusBadRequest, errors.New("invalid_request")
	}
}

func bearerTokenFromAuthorizationHeader(header string) (string, bool) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		return "", false
	}
	return token, true
}

func isSupportedPlayerDataProfilePlayMode(playMode string) bool {
	switch playMode {
	case playerDataProfilePlayModeSinglePlayer, playerDataProfilePlayModeMultiplayer, playerDataProfilePlayModeMultiplayerSimulation:
		return true
	default:
		return false
	}
}

func isSupportedPlayerDataProfileIdentityKind(identityKind string) bool {
	switch identityKind {
	case playerdata.IdentityKindGuest, playerdata.IdentityKindLocalProfile, playerdata.IdentityKindAuthenticatedAccount:
		return true
	default:
		return false
	}
}

func writePlayerDataProfileError(w http.ResponseWriter, statusCode int, message string) {
	writePlayerDataProfileJSON(w, statusCode, playerDataErrorResponse{Error: message})
}

func writePlayerDataProfileJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
