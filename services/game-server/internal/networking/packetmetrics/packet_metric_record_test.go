package packetmetrics

import "testing"

func TestPacketMetricRecordCloneReturnsCopy(t *testing.T) {
	record := PacketMetricRecord{
		PacketFamily:    "realtime",
		Lane:            "overlay",
		Bytes:           128,
		RecordCount:     3,
		CreateCount:     1,
		UpdateCount:     2,
		DeleteCount:     0,
		PriorityBand:    "high",
		DeferredCount:   4,
		SupersededCount: 5,
		RequiredCount:   6,
		BudgetTarget:    256,
		BudgetStatus:    "within_budget",
		SendReason:      "delta",
		ChunkDecision:   "chunked",
		ResyncDecision:  "not_required",
	}

	clone := record.Clone()

	if clone != record {
		t.Fatalf("expected clone to match record, got %#v want %#v", clone, record)
	}
}

func TestPacketMetricRecordCapturesLaneAndPacketFamily(t *testing.T) {
	record := PacketMetricRecord{PacketFamily: "world_delta", Lane: "world"}
	if record.PacketFamily != "world_delta" || record.Lane != "world" {
		t.Fatalf("record = %#v, want lane and packet family preserved", record)
	}
}

