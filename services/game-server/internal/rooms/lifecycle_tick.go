package rooms

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func TickRoomGameOverLifecycle(room *Room, broadcastRoomSnapshot func(*Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Debug("room game over detected",
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
		reporter = NoopMatchResultReporter{}
	}
	if room.MatchResultReported() {
		return false
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		return false
	}

	if err := reporter.ReportMatchResult(summary); err != nil {
		logging.Rooms.Error("room match result report failed",
			err,
			logging.FieldRoomID, room.ID,
		)
		return false
	}

	room.MarkMatchResultReported()
	return true
}
