package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

func (room *Room) StartGameForMember(playerID string, newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if _, ok := room.Members[playerID]; !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	if playerID != room.OwnerID {
		return &RoomDomainError{Code: RoomErrorNotRoomOwner, Message: "Only the room owner can start the game."}
	}

	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{Code: RoomErrorRoomInGame, Message: "Room is already in game."}
	default:
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Game can only be started from the lobby."}
	}

	for _, connectedMember := range room.Members {
		if !connectedMember.Ready {
			return &RoomDomainError{Code: RoomErrorNotReady, Message: "All connected members must be ready."}
		}
	}

	if room.Game == nil {
		room.Game = newGame()
	}
	room.Game.Start()
	room.State = RoomStateInGame

	return nil
}

func (room *Room) StartSinglePlayerGame(newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		room.Game = newGame()
	}
	room.Game.Start()
	room.State = RoomStateInGame

	return nil
}
