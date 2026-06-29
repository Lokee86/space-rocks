package realtime

import (
	"strings"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type ShadowRealtimeResult struct {
	Snapshot        game.GameplayPresentationSnapshot
	SessionState    RealtimeSessionState
	Candidates      []RealtimeLaneCandidate
	PlannedRecords   []ScheduleRecord
	SendPlan        SendPlan
	MetricRecord    packetmetrics.PacketMetricRecord
	EncodedBytes    map[Lane]int
	TotalEncodedBytes int
}

func BuildShadowRealtimeResult(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ShadowRealtimeResult {
	candidatePlan := AssembleRealtimeLaneCandidates(snapshot, state)

	records := make([]ScheduleRecord, 0, len(candidatePlan.Candidates))
	encodedBytes := make(map[Lane]int, len(candidatePlan.Candidates))
	for _, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(candidate))
		recordedBytes := encodeShadowLanePacket(candidate)
		if recordedBytes > 0 {
			encodedBytes[candidate.Lane] = recordedBytes
		}
	}

	sendPlan := SelectSendPlan(records)
	metricRecord := sendPlan.Summary.ToPacketMetricRecord("shadow", LaneWorld, "unspecified", HardCapBytes, "shadow", "shadow")
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
		Candidates:        candidatePlan.Candidates,
		PlannedRecords:    records,
		SendPlan:          sendPlan,
		MetricRecord:      metricRecord,
		EncodedBytes:      encodedBytes,
		TotalEncodedBytes: totalEncodedBytes,
	}
}

func ShadowLaneMetricRecords(result ShadowRealtimeResult) []packetmetrics.PacketMetricRecord {
	records := make([]packetmetrics.PacketMetricRecord, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		record := result.SendPlan.Summary.ToPacketMetricRecord(string(candidate.Lane), candidate.Lane, "unspecified", HardCapBytes, "shadow", "shadow")
		record.Bytes = result.EncodedBytes[candidate.Lane]
		records = append(records, record)
	}
	return records
}

func ShadowRealtimeSummaryFields(result ShadowRealtimeResult) []any {
	fields := []any{
		"shadow_vs_sent", "shadow",
		"lane_packet_families", shadowLaneFamilySummary(result.PlannedRecords),
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

func shadowLaneFamilySummary(records []ScheduleRecord) string {
	if len(records) == 0 {
		return ""
	}

	parts := make([]string, 0, len(records))
	for _, record := range records {
		parts = append(parts, string(record.Lane)+":"+record.PacketFamily)
	}
	return strings.Join(parts, ",")
}

func encodeShadowLanePacket(candidate RealtimeLaneCandidate) int {
	var packet any
	switch candidate.Kind {
	case RealtimeLaneCandidateKindFull:
		packet = candidate.Full
	case RealtimeLaneCandidateKindDelta:
		packet = candidate.Delta
	case RealtimeLaneCandidateKindEventBatch:
		packet = candidate.Full
	}
	if packet == nil {
		return 0
	}
	encoded, err := packetcodec.Encode(packet)
	if err != nil {
		return 0
	}
	return len(encoded)
}

func scheduleRecordForCandidate(candidate RealtimeLaneCandidate) ScheduleRecord {
	record := ScheduleRecord{Lane: candidate.Lane, Priority: PriorityMedium, DeliveryClass: DeliveryClassDeferrable}
	switch candidate.Lane {
	case LaneWorld:
		record.PacketFamily = PacketFamilyWorldFull
	case LaneOverlay:
		record.PacketFamily = PacketFamilyOverlayFull
	case LaneSession:
		record.PacketFamily = PacketFamilySessionFull
	case LaneEvent:
		record.PacketFamily = PacketFamilyEventBatch
	}
	switch candidate.Kind {
	case RealtimeLaneCandidateKindFull:
		record.RecordKind = "full"
		record.DeliveryClass = DeliveryClassRequired
		record.Priority = PriorityCritical
		record.RequiredCount = 1
	case RealtimeLaneCandidateKindDelta:
		record.RecordKind = "delta"
		record.DeliveryClass = DeliveryClassHotSupersedable
		record.Priority = PriorityHigh
		record.SupersessionKey = string(candidate.Lane)
	case RealtimeLaneCandidateKindEventBatch:
		record.RecordKind = "event_batch"
		record.DeliveryClass = DeliveryClassEventOnce
		record.Priority = PriorityHigh
	}
	return record
}
