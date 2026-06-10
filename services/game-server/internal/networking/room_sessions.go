package networking

import (
	"strconv"
	"sync"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

var roomSessions = struct {
	sync.Mutex
	byRoom map[*rooms.Room]map[string]*webSocketSession
}{
	byRoom: make(map[*rooms.Room]map[string]*webSocketSession),
}

func addSessionMember(room *rooms.Room, sessionID string, session *webSocketSession) {
	room.AddMember(rooms.NewRoomMember(sessionID))
	if accountID := accountIDForSession(session); accountID != "" {
		room.SetMemberAccountIDForSession(sessionID, accountID)
	}
	attachRoomSession(room, sessionID, session)
}

func attachRoomSession(room *rooms.Room, sessionID string, session *webSocketSession) {
	if session == nil {
		return
	}

	roomSessions.Lock()
	defer roomSessions.Unlock()

	sessions := roomSessions.byRoom[room]
	if sessions == nil {
		sessions = make(map[string]*webSocketSession)
		roomSessions.byRoom[room] = sessions
	}
	sessions[sessionID] = session
}

func detachRoomSession(room *rooms.Room, sessionID string) {
	roomSessions.Lock()
	defer roomSessions.Unlock()

	sessions := roomSessions.byRoom[room]
	if sessions == nil {
		return
	}
	delete(sessions, sessionID)
	if len(sessions) == 0 {
		delete(roomSessions.byRoom, room)
	}
}

func snapshotRoomSessions(room *rooms.Room, memberIDs []string) []*webSocketSession {
	roomSessions.Lock()
	defer roomSessions.Unlock()

	sessionsByMember := roomSessions.byRoom[room]
	if sessionsByMember == nil {
		return nil
	}

	sessions := make([]*webSocketSession, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if session := sessionsByMember[memberID]; session != nil {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

func accountIDForSession(session *webSocketSession) string {
	if session == nil {
		return ""
	}

	identity := session.SessionIdentity()
	if !identity.IsAuthenticatedAccount() {
		return ""
	}

	return strconv.FormatInt(identity.AccountUserID, 10)
}
