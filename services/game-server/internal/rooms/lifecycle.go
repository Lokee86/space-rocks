package rooms

import "github.com/Lokee86/space-rocks/server/internal/game"

func (manager *RoomManager) StartRoomGame(roomID string, memberID string) (*Room, *RoomDomainError) {
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

	if roomErr := room.StartGameForMember(memberID, game.New); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}

func (manager *RoomManager) CreateStartedSinglePlayerRoom(memberID string) (*Room, *RoomDomainError) {
	room, err := manager.CreateSinglePlayerRoom(memberID)
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

func (manager *RoomManager) ReturnRoomToLobby(roomID string, memberID string) (*Room, *RoomDomainError) {
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

	if roomErr := room.ResetToLobby(memberID); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}
