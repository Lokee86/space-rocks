package realtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetmetrics"
)

func TestSendPlanSummaryToPacketMetricRecordMapsPacketFamilyAndLane(t *testing.T) {
	summary := SendPlanSummary{}

	record := summary.ToPacketMetricRecord(PacketFamilyWorldDelta, LaneWorld)

	if record.PacketFamily != PacketFamilyWorldDelta || record.Lane != string(LaneWorld) {
		t.Fatalf("record = %#v, want packet family and lane preserved", record)
	}

	_ = packetmetrics.PacketMetricRecord{}
}
