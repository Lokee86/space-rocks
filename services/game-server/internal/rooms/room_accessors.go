package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

func (room *Room) CurrentState() RoomState {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.State
}

func (room *Room) CurrentGame() *game.Game {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.Game
}

func (room *Room) ActivePlayerCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.ActivePlayers
}
