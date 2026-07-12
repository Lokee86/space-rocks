package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func TickRoomGameOverLifecycle(room *Room, broadcastRoomSnapshot func(*Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}
	logging.Rooms.Info("room game over detected", logging.FieldRoomID, room.ID)
	broadcastRoomSnapshot(room)
	return true
}

func ReportResolvedMatchResultOnce(room *Room, reporter MatchResultReporter) bool {
	return ReportResolvedMatchResultOnceForReason(room, reporter, "game_over")
}

func ReportResolvedMatchResultOnceForReason(room *Room, reporter MatchResultReporter, reason string) bool {
	if room == nil {
		return false
	}
	if reporter == nil {
		reporter = NoopMatchResultReporter{}
	}
	room.mu.Lock()
	if room.match.matchResultReported || room.match.matchResultReporting {
		room.mu.Unlock()
		return false
	}
	summary, ok := room.match.ResolvedSummary()
	if !ok {
		room.mu.Unlock()
		return false
	}
	roomMatchID := room.match.CurrentMatchID()
	claimedSummaryID := summary.MatchID
	claimedSummary := summary
	claimedSummary.Players = append([]playerdata.PlayerMatchSummary(nil), summary.Players...)
	room.match.matchResultReporting = true
	room.mu.Unlock()

	logging.Rooms.Info("match result report started", logging.FieldRoomID, room.ID, "reason", reason, "match_id", summary.MatchID, "mode", summary.Mode, "player_count", len(summary.Players))
	err := reporter.ReportMatchResult(claimedSummary)

	room.mu.Lock()
	currentSummary, hasSummary := room.match.ResolvedSummary()
	claimed := room.match.CurrentMatchID() == roomMatchID && room.match.matchResultReporting && hasSummary && currentSummary.MatchID == claimedSummaryID
	if claimed {
		room.match.matchResultReporting = false
		if err == nil {
			room.match.matchResultReported = true
		}
	}
	room.mu.Unlock()
	if err != nil {
		logging.Rooms.Error("room match result report failed", err, logging.FieldRoomID, room.ID, "reason", reason, "match_id", summary.MatchID, "player_count", len(summary.Players))
		return false
	}
	if !claimed {
		return false
	}
	logging.Rooms.Info("match result report succeeded", logging.FieldRoomID, room.ID, "reason", reason, "match_id", summary.MatchID, "player_count", len(summary.Players))
	return true
}
