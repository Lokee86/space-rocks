package realtime

import "github.com/Lokee86/space-rocks/services/game-server/internal/networking/packetmetrics"

type LaneSendResult struct {
	PacketsBuilt              int
	PacketsWrittenOrEnqueued  int
	BytesWrittenOrEnqueued    int
	EventIDsWrittenOrEnqueued int
	MetricSummaries           []packetmetrics.PacketMetricRecord
	Err                       error
}
