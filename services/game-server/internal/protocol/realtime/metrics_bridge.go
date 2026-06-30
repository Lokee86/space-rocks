package realtime

import "github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"

func (summary SendPlanSummary) ToPacketMetricRecord(packetFamily string, lane Lane, priorityBand string, budgetTarget int, budgetStatus string) packetmetrics.PacketMetricRecord {
	return packetmetrics.PacketMetricRecord{
		PacketFamily:    packetFamily,
		Lane:            string(lane),
		RecordCount:     summary.IncludedCount,
		CreateCount:     summary.CreateCount,
		UpdateCount:     summary.UpdateCount,
		DeleteCount:     summary.DeleteCount,
		PriorityBand:    priorityBand,
		DeferredCount:   summary.DeferredCount,
		SupersededCount: summary.SupersededCount,
		RequiredCount:   summary.RequiredCount,
		BudgetTarget:    budgetTarget,
		BudgetStatus:    budgetStatus,
	}
}
