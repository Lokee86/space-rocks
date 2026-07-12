package networking

import "github.com/Lokee86/space-rocks/services/game-server/internal/rooms"

// SessionContext is an immutable snapshot of the session's coherent room state.
type SessionContext struct {
	Room         *rooms.Room
	RoomID       string
	GamePlayerID string
}

func (session *webSocketSession) sessionContext() SessionContext {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.context
}

func (session *webSocketSession) bindRoom(room *rooms.Room) {
	context := SessionContext{Room: room}
	if room != nil {
		context.RoomID = room.ID
	}
	session.mu.Lock()
	session.context = context
	session.mu.Unlock()
}

func (session *webSocketSession) clearRoomContext() {
	session.mu.Lock()
	session.context = SessionContext{}
	session.mu.Unlock()
}

func (session *webSocketSession) clearRoomContextIfMatch(expected SessionContext) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.context != expected {
		return false
	}
	session.context = SessionContext{}
	return true
}

func (session *webSocketSession) setGamePlayerIDForRoom(expectedRoom *rooms.Room, playerID string) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.context.Room != expectedRoom {
		return false
	}
	session.context.GamePlayerID = playerID
	return true
}

func (session *webSocketSession) clearGamePlayerIDForRoom(expectedRoom *rooms.Room) bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.context.Room != expectedRoom {
		return false
	}
	session.context.GamePlayerID = ""
	return true
}

func (session *webSocketSession) sessionContextMatches(context SessionContext) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.context == context
}
