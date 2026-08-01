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
	return manager.CreateStartedSinglePlayerRoomWithConfig(sessionID, modeConfig, 1)
}

func (manager *RoomManager) CreateStartedSinglePlayerRoomWithConfig(sessionID string, modeConfig modes.RoomModeConfig, maxPlayers int) (*Room, *RoomDomainError) {
	capacity := singlePlayerRoomCapacity(modeConfig, maxPlayers)
	room, err := manager.CreateSinglePlayerRoomWithConfig(sessionID, modeConfig, capacity)
	if err != nil {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Could not create room.",
		}
	}

	if modes.NormalizeRoomModeConfig(modeConfig).PresetID == modes.PresetDeathmatch {
		for botIndex := 1; botIndex < capacity; botIndex++ {
			if _, roomErr := room.AddBotForOwnerSession(sessionID); roomErr != nil {
				return nil, roomErr
			}
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
