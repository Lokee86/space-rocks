package rooms

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

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

	if roomErr := room.StartGameForSession(sessionID, game.New); roomErr != nil {
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

	if roomErr := room.ResetToLobbyForSession(sessionID); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}
