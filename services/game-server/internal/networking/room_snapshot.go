package networking

import (
	"encoding/json"
	"sort"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func BuildRoomSnapshot(room *rooms.Room, localMemberID string) game.RoomSnapshot {
	memberSnapshot := room.MembersSnapshot()
	sort.Slice(memberSnapshot, func(left, right int) bool {
		return memberSnapshot[left].SessionID < memberSnapshot[right].SessionID
	})

	members := make([]game.RoomMemberState, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		members = append(members, game.RoomMemberState{
			MemberID:  member.SessionID,
			Ready:     member.Ready,
			Connected: member.Connected,
		})
	}

	return game.RoomSnapshot{
		Type:          game.PacketTypeRoomSnapshot,
		RoomCode:      room.ID,
		RoomState:     string(room.State),
		Members:       members,
		LocalMemberID: localMemberID,
		MaxPlayers:    rooms.MaxPlayersPerRoom,
	}
}

func (session *webSocketSession) EnqueueRoomSnapshot(room *rooms.Room) {
	packet := BuildRoomSnapshot(room, session.currentMemberID)
	payload, err := json.Marshal(packet)
	if err != nil {
		logging.Network.Error("room snapshot marshal failed", err,
			logging.FieldRoomID, room.ID,
			"session_id", session.sessionID,
		)
		return
	}

	session.enqueue(payload)
}

func BroadcastRoomSnapshot(room *rooms.Room) {
	memberSnapshot := room.MembersSnapshot()
	sort.Slice(memberSnapshot, func(left, right int) bool {
		return memberSnapshot[left].SessionID < memberSnapshot[right].SessionID
	})

	memberIDs := make([]string, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		memberIDs = append(memberIDs, member.SessionID)
	}

	sessions := snapshotRoomSessions(room, memberIDs)

	for _, session := range sessions {
		session.EnqueueRoomSnapshot(room)
	}
}
