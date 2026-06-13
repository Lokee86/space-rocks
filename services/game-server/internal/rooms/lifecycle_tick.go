package rooms

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func TickRoomGameOverLifecycle(room *Room, broadcastRoomSnapshot func(*Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Info("room game over detected",
		logging.FieldRoomID, room.ID,
	)
	broadcastRoomSnapshot(room)
	return true
}

func ReportResolvedMatchResultOnce(room *Room, reporter MatchResultReporter) bool {
	if room == nil {
		return false
	}
	if reporter == nil {
		logging.Rooms.Warn("match result reporter missing; using noop reporter",
			logging.FieldRoomID, room.ID,
		)
		reporter = NoopMatchResultReporter{}
	}
	if room.MatchResultReported() {
		logging.Rooms.Info("match result report skipped: already reported",
			logging.FieldRoomID, room.ID,
		)
		return false
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		logging.Rooms.Warn("match result report skipped: missing resolved summary",
			logging.FieldRoomID, room.ID,
		)
		return false
	}

	logging.Rooms.Info("match result report started",
		logging.FieldRoomID, room.ID,
		"match_id", summary.MatchID,
		"mode", summary.Mode,
		"player_count", len(summary.Players),
	)
	if err := reporter.ReportMatchResult(summary); err != nil {
		logging.Rooms.Error("room match result report failed",
			err,
			logging.FieldRoomID, room.ID,
			"match_id", summary.MatchID,
			"player_count", len(summary.Players),
		)
		return false
	}

	room.MarkMatchResultReported()
	logging.Rooms.Info("match result report succeeded",
		logging.FieldRoomID, room.ID,
		"match_id", summary.MatchID,
		"player_count", len(summary.Players),
	)
	return true
}
