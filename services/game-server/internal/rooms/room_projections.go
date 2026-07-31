package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

// GameplayContext is the coherent room runtime view used by networking operations.
type GameplayContext struct {
	State   RoomState
	Game    *game.Game
	MatchID string
}

func (room *Room) GameplayContext() GameplayContext {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.gameplayContextLocked()
}

func (room *Room) GameplayContextMatches(expected GameplayContext) bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	actual := room.gameplayContextLocked()
	return actual.State == expected.State && actual.Game == expected.Game && actual.MatchID == expected.MatchID
}

func (room *Room) gameplayContextLocked() GameplayContext {
	return GameplayContext{State: room.State, Game: room.match.Game(), MatchID: room.match.CurrentMatchID()}
}

type RoomSnapshot struct {
	RoomID                string
	State                 RoomState
	CurrentMatchID        string
	Members               []RoomMember
	LocalPlayerID         string
	OwnerID               string
	ModeConfig            modes.RoomModeConfig
	ResolvedModeID        string
	ModeLocked            bool
	TeamConfig            teams.Config
	TeamAssignments       teams.Assignments
	TeamAssignmentsLocked bool
	MaxPlayers            int
	ResolvedSummary       playerdata.MatchResultSummary
	HasResolvedMatch      bool
}

func (room *Room) SnapshotForSession(sessionID string) RoomSnapshot {
	room.mu.Lock()
	defer room.mu.Unlock()

	localPlayerID, _ := room.membership.playerIDForSession(sessionID)
	resolvedSummary, hasResolvedMatch := room.match.ResolvedSummary()
	if hasResolvedMatch {
		resolvedSummary.Players = append([]playerdata.PlayerMatchSummary(nil), resolvedSummary.Players...)
	}
	resolvedModeID := ""
	modeLocked := room.roomMode.resolved != nil
	if modeLocked {
		resolvedModeID = string(room.roomMode.resolved.ModeID)
	}
	return RoomSnapshot{
		RoomID:                room.ID,
		State:                 room.State,
		CurrentMatchID:        room.match.CurrentMatchID(),
		Members:               room.membership.membersSnapshot(),
		LocalPlayerID:         localPlayerID,
		OwnerID:               room.membership.ownerIDValue(),
		ModeConfig:            room.roomMode.config,
		ResolvedModeID:        resolvedModeID,
		ModeLocked:            modeLocked,
		TeamConfig:            room.roomTeams.rules,
		TeamAssignments:       copyTeamAssignments(room.roomTeams.assignments),
		TeamAssignmentsLocked: room.roomTeams.locked,
		MaxPlayers:            room.MaxPlayers,
		ResolvedSummary:       resolvedSummary,
		HasResolvedMatch:      hasResolvedMatch,
	}
}
