package networking

import (
	"sort"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func BuildRoomSnapshot(room *rooms.Room, localSessionID string) game.RoomSnapshot {
	memberSnapshot := room.MembersSnapshot()
	sort.Slice(memberSnapshot, func(left, right int) bool {
		return memberSnapshot[left].SessionID < memberSnapshot[right].SessionID
	})

	members := make([]game.RoomMemberState, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		members = append(members, game.RoomMemberState{
			PlayerID:  member.PlayerID,
			Ready:     member.Ready,
			Connected: member.Connected,
		})
	}

	localPlayerID, _ := room.PlayerIDForSession(localSessionID)

	return game.RoomSnapshot{
		Type:          game.PacketTypeRoomSnapshot,
		RoomCode:      room.ID,
		RoomState:     string(room.State),
		Members:       members,
		LocalPlayerID: localPlayerID,
		OwnerID:       room.OwnerID(),
		MaxPlayers:    rooms.MaxPlayersPerRoom,
	}
}

func (session *webSocketSession) EnqueueRoomSnapshot(room *rooms.Room) {
	packet := BuildRoomSnapshot(room, session.sessionID)
	payload, err := packetcodec.Encode(packet)
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
