package realtime

import "github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"

type LaneSendResult struct {
	PacketsBuilt             int
	PacketsWrittenOrEnqueued int
	BytesWrittenOrEnqueued   int
	EventIDsWrittenOrEnqueued int
	MetricSummaries          []packetmetrics.PacketMetricRecord
	Err                      error
}
