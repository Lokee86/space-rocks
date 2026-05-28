package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

func (manager *RoomManager) StartRoomGame(roomID string, sessionID string) (*Room, *RoomDomainError) {
	roomID = NormalizeRoomID(roomID)

	manager.mu.Lock()
	room, ok := manager.rooms[roomID]
	manager.mu.Unlock()
	if !ok {
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomNotFound,
			Message: "Room was not found.",
		}
	}

	playerID, ok := room.PlayerIDForSession(sessionID)
	if !ok {
		return nil, &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	if roomErr := room.StartGameForMember(playerID, game.New); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}

func (manager *RoomManager) CreateStartedSinglePlayerRoom(sessionID string) (*Room, *RoomDomainError) {
	room, err := manager.CreateSinglePlayerRoom(sessionID)
	if err != nil {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Could not create room.",
		}
	}

	if roomErr := room.StartSinglePlayerGame(game.New); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}

func (manager *RoomManager) ReturnRoomToLobby(roomID string, sessionID string) (*Room, *RoomDomainError) {
	roomID = NormalizeRoomID(roomID)

	manager.mu.Lock()
	room, ok := manager.rooms[roomID]
	manager.mu.Unlock()
	if !ok {
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomNotFound,
			Message: "Room was not found.",
		}
	}

	playerID, ok := room.PlayerIDForSession(sessionID)
	if !ok {
		return nil, &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	if roomErr := room.ResetToLobby(playerID); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}
