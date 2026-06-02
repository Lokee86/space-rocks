package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

func (room *Room) StartGameForMember(playerID string, newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if roomErr := room.validateStartLocked(playerID); roomErr != nil {
		return roomErr
	}

	if roomErr := room.markStartingLocked(); roomErr != nil {
		return roomErr
	}
	if room.Game == nil {
		room.Game = newGame()
	}
	room.Game.Start()
	if roomErr := room.markInGameLocked(); roomErr != nil {
		return roomErr
	}

	return nil
}

func (room *Room) MarkStarting() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.markStartingLocked()
}

func (room *Room) MarkInGame() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.markInGameLocked()
}

func (room *Room) markStartingLocked() *RoomDomainError {
	if room.State != RoomStateLobby {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only start from the lobby."}
	}

	room.State = RoomStateStarting
	return nil
}

func (room *Room) markInGameLocked() *RoomDomainError {
	if room.State != RoomStateStarting {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room can only enter in-game from starting."}
	}

	room.State = RoomStateInGame
	return nil
}

func (room *Room) MarkGameOver() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateInGame {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room can only move to game over from in-game.",
		}
	}

	room.State = RoomStateGameOver
	return nil
}

func (room *Room) MarkGameOverIfComplete() bool {
	if room == nil || room.State != RoomStateInGame || !room.IsGameOver() {
		return false
	}

	return room.MarkGameOver() == nil
}

func (room *Room) ResetToLobby(playerID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if _, ok := room.Members[playerID]; !ok {
		return &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	if room.State != RoomStateGameOver {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room can only return to lobby from game over.",
		}
	}

	for _, member := range room.Members {
		member.SetReady(false)
	}
	if room.Game != nil {
		room.Game.Stop()
	}
	room.Game = nil
	room.State = RoomStateLobby
	return nil
}

func (room *Room) IsGameOver() bool {
	if room == nil || room.State != RoomStateInGame || room.Game == nil {
		return false
	}

	return room.Game.MatchDecision().IsOver
}

func (room *Room) StartSinglePlayerGame(newGame func() *game.Game) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Game can only be started from the lobby."}
	}
	if len(room.Members) == 0 {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	if roomErr := room.markStartingLocked(); roomErr != nil {
		return roomErr
	}
	if room.Game == nil {
		room.Game = newGame()
	}
	room.Game.Start()
	if roomErr := room.markInGameLocked(); roomErr != nil {
		return roomErr
	}

	return nil
}
