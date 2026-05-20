package networking

import (
	"strings"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

const (
	DefaultRoomID        = "default"
	RoomCleanupGraceTime = 30 * time.Second
)

type Room struct {
	ID             string
	Game           *game.Game
	activePlayers  int
	cleanupTimer   *time.Timer
	cleanupVersion int
}

type RoomManager struct {
	mu           sync.Mutex
	rooms        map[string]*Room
	cleanupDelay time.Duration
}

func NewRoomManager() *RoomManager {
	return NewRoomManagerWithCleanupDelay(RoomCleanupGraceTime)
}

func NewRoomManagerWithCleanupDelay(cleanupDelay time.Duration) *RoomManager {
	manager := &RoomManager{
		rooms:        make(map[string]*Room),
		cleanupDelay: cleanupDelay,
	}
	manager.GetOrCreate(DefaultRoomID)

	return manager
}

func (manager *RoomManager) DefaultRoom() *Room {
	return manager.GetOrCreate(DefaultRoomID)
}

func (manager *RoomManager) Join(roomID string) (*Room, func()) {
	roomID = normalizeRoomID(roomID)

	manager.mu.Lock()
	room := manager.getOrCreateLocked(roomID)
	room.activePlayers++
	room.cleanupVersion++
	if room.cleanupTimer != nil {
		room.cleanupTimer.Stop()
		room.cleanupTimer = nil
		logging.Rooms.Debug("room cleanup canceled",
			logging.FieldRoomID, roomID,
			"active_players", room.activePlayers,
		)
	}
	logging.Rooms.Debug("room joined",
		logging.FieldRoomID, roomID,
		"active_players", room.activePlayers,
	)
	manager.mu.Unlock()

	return room, func() {
		manager.leave(roomID)
	}
}

func (manager *RoomManager) GetOrCreate(roomID string) *Room {
	roomID = normalizeRoomID(roomID)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.getOrCreateLocked(roomID)
}

func (manager *RoomManager) StopAll() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for roomID, room := range manager.rooms {
		if room.cleanupTimer != nil {
			room.cleanupTimer.Stop()
			room.cleanupTimer = nil
		}
		logging.Rooms.Debug("room stopped", logging.FieldRoomID, roomID)
		room.Game.Stop()
		delete(manager.rooms, roomID)
	}
}

func (manager *RoomManager) RoomCount() int {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return len(manager.rooms)
}

func (manager *RoomManager) getOrCreateLocked(roomID string) *Room {
	if room, ok := manager.rooms[roomID]; ok {
		return room
	}

	room := &Room{
		ID:   roomID,
		Game: game.New(),
	}
	room.Game.Start()
	manager.rooms[roomID] = room
	logging.Rooms.Debug("room created", logging.FieldRoomID, roomID)

	return room
}

func (manager *RoomManager) leave(roomID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	room, ok := manager.rooms[roomID]
	if !ok {
		return
	}

	if room.activePlayers > 0 {
		room.activePlayers--
	}
	logging.Rooms.Debug("room left",
		logging.FieldRoomID, roomID,
		"active_players", room.activePlayers,
	)
	if room.activePlayers > 0 {
		return
	}

	room.cleanupVersion++
	cleanupVersion := room.cleanupVersion
	if room.cleanupTimer != nil {
		room.cleanupTimer.Stop()
	}
	room.cleanupTimer = time.AfterFunc(manager.cleanupDelay, func() {
		manager.cleanupEmptyRoom(roomID, cleanupVersion)
	})
	logging.Rooms.Debug("room cleanup scheduled",
		logging.FieldRoomID, roomID,
		"cleanup_delay", manager.cleanupDelay.String(),
		"cleanup_version", cleanupVersion,
	)
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
	if room.activePlayers > 0 {
		logging.Rooms.Debug("room cleanup skipped; room active",
			logging.FieldRoomID, roomID,
			"active_players", room.activePlayers,
			"cleanup_version", cleanupVersion,
		)
		return
	}
	if room.cleanupVersion != cleanupVersion {
		logging.Rooms.Debug("room cleanup skipped; stale cleanup",
			logging.FieldRoomID, roomID,
			"cleanup_version", cleanupVersion,
			"current_cleanup_version", room.cleanupVersion,
		)
		return
	}

	room.Game.Stop()
	delete(manager.rooms, roomID)
	logging.Rooms.Debug("room cleaned up",
		logging.FieldRoomID, roomID,
		"cleanup_version", cleanupVersion,
	)
}

func normalizeRoomID(roomID string) string {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return DefaultRoomID
	}

	return roomID
}
