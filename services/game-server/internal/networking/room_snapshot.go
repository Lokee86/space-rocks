package networking

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func BuildRoomSnapshot(room *rooms.Room, localSessionID string) game.RoomSnapshot {
	projection := room.SnapshotForSession(localSessionID)
	memberSnapshot := projection.Members
	sort.Slice(memberSnapshot, func(left, right int) bool {
		return memberSnapshot[left].SessionID < memberSnapshot[right].SessionID
	})

	members := make([]game.RoomMemberState, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		members = append(members, game.RoomMemberState{
			PlayerID:  member.PlayerID,
			Ready:     member.Ready,
			Connected: member.Connected,
			IsBot:     member.IsBot,
			TeamID:    string(projection.TeamAssignments[member.MemberID]),
		})
	}

	return game.RoomSnapshot{
		Type:                  game.PacketTypeRoomSnapshot,
		RoomCode:              projection.RoomID,
		RoomState:             string(projection.State),
		CurrentMatchID:        projection.CurrentMatchID,
		Members:               members,
		LocalPlayerID:         projection.LocalPlayerID,
		OwnerID:               projection.OwnerID,
		MaxPlayers:            projection.MaxPlayers,
		PresetID:              string(projection.ModeConfig.PresetID),
		ModeID:                projection.ResolvedModeID,
		ModeLocked:            projection.ModeLocked,
		StartingLives:         projection.ModeConfig.StartingLives,
		InfiniteLives:         projection.ModeConfig.InfiniteLives,
		TargetScore:           projection.ModeConfig.TargetScore,
		TeamStructure:         string(projection.TeamConfig.Structure),
		TeamAssignmentMode:    string(projection.TeamConfig.AssignmentMode),
		TeamCount:             rooms.TeamCountForConfig(projection.TeamConfig),
		TeamAssignmentsLocked: projection.TeamAssignmentsLocked,
		MatchResult:           buildRoomMatchResultSummary(projection),
	}
}

func buildRoomMatchResultSummary(projection rooms.RoomSnapshot) game.RoomMatchResultSummary {
	if !projection.HasResolvedMatch {
		return game.RoomMatchResultSummary{}
	}

	summary := projection.ResolvedSummary
	matchResult := game.RoomMatchResultSummary{
		MatchID: summary.MatchID,
		Mode:    string(summary.Mode),
	}

	if len(summary.Players) == 0 {
		return matchResult
	}

	matchResult.Players = make([]game.RoomPlayerMatchSummary, 0, len(summary.Players))
	for _, player := range summary.Players {
		matchResult.Players = append(matchResult.Players, game.RoomPlayerMatchSummary{
			GamePlayerID: player.GamePlayerID,
			TeamID:       player.TeamID,
			Score:        player.Score,
			ShipDeaths:   player.ShipDeaths,
			Won:          player.Won,
		})
	}

	return matchResult
}

func (session *webSocketSession) EnqueueRoomSnapshot(room *rooms.Room) {
	packet := BuildRoomSnapshot(room, session.sessionID)
	packet.BuildOptions, packet.LoadoutSelection = session.playerBuildPacketStates()
	payload, err := packetcodec.Encode(packet)
	if err != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     room.ID,
				PacketType: game.PacketTypeRoomSnapshot,
			},
			Fields: observability.Fields{
				"error_code":   "room_snapshot_encode_failed",
				"failure_mode": "room_snapshot_encode_failed",
			},
		})
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
