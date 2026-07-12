package rooms

import (
	"fmt"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
)

type RoomManager struct {
	mu           sync.Mutex
	rooms        map[string]*Room
	cleanupDelay time.Duration
}

var stopRoomForCleanup = func(room *Room) { room.StopGameIfPresent() }

type RoomDomainError struct {
	Code    string
	Message string
}

type LeaveRoomResult struct {
	Room             *Room
	RoomID           string
	SessionID        string
	RemovedMember    RoomMember
	MemberRemoved    bool
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

func (manager *RoomManager) CreateSinglePlayerRoom(sessionID string) (*Room, error) {
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
		room.AddMemberSessionID(sessionID)
		manager.rooms[roomID] = room
		logging.Rooms.Debug("single-player room created", logging.FieldRoomID, roomID)

		return room, nil
	}

	return nil, fmt.Errorf("generate unique room code")
}

func (manager *RoomManager) JoinRoom(sessionID string, roomCode string) (*Room, *RoomDomainError) {
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

	if roomErr := room.JoinMember(sessionID); roomErr != nil {
		return nil, roomErr
	}
	return room, nil
}

func (manager *RoomManager) LeaveRoom(roomID string, sessionID string) (*LeaveRoomResult, *RoomDomainError) {
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

	removedMember, remainingMembers, memberRemoved := room.RemoveMemberForSession(sessionID)

	return &LeaveRoomResult{
		Room:             room,
		RoomID:           roomID,
		SessionID:        sessionID,
		RemovedMember:    removedMember,
		MemberRemoved:    memberRemoved,
		RemainingMembers: remainingMembers,
	}, nil
}

func (manager *RoomManager) SetReady(roomID string, sessionID string, ready bool) (*Room, *RoomDomainError) {
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

	if roomErr := room.SetReadyForSessionInLobby(sessionID, ready); roomErr != nil {
		return nil, roomErr
	}

	return room, nil
}

func (manager *RoomManager) StopAll() {
	manager.mu.Lock()
	rooms := make(map[string]*Room, len(manager.rooms))
	for roomID, room := range manager.rooms {
		rooms[roomID] = room
		delete(manager.rooms, roomID)
	}
	manager.mu.Unlock()
	for roomID, room := range rooms {
		room.StopCleanupTimer()
		logging.Rooms.Debug("room stopped", logging.FieldRoomID, roomID)
		room.StopGameIfPresent()
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

	room, ok := manager.rooms[roomID]
	if !ok {
		manager.mu.Unlock()
		logging.Rooms.Debug("room cleanup skipped; room already removed",
			logging.FieldRoomID, roomID,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	population := room.Population()
	if population.ActivePlayers > 0 {
		manager.mu.Unlock()
		logging.Rooms.Debug("room cleanup skipped; room active",
			logging.FieldRoomID, roomID,
			"active_players", population.ActivePlayers,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	if population.Members != 0 {
		manager.mu.Unlock()
		logging.Rooms.Debug("room cleanup skipped; room has members",
			logging.FieldRoomID, roomID,
			"members", population.Members,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	if !room.CleanupVersionMatches(cleanupVersion) {
		manager.mu.Unlock()
		logging.Rooms.Debug("room cleanup skipped; stale cleanup",
			logging.FieldRoomID, roomID,
			"cleanup_version", cleanupVersion,
			"current_cleanup_version", room.CurrentCleanupVersion(),
		)
		return
	}

	delete(manager.rooms, roomID)
	manager.mu.Unlock()
	stopRoomForCleanup(room)
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
