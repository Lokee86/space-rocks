package rooms

import (
	"strconv"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
)

type roomMatch struct {
	game                 *game.Game
	activePlayers        int
	matchNumber          int
	currentMatchID       string
	resolvedSummary      *playerdata.MatchResultSummary
	matchResultReported  bool
}

func newRoomMatch(gameInstance *game.Game) *roomMatch {
	return &roomMatch{
		game: gameInstance,
	}
}

func (rm *roomMatch) Game() *game.Game {
	return rm.game
}

func (rm *roomMatch) SetGame(gameInstance *game.Game) {
	rm.game = gameInstance
}

func (rm *roomMatch) ClearGame() {
	rm.game = nil
}

func (rm *roomMatch) ActivePlayers() int {
	return rm.activePlayers
}

func (rm *roomMatch) SetActivePlayers(count int) {
	rm.activePlayers = count
}

func (rm *roomMatch) BeginNextMatch(roomID string) string {
	rm.matchNumber++
	rm.currentMatchID = roomID + "-match-" + strconv.Itoa(rm.matchNumber)
	rm.matchResultReported = false
	rm.ClearResolvedSummary()
	return rm.currentMatchID
}

func (rm *roomMatch) CurrentMatchID() string {
	return rm.currentMatchID
}

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
}

func (rm *roomMatch) MarkMatchResultReported() {
	rm.matchResultReported = true
}

func (rm *roomMatch) MatchResultReported() bool {
	return rm.matchResultReported
}
