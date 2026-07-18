package rooms

import (
	"strconv"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchresults"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
	"github.com/google/uuid"
)

type roomMatch struct {
	game                  *game.Game
	activePlayers         int
	activeSessionIDs      map[string]struct{}
	participantIdentities map[string]matchParticipantIdentity
	matchNumber           int
	currentMatchID        string
	currentTraceID        string
	resolvedSummary       *playerdata.MatchResultSummary
	matchSummary          *matchresults.MatchSummary
	matchDispatch         *matchresults.DispatchSlices
	endOfMatchFlow        *matchresults.EndOfMatchFlow
	matchResultReported   bool
	matchResultReporting  bool
}

type matchParticipantIdentity struct {
	AccountID      string
	LocalProfileID string
	IsBot          bool
}

func newRoomMatch(gameInstance *game.Game) *roomMatch {
	return &roomMatch{
		game:                  gameInstance,
		activeSessionIDs:      make(map[string]struct{}),
		participantIdentities: make(map[string]matchParticipantIdentity),
		endOfMatchFlow:        matchresults.NewEndOfMatchFlow(),
	}
}

func (rm *roomMatch) Game() *game.Game { return rm.game }

func (rm *roomMatch) SetGame(gameInstance *game.Game) { rm.game = gameInstance }

func (rm *roomMatch) ClearGame() { rm.game = nil }

func (rm *roomMatch) ActivePlayers() int { return rm.activePlayers }

func (rm *roomMatch) SetActivePlayers(count int) {
	if count < 0 {
		count = 0
	}
	rm.activePlayers = count
	if count == 0 {
		rm.ResetActiveSessions()
	}
	// Nonzero values are compatibility/test setup and do not fabricate ownership.
}

func (rm *roomMatch) SessionActive(sessionID string) bool {
	_, active := rm.activeSessionIDs[sessionID]
	return active
}

func (rm *roomMatch) ActivateSession(sessionID string) bool {
	if rm.SessionActive(sessionID) {
		return false
	}
	rm.activeSessionIDs[sessionID] = struct{}{}
	rm.activePlayers++
	return true
}

func (rm *roomMatch) DeactivateSession(sessionID string) bool {
	if !rm.SessionActive(sessionID) {
		return false
	}
	delete(rm.activeSessionIDs, sessionID)
	if rm.activePlayers > 0 {
		rm.activePlayers--
	}
	return true
}

func (rm *roomMatch) ResetActiveSessions() {
	rm.activeSessionIDs = make(map[string]struct{})
	rm.activePlayers = 0
}

func (rm *roomMatch) BeginNextMatch(roomID string) string {
	rm.matchNumber++
	rm.currentMatchID = roomID + "-match-" + strconv.Itoa(rm.matchNumber)
	rm.currentTraceID = uuid.NewString()
	rm.ResetActiveSessions()
	rm.participantIdentities = make(map[string]matchParticipantIdentity)
	rm.matchResultReported = false
	rm.matchResultReporting = false
	rm.endOfMatchFlow = matchresults.NewEndOfMatchFlow()
	rm.matchSummary = nil
	rm.matchDispatch = nil
	rm.ClearResolvedSummary()
	if rm.game != nil {
		rm.game.SetMatchContext(rm.currentMatchID, rm.currentTraceID)
	}
	return rm.currentMatchID
}

func (rm *roomMatch) RememberParticipantIdentity(member RoomMember) {
	if member.PlayerID == "" {
		return
	}
	rm.participantIdentities[member.PlayerID] = matchParticipantIdentity{
		AccountID:      member.AccountID,
		LocalProfileID: member.LocalProfileID,
		IsBot:          member.IsBot,
	}
}

func (rm *roomMatch) CurrentMatchID() string { return rm.currentMatchID }

func (rm *roomMatch) CurrentTraceID() string { return rm.currentTraceID }

func (rm *roomMatch) SetResolvedSummary(summary playerdata.MatchResultSummary) {
	rm.resolvedSummary = &summary
}

func (rm *roomMatch) ResolvedSummary() (playerdata.MatchResultSummary, bool) {
	if rm.resolvedSummary == nil {
		return playerdata.MatchResultSummary{}, false
	}
	return *rm.resolvedSummary, true
}

func (rm *roomMatch) ClearResolvedSummary() {
	rm.resolvedSummary = nil
	rm.matchSummary = nil
	rm.matchDispatch = nil
	rm.matchResultReporting = false
}

func (rm *roomMatch) ResolveEndOfMatch(input matchresults.BuildInput) (matchresults.MatchSummary, matchresults.DispatchSlices, bool, error) {
	if rm.endOfMatchFlow == nil {
		rm.endOfMatchFlow = matchresults.NewEndOfMatchFlow()
	}
	summary, emitted, err := rm.endOfMatchFlow.Run(input)
	if err != nil {
		return matchresults.MatchSummary{}, matchresults.DispatchSlices{}, false, err
	}
	if rm.matchSummary == nil {
		stored := summary
		dispatch := (matchresults.MatchSummaryDispatcher{}).Dispatch(summary)
		if rm.resolvedSummary != nil {
			dispatch.Persistence = *rm.resolvedSummary
		} else {
			rm.SetResolvedSummary(dispatch.Persistence)
		}
		rm.matchSummary = &stored
		rm.matchDispatch = &dispatch
	}
	return *rm.matchSummary, *rm.matchDispatch, emitted, nil
}

func (rm *roomMatch) MatchSummary() (matchresults.MatchSummary, bool) {
	if rm.endOfMatchFlow == nil {
		return matchresults.MatchSummary{}, false
	}
	return rm.endOfMatchFlow.Summary()
}

func (rm *roomMatch) MatchDispatch() (matchresults.DispatchSlices, bool) {
	if rm.matchDispatch == nil {
		return matchresults.DispatchSlices{}, false
	}
	return *rm.matchDispatch, true
}

func (rm *roomMatch) MarkMatchResultReported() {
	rm.matchResultReported = true
	rm.matchResultReporting = false
}

func (rm *roomMatch) MatchResultReported() bool { return rm.matchResultReported }
