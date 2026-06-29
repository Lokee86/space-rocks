package realtime

import (
	"strings"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type ActiveRealtimeResult struct {
	Snapshot          game.GameplayPresentationSnapshot
	SessionState      RealtimeSessionState
	Candidates        []RealtimeLaneCandidate
	PlannedRecords    []ScheduleRecord
	SendPlan          SendPlan
	MetricRecord      packetmetrics.PacketMetricRecord
	MetricSummaries   []packetmetrics.PacketMetricRecord
	EncodedPackets    map[Lane][]byte
	EncodedBytes      map[Lane]int
	EventBatchEventIDs []string
	TotalEncodedBytes int
	Mode              string
}

func BuildActiveRealtimeResultForGame(gameInstance *game.Game, playerID string, state RealtimeSessionState) (ActiveRealtimeResult, error) {
	snapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	return BuildActiveRealtimeResult(snapshot, state), nil
}

func BuildActiveRealtimeResult(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ActiveRealtimeResult {
	prepared := prepareRealtimeSendPlan(snapshot, state)
	encodedPackets := make(map[Lane][]byte, len(prepared.CandidatePlan.Candidates))
	encodedBytes := make(map[Lane]int, len(prepared.CandidatePlan.Candidates))
	for _, candidate := range prepared.CandidatePlan.Candidates {
		encodedPacket, recordedBytes := encodeLanePacket(candidate)
		if recordedBytes > 0 {
			encodedPackets[candidate.Lane] = encodedPacket
			encodedBytes[candidate.Lane] = recordedBytes
		}
	}

	result := ActiveRealtimeResult{
		Snapshot:         snapshot,
		SessionState:     state,
		Candidates:       prepared.CandidatePlan.Candidates,
		PlannedRecords:   prepared.Records,
		SendPlan:         prepared.SendPlan,
		EncodedPackets:   encodedPackets,
		EncodedBytes:     encodedBytes,
		EventBatchEventIDs: activeEventBatchEventIDs(snapshot.PendingEvents),
		Mode:             "active",
	}
	result.MetricRecord = result.SendPlan.Summary.ToPacketMetricRecord("active", LaneWorld, "unspecified", HardCapBytes, "sent", "sent")
	totalEncodedBytes := 0
	for _, recordedBytes := range encodedBytes {
		totalEncodedBytes += recordedBytes
	}
	result.TotalEncodedBytes = totalEncodedBytes
	if len(encodedBytes) > 0 {
		result.MetricRecord.Bytes = encodedBytes[LaneWorld]
	}
	result.MetricSummaries = ActiveLaneMetricRecords(result)
	return result
}

func ActiveLaneMetricRecords(result ActiveRealtimeResult) []packetmetrics.PacketMetricRecord {
	records := make([]packetmetrics.PacketMetricRecord, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		record := result.SendPlan.Summary.ToPacketMetricRecord(string(candidate.Lane), candidate.Lane, "unspecified", HardCapBytes, "sent", "sent")
		record.Bytes = result.EncodedBytes[candidate.Lane]
		records = append(records, record)
	}
	return records
}

func ActiveRealtimeSummaryFields(result ActiveRealtimeResult) []any {
	fields := []any{
		"shadow_vs_sent", "sent",
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

func activeEventBatchEventIDs(pending []game.PendingPresentationEvent) []string {
	if len(pending) == 0 {
		return nil
	}

	ids := make([]string, 0, len(pending))
	for _, event := range pending {
		ids = append(ids, event.EventID)
	}
	return ids
}

func encodeLanePacket(candidate RealtimeLaneCandidate) ([]byte, int) {
	packet := WireLanePacket(candidate)
	if packet == nil {
		return nil, 0
	}
	encoded, err := packetcodec.Encode(packet)
	if err != nil {
		return nil, 0
	}
	return encoded, len(encoded)
}

func laneFamilySummary(records []ScheduleRecord) string {
	if len(records) == 0 {
		return ""
	}

	parts := make([]string, 0, len(records))
	for _, record := range records {
		parts = append(parts, string(record.Lane)+":"+record.PacketFamily)
	}
	return strings.Join(parts, ",")
}
