package networking

import (
	"strings"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
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
	}
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
}

func (manager *RoomManager) cleanupEmptyRoom(roomID string, cleanupVersion int) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	room, ok := manager.rooms[roomID]
	if !ok || room.activePlayers > 0 || room.cleanupVersion != cleanupVersion {
		return
	}

	room.Game.Stop()
	delete(manager.rooms, roomID)
}

func normalizeRoomID(roomID string) string {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return DefaultRoomID
	}

	return roomID
}
