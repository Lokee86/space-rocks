package rooms

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"

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

	if roomErr := room.StartGameForSession(sessionID, manager.newGame); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}

func (manager *RoomManager) CreateStartedSinglePlayerRoom(sessionID string) (*Room, *RoomDomainError) {
	return manager.CreateStartedSinglePlayerRoomWithModeConfig(sessionID, modes.DefaultRoomModeConfig())
}

func (manager *RoomManager) CreateStartedSinglePlayerRoomWithModeConfig(sessionID string, modeConfig modes.RoomModeConfig) (*Room, *RoomDomainError) {
	room, err := manager.CreateSinglePlayerRoomWithModeConfig(sessionID, modeConfig)
	if err != nil {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Could not create room.",
		}
	}

	if roomErr := room.StartSinglePlayerGame(manager.newGame); roomErr != nil {
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
