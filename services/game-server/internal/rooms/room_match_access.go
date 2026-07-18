package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchresults"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
	"github.com/google/uuid"
)

func (room *Room) GameInstance() *game.Game {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.Game()
}

func (room *Room) SetGameInstance(gameInstance *game.Game) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.match.SetGame(gameInstance)
}

func (room *Room) ClearGameInstance() {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.match.ClearGame()
}

// SetActivePlayerCount is retained for compatibility and test setup.
func (room *Room) SetActivePlayerCount(count int) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.match.SetActivePlayers(count)
}

func (room *Room) CurrentMatchID() string {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.CurrentMatchID()
}

func (room *Room) CurrentMatchTraceID() string {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.CurrentTraceID()
}

func (room *Room) CurrentOrCreateMatchTraceID() string {
	room.mu.Lock()
	defer room.mu.Unlock()
	if traceID := room.match.CurrentTraceID(); traceID != "" {
		return traceID
	}
	traceID := uuid.NewString()
	room.match.currentTraceID = traceID
	return traceID
}

func (room *Room) ResolvedMatchSummary() (playerdata.MatchResultSummary, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.ResolvedSummary()
}

func (room *Room) MatchSummary() (matchresults.MatchSummary, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.MatchSummary()
}

func (room *Room) MatchResultDispatch() (matchresults.DispatchSlices, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.MatchDispatch()
}

func (room *Room) MatchResultReported() bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.MatchResultReported()
}

func (room *Room) MarkMatchResultReported() {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.match.MarkMatchResultReported()
}
