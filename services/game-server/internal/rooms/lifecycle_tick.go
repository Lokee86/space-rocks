package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func TickRoomGameOverLifecycle(room *Room, broadcastRoomSnapshot func(*Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameGameOverDetected,
		Context: observability.Context{
			TraceID: traceIDForLifecycle(room),
			RoomID:  room.ID,
			MatchID: room.CurrentMatchID(),
		},
		Fields: observability.Fields{"reason_code": "simulation_complete"},
	})
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
	traceID := traceIDForLifecycle(room)
	logging.Emit(observability.Request{
		Event: observability.EventNameMatchResultReportStarted,
		Context: observability.Context{
			TraceID: traceID,
			RoomID:  room.ID,
			MatchID: summary.MatchID,
		},
		Fields: observability.Fields{
			"reason_code":  reason,
			"mode":         string(summary.Mode),
			"player_count": len(summary.Players),
		},
	})
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
		logging.Emit(observability.Request{
			Event: observability.EventNameMatchResultReportFailed,
			Context: observability.Context{
				TraceID: traceID,
				RoomID:  room.ID,
				MatchID: summary.MatchID,
			},
			Fields: observability.Fields{
				"reason_code":  reason,
				"failure_mode": "report_failed",
				"player_count": len(summary.Players),
			},
		})
		return false
	}
	if !claimed {
		return false
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameMatchResultReportSucceeded,
		Context: observability.Context{
			TraceID: traceID,
			RoomID:  room.ID,
			MatchID: summary.MatchID,
		},
		Fields: observability.Fields{
			"reason_code":  reason,
			"player_count": len(summary.Players),
		},
	})
	return true
}

func traceIDForLifecycle(room *Room) string {
	return room.CurrentOrCreateMatchTraceID()
}
