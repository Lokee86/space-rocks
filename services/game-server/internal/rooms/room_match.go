package rooms

import (
	"strconv"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

type roomMatch struct {
	game                 *game.Game
	activePlayers        int
	activeSessionIDs     map[string]struct{}
	matchNumber          int
	currentMatchID       string
	resolvedSummary      *playerdata.MatchResultSummary
	matchResultReported  bool
	matchResultReporting bool
}

func newRoomMatch(gameInstance *game.Game) *roomMatch {
	return &roomMatch{game: gameInstance, activeSessionIDs: make(map[string]struct{})}
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
	rm.ResetActiveSessions()
	rm.matchResultReported = false
	rm.matchResultReporting = false
	rm.ClearResolvedSummary()
	return rm.currentMatchID
}

func (rm *roomMatch) CurrentMatchID() string { return rm.currentMatchID }

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
	rm.matchResultReporting = false
}

func (rm *roomMatch) MarkMatchResultReported() {
	rm.matchResultReported = true
	rm.matchResultReporting = false
}

func (rm *roomMatch) MatchResultReported() bool { return rm.matchResultReported }
