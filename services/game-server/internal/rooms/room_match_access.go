package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

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
