package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

type roomMatch struct {
	game          *game.Game
	activePlayers int
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
