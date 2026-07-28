package rooms

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

type GameFactory func() *game.Game

func normalizeGameFactory(factory GameFactory) GameFactory {
	if factory != nil {
		return factory
	}
	return game.New
}

func (manager *RoomManager) newGame() *game.Game {
	return manager.gameFactory()
}
