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

	if roomErr := room.ValidateStart(memberID); roomErr != nil {
		return nil, roomErr
	}
	if roomErr := room.MarkStarting(); roomErr != nil {
		return nil, roomErr
	}
	if room.Game == nil {
		room.Game = game.New()
	}
	room.Game.Start()
	if roomErr := room.MarkInGame(); roomErr != nil {
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

	room.Game = game.New()
	room.Game.Start()
	room.State = RoomStateInGame

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
