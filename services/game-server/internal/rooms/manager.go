package rooms

import (
	"fmt"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/logging"
)

type RoomManager struct {
	mu           sync.Mutex
	rooms        map[string]*Room
	cleanupDelay time.Duration
}

type RoomDomainError struct {
	Code    string
	Message string
}

type LeaveRoomResult struct {
	Room             *Room
	RoomID           string
	MemberID         string
	RemainingMembers int
}

func (err *RoomDomainError) Error() string {
	return err.Message
}

func NewRoomManager() *RoomManager {
	return NewRoomManagerWithCleanupDelay(RoomCleanupGraceTime)
}

func NewRoomManagerWithCleanupDelay(cleanupDelay time.Duration) *RoomManager {
	manager := &RoomManager{
		rooms:        make(map[string]*Room),
		cleanupDelay: cleanupDelay,
	}

	return manager
}

func (manager *RoomManager) Find(roomID string) (*Room, bool) {
	roomID = NormalizeRoomID(roomID)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	room, ok := manager.rooms[roomID]
	return room, ok
}

func (manager *RoomManager) CreateLobbyRoom() (*Room, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for attempts := 0; attempts < 16; attempts++ {
		roomID, err := GenerateRoomCode()
		if err != nil {
			return nil, err
		}
		if _, ok := manager.rooms[roomID]; ok {
			continue
		}

		room := NewRoom(roomID, RoomStateLobby, nil)
		manager.rooms[roomID] = room
		logging.Rooms.Debug("lobby room created", logging.FieldRoomID, roomID)

		return room, nil
	}

	return nil, fmt.Errorf("generate unique room code")
}

func (manager *RoomManager) CreateSinglePlayerRoom(memberID string) (*Room, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for attempts := 0; attempts < 16; attempts++ {
		roomID, err := GenerateRoomCode()
		if err != nil {
			return nil, err
		}
		if _, ok := manager.rooms[roomID]; ok {
			continue
		}

		room := NewRoom(roomID, RoomStateLobby, nil)
		room.SetJoinable(false)
		room.AddMemberID(memberID)
		manager.rooms[roomID] = room
		logging.Rooms.Debug("single-player room created", logging.FieldRoomID, roomID)

		return room, nil
	}

	return nil, fmt.Errorf("generate unique room code")
}

func (manager *RoomManager) JoinRoom(memberID string, roomCode string) (*Room, *RoomDomainError) {
	roomCode = NormalizeRoomCode(roomCode)
	if !IsValidRoomCode(roomCode) {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomCode,
			Message: "Room code is invalid.",
		}
	}

	manager.mu.Lock()
	room, ok := manager.rooms[roomCode]
	manager.mu.Unlock()
	if !ok {
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomNotFound,
			Message: "Room was not found.",
		}
	}

	if roomErr := room.JoinMember(memberID); roomErr != nil {
		return nil, roomErr
	}
	return room, nil

	switch room.CurrentState() {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomInGame,
			Message: "Room is already in game.",
		}
	case RoomStateClosed:
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomClosed,
			Message: "Room is closed.",
		}
	default:
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room is not joinable.",
		}
	}

	if !room.IsJoinable() {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room is not joinable.",
		}
	}

	if room.IsFull() {
		return nil, &RoomDomainError{
			Code:    RoomErrorRoomFull,
			Message: "Room is full.",
		}
	}

	room.AddMemberID(memberID)
	return room, nil
}

func (manager *RoomManager) LeaveRoom(roomID string, memberID string) (*LeaveRoomResult, *RoomDomainError) {
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

	if memberID != "" {
		room.RemoveMember(memberID)
	}

	return &LeaveRoomResult{
		Room:             room,
		RoomID:           roomID,
		MemberID:         memberID,
		RemainingMembers: room.MemberCount(),
	}, nil
}

func (manager *RoomManager) SetReady(roomID string, memberID string, ready bool) (*Room, *RoomDomainError) {
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

	if room.CurrentState() != RoomStateLobby {
		return nil, &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Ready state can only be changed in the lobby.",
		}
	}

	if !room.SetMemberReady(memberID, ready) {
		return nil, &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	return room, nil
}

func (manager *RoomManager) StopAll() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for roomID, room := range manager.rooms {
		if room.CleanupTimer != nil {
			room.CleanupTimer.Stop()
			room.CleanupTimer = nil
		}
		logging.Rooms.Debug("room stopped", logging.FieldRoomID, roomID)
		if room.Game != nil {
			room.Game.Stop()
		}
		delete(manager.rooms, roomID)
	}
}

func (manager *RoomManager) RoomCount() int {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return len(manager.rooms)
}

func (manager *RoomManager) ScheduleCleanupIfEmpty(roomID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	room, ok := manager.rooms[roomID]
	if !ok || !room.ShouldCleanup() {
		return
	}

	manager.scheduleCleanupLocked(roomID, room)
}

func (manager *RoomManager) cleanupEmptyRoom(roomID string, cleanupVersion int) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	room, ok := manager.rooms[roomID]
	if !ok {
		logging.Rooms.Debug("room cleanup skipped; room already removed",
			logging.FieldRoomID, roomID,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	activePlayers := room.ActivePlayerCount()
	if activePlayers > 0 {
		logging.Rooms.Debug("room cleanup skipped; room active",
			logging.FieldRoomID, roomID,
			"active_players", activePlayers,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	if !room.ShouldCleanup() {
		logging.Rooms.Debug("room cleanup skipped; room has members",
			logging.FieldRoomID, roomID,
			"members", room.MemberCount(),
			"cleanup_version", cleanupVersion,
		)
		return
	}
	if !room.CleanupVersionMatches(cleanupVersion) {
		logging.Rooms.Debug("room cleanup skipped; stale cleanup",
			logging.FieldRoomID, roomID,
			"cleanup_version", cleanupVersion,
			"current_cleanup_version", room.CurrentCleanupVersion(),
		)
		return
	}

	room.StopGameIfPresent()
	delete(manager.rooms, roomID)
	logging.Rooms.Debug("room cleaned up",
		logging.FieldRoomID, roomID,
		"cleanup_version", cleanupVersion,
	)
}

func (manager *RoomManager) scheduleCleanupLocked(roomID string, room *Room) {
	cleanupVersion := room.ScheduleCleanupTimer(manager.cleanupDelay, func(cleanupVersion int) {
		manager.cleanupEmptyRoom(roomID, cleanupVersion)
	})
	logging.Rooms.Debug("room cleanup scheduled",
		logging.FieldRoomID, roomID,
		"cleanup_delay", manager.cleanupDelay.String(),
		"cleanup_version", cleanupVersion,
	)
}
