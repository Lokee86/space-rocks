package realtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
)

func TestActiveLaneMetricsAreMarkedSentNotShadow(t *testing.T) {
	summary := SendPlanSummary{
		IncludedCount:   2,
		DeferredCount:   1,
		SupersededCount: 0,
		RequiredCount:   2,
		CreateCount:     1,
		UpdateCount:     1,
		DeleteCount:     0,
	}

	record := summary.ToPacketMetricRecord(PacketFamilyWorldFull, LaneWorld, "high", HardCapBytes, "within_budget", "sent")
	if record.ShadowVsSent != "sent" {
		t.Fatalf("shadow_vs_sent = %q, want sent", record.ShadowVsSent)
	}

	result := ActiveRealtimeResult{
		Candidates: []RealtimeLaneCandidate{{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull}},
		SendPlan: SendPlan{Summary: summary},
		EncodedBytes: map[Lane]int{LaneWorld: 128},
		Mode: "active",
	}

	records := ActiveLaneMetricRecords(result)
	if len(records) != 1 {
		t.Fatalf("expected 1 metric record, got %d", len(records))
	}
	if records[0].ShadowVsSent != "sent" {
		t.Fatalf("active metric shadow_vs_sent = %q, want sent", records[0].ShadowVsSent)
	}
	if records[0].Bytes != 128 {
		t.Fatalf("active metric bytes = %d, want 128", records[0].Bytes)
	}

	_ = packetmetrics.PacketMetricRecord{}
}
