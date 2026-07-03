package realtime

import "github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"

func (summary SendPlanSummary) ToPacketMetricRecord(packetFamily string, lane Lane) packetmetrics.PacketMetricRecord {
	return packetmetrics.PacketMetricRecord{
		PacketFamily: packetFamily,
		Lane:         string(lane),
	}
}
