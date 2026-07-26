package realtime

import "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetmetrics"

func (summary SendPlanSummary) ToPacketMetricRecord(packetFamily string, lane Lane) packetmetrics.PacketMetricRecord {
	return packetmetrics.PacketMetricRecord{
		PacketFamily: packetFamily,
		Lane:         string(lane),
	}
}
