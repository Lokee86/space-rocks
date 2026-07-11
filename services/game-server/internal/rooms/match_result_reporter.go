package rooms

import "github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"

type MatchResultReporter interface {
	ReportMatchResult(summary playerdata.MatchResultSummary) error
}

type NoopMatchResultReporter struct{}

func (NoopMatchResultReporter) ReportMatchResult(summary playerdata.MatchResultSummary) error {
	return nil
}
