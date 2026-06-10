package rooms

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
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

func (room *Room) ResolvedMatchSummary() (playerdata.MatchResultSummary, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.ResolvedSummary()
}
