package networking

import (
	"fmt"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

const (
	loadoutErrorUnavailable = "loadout_unavailable"
	loadoutErrorInvalid     = "loadout_invalid"
	loadoutErrorLocked      = "loadout_locked"
)

type sessionBuildState struct {
	identity  protocol.PlayerDataIdentity
	playMode  string
	context   playerbuild.LoadedBuildContext
	selection playerbuild.LoadoutSelection
	resolved  playerbuild.ResolvedPlayerBuild
	valid     bool
	errorCode string
	message   string
}

func (session *webSocketSession) handleLoadoutOptionsRequest(localProfileID, playMode, modeID, traceID string) {
	if err := validateRequestedPlayMode(playMode); err != nil {
		session.setBuildError(loadoutErrorUnavailable, err.Error())
		session.enqueueLoadoutOptions(traceID)
		return
	}
	_ = session.loadPlayerBuildOptions(localProfileID, playMode, modeID, traceID)
	session.enqueueLoadoutOptions(traceID)
}

func (session *webSocketSession) handleSetLoadoutRequest(
	traceID string,
	selectedOwnedShipID string,
	selectedWeaponsByPoint map[string]string,
	selectedModulesBySlot map[string]string,
	startingAmmoByPoint map[string]int,
) {
	if context := session.sessionContext(); context.Room != nil && context.Room.CurrentState() != rooms.RoomStateLobby {
		session.setBuildError(loadoutErrorLocked, "Loadouts cannot change after the match starts.")
		session.enqueueLoadoutOptions(traceID)
		return
	}

	session.mu.RLock()
	service := session.buildService
	state := session.buildState
	session.mu.RUnlock()

	if service == nil || state.context.Options.PlayerID == "" {
		session.setBuildError(loadoutErrorUnavailable, "Loadout options have not been loaded.")
		session.enqueueLoadoutOptions(traceID)
		return
	}

	selection := playerbuild.LoadoutSelection{
		PlayerID:               state.context.Options.PlayerID,
		ModeID:                 state.context.Options.ModeID,
		SelectedOwnedShipID:    selectedOwnedShipID,
		SelectedWeaponsByPoint: weaponPointMap(selectedWeaponsByPoint),
		SelectedModulesBySlot:  moduleSlotMap(selectedModulesBySlot),
		StartingAmmoByPoint:    ammoPointMap(startingAmmoByPoint),
	}
	resolved, err := service.ResolveSelection(state.context, selection)
	if err != nil {
		session.setBuildError(loadoutErrorInvalid, err.Error())
		session.enqueueLoadoutOptions(traceID)
		return
	}

	session.mu.Lock()
	state = session.buildState
	state.selection = selection
	state.resolved = resolved
	state.valid = true
	state.errorCode = ""
	state.message = ""
	session.buildState = state
	session.mu.Unlock()

	session.enqueueLoadoutOptions(traceID)
	if context := session.sessionContext(); context.Room != nil {
		session.EnqueueRoomSnapshot(context.Room)
	}
}

func (session *webSocketSession) preparePlayerBuildForRoom(room *rooms.Room, localProfileID, playMode, traceID string) {
	if room == nil || session.buildService == nil {
		return
	}
	projection := room.SnapshotForSession(session.sessionID)
	modeID := projection.ResolvedModeID
	if modeID == "" {
		modeID = string(projection.ModeConfig.PresetID)
	}
	if modeID == "" {
		modeID = string(modes.ModeArcadeSurvival)
	}
	identity := session.playerDataIdentity(playMode, localProfileID)

	session.mu.RLock()
	state := session.buildState
	canReuse := state.valid && state.playMode == playMode && state.identity == identity && state.context.Options.ModeID == modeID
	session.mu.RUnlock()
	if canReuse {
		return
	}
	_ = session.loadPlayerBuildOptions(localProfileID, playMode, modeID, traceID)
}

func (session *webSocketSession) loadPlayerBuildOptions(localProfileID, playMode, modeID, traceID string) error {
	session.mu.RLock()
	service := session.buildService
	session.mu.RUnlock()
	if service == nil {
		return nil
	}
	if modeID == "" {
		modeID = string(modes.ModeArcadeSurvival)
	}
	identity := session.playerDataIdentity(playMode, localProfileID)
	loaded, err := service.LoadOptions(
		session.sessionID,
		identity,
		protocol.PlayerDataRequestContext{PlayMode: playMode, TraceID: traceID},
		playerbuild.Rules{ModeID: modeID},
	)
	if err != nil {
		session.mu.Lock()
		session.buildState = sessionBuildState{
			identity: identity, playMode: playMode,
			errorCode: loadoutErrorUnavailable, message: err.Error(),
		}
		session.mu.Unlock()
		return err
	}

	selection := loaded.Options.FallbackLoadout
	resolved, resolveErr := service.ResolveSelection(loaded, selection)
	state := sessionBuildState{
		identity: identity, playMode: playMode, context: loaded, selection: selection,
	}
	if resolveErr != nil {
		state.errorCode = loadoutErrorInvalid
		state.message = resolveErr.Error()
	} else {
		state.resolved = resolved
		state.valid = true
	}
	session.mu.Lock()
	session.buildState = state
	session.mu.Unlock()
	return resolveErr
}

func (session *webSocketSession) playerDataIdentity(playMode, localProfileID string) protocol.PlayerDataIdentity {
	if playMode == playerdata.PlayModeSinglePlayer {
		if localProfileID != "" {
			return protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile, LocalProfileID: localProfileID}
		}
		return protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest}
	}
	identity := session.SessionIdentity()
	if identity.IsAuthenticatedAccount() {
		return protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindAuthenticatedAccount, AccountID: identity.AccountID}
	}
	return protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest}
}

func (session *webSocketSession) setBuildError(code, message string) {
	session.mu.Lock()
	state := session.buildState
	state.errorCode = code
	state.message = message
	session.buildState = state
	session.mu.Unlock()
}

func (session *webSocketSession) resolvedBuildTemplate() playerbuild.ResolvedPlayerBuild {
	session.mu.RLock()
	defer session.mu.RUnlock()
	if session.buildState.valid {
		return session.buildState.resolved.Clone()
	}
	return playerbuild.DefaultResolvedBuild("pending-player")
}

func validateRequestedPlayMode(playMode string) error {
	switch playMode {
	case playerdata.PlayModeSinglePlayer, playerdata.PlayModeMultiplayer:
		return nil
	default:
		return fmt.Errorf("unsupported loadout play mode %q", playMode)
	}
}
