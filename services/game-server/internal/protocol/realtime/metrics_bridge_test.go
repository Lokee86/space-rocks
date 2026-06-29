package realtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
)

func TestSendPlanSummaryToPacketMetricRecordMapsCountsAndBudgets(t *testing.T) {
	summary := SendPlanSummary{
		IncludedCount:   7,
		DeferredCount:   3,
		SupersededCount: 2,
		RequiredCount:   5,
		CreateCount:     4,
		UpdateCount:     6,
		DeleteCount:     1,
	}

	record := summary.ToPacketMetricRecord(PacketFamilyWorldDelta, LaneWorld, "high", 500, "within_budget", "shadow")

	if record.PacketFamily != PacketFamilyWorldDelta || record.Lane != string(LaneWorld) {
		t.Fatalf("record = %#v, want packet family and lane preserved", record)
	}
	if record.RecordCount != 7 || record.CreateCount != 4 || record.UpdateCount != 6 || record.DeleteCount != 1 {
		t.Fatalf("record counts = %#v, want send-plan summary copied", record)
	}
	if record.DeferredCount != 3 || record.SupersededCount != 2 || record.RequiredCount != 5 {
		t.Fatalf("record deferred/superseded/required = %#v, want summary copied", record)
	}
	if record.BudgetTarget != 500 || record.BudgetStatus != "within_budget" {
		t.Fatalf("budget fields = %#v, want preserved", record)
	}
	if record.ShadowVsSent != "shadow" {
		t.Fatalf("shadow_vs_sent = %q, want shadow", record.ShadowVsSent)
	}

	sent := summary.ToPacketMetricRecord(PacketFamilyWorldDelta, LaneWorld, "high", 500, "within_budget", "sent")
	if sent.ShadowVsSent != "sent" {
		t.Fatalf("shadow_vs_sent = %q, want sent", sent.ShadowVsSent)
	}

	_ = packetmetrics.PacketMetricRecord{}
}
