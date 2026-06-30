package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
)

type ShadowRealtimeResult struct {
	Snapshot          game.GameplayPresentationSnapshot
	SessionState      RealtimeSessionState
	Candidates        []RealtimeLaneCandidate
	PlannedRecords    []ScheduleRecord
	SendPlan          SendPlan
	MetricRecord      packetmetrics.PacketMetricRecord
	EncodedBytes      map[Lane]int
	TotalEncodedBytes int
}

func BuildShadowRealtimeResult(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ShadowRealtimeResult {
	prepared := prepareRealtimeSendPlan(snapshot, state)
	encodedBytes := make(map[Lane]int, len(prepared.CandidatePlan.Candidates))
	for _, candidate := range prepared.CandidatePlan.Candidates {
		_, recordedBytes := encodeLanePacket(candidate)
		if recordedBytes > 0 {
			encodedBytes[candidate.Lane] = recordedBytes
		}
	}

	metricRecord := prepared.SendPlan.Summary.ToPacketMetricRecord("shadow", LaneWorld, "unspecified", HardCapBytes, "shadow")
	totalEncodedBytes := 0
	for _, recordedBytes := range encodedBytes {
		totalEncodedBytes += recordedBytes
	}
	if len(encodedBytes) > 0 {
		metricRecord.Bytes = encodedBytes[LaneWorld]
	}

	return ShadowRealtimeResult{
		Snapshot:          snapshot,
		SessionState:      state,
		Candidates:        prepared.CandidatePlan.Candidates,
		PlannedRecords:    prepared.Records,
		SendPlan:          prepared.SendPlan,
		MetricRecord:      metricRecord,
		EncodedBytes:      encodedBytes,
		TotalEncodedBytes: totalEncodedBytes,
	}
}

func ShadowLaneMetricRecords(result ShadowRealtimeResult) []packetmetrics.PacketMetricRecord {
	records := make([]packetmetrics.PacketMetricRecord, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		record := result.SendPlan.Summary.ToPacketMetricRecord(string(candidate.Lane), candidate.Lane, "unspecified", HardCapBytes, "shadow")
		record.Bytes = result.EncodedBytes[candidate.Lane]
		records = append(records, record)
	}
	return records
}

func ShadowRealtimeSummaryFields(result ShadowRealtimeResult) []any {
	fields := []any{
		"lane_packet_families", laneFamilySummary(result.PlannedRecords),
		"encoded_bytes", result.TotalEncodedBytes,
		"included_count", len(result.SendPlan.Included),
		"deferred_count", len(result.SendPlan.Deferred),
		"superseded_count", result.SendPlan.Summary.SupersededCount,
	}
	if pendingEvents := len(result.Snapshot.PendingEvents); pendingEvents > 0 {
		fields = append(fields, "event_batch_count", pendingEvents)
	}
	return fields
}
