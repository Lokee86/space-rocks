package realtime

import (
	"fmt"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/packetmetrics"
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

func BuildShadowRealtimeResult(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) (ShadowRealtimeResult, error) {
	prepared, err := prepareRealtimeSendPlan(snapshot, state)
	if err != nil {
		return ShadowRealtimeResult{}, fmt.Errorf("expand shadow candidates: %w", err)
	}
	encodedBytes := make(map[Lane]int, len(prepared.CandidatePlan.Candidates))
	for _, candidate := range prepared.CandidatePlan.Candidates {
		_, recordedBytes, err := encodeLanePacket(candidate)
		if err != nil {
			return ShadowRealtimeResult{}, fmt.Errorf("encode shadow candidate lane=%q family=%q: %w", candidate.Lane(), candidate.PacketFamily(), err)
		}
		if recordedBytes > 0 {
			encodedBytes[candidate.Lane()] = recordedBytes
		}
	}

	metricRecord := prepared.SendPlan.Summary.ToPacketMetricRecord("shadow", LaneWorld)
	totalEncodedBytes := 0
	for _, recordedBytes := range encodedBytes {
		totalEncodedBytes += recordedBytes
	}
	if len(encodedBytes) > 0 {
		metricRecord.Bytes = encodedBytes[LaneWorld]
	}

	return ShadowRealtimeResult{
		Snapshot:          snapshot,
		SessionState:      prepared.SessionState,
		Candidates:        prepared.CandidatePlan.Candidates,
		PlannedRecords:    prepared.Records,
		SendPlan:          prepared.SendPlan,
		MetricRecord:      metricRecord,
		EncodedBytes:      encodedBytes,
		TotalEncodedBytes: totalEncodedBytes,
	}, nil
}

func ShadowLaneMetricRecords(result ShadowRealtimeResult) []packetmetrics.PacketMetricRecord {
	records := make([]packetmetrics.PacketMetricRecord, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		record := result.SendPlan.Summary.ToPacketMetricRecord(candidate.PacketFamily(), candidate.Lane())
		diagnostics := CandidateWriteDiagnosticsFor(candidate, result.SessionState, result.EncodedBytes[candidate.Lane()])
		record.Bytes = result.EncodedBytes[candidate.Lane()]
		record.Channel = diagnostics.Channel
		record.EncodedBytes = diagnostics.EncodedBytes
		record.WorldHotCount = diagnostics.WorldHotCount
		record.ShipHotCount = diagnostics.ShipHotCount
		record.AsteroidHotCount = diagnostics.AsteroidHotCount
		record.BulletHotCount = diagnostics.BulletHotCount
		record.ShipOffloadedCount = diagnostics.ShipOffloadedCount
		record.AsteroidOffloadedCount = diagnostics.AsteroidOffloadedCount
		record.BulletOffloadedCount = diagnostics.BulletOffloadedCount
		record.ShipMode = string(diagnostics.ShipMode)
		record.AsteroidMode = string(diagnostics.AsteroidMode)
		record.BulletMode = string(diagnostics.BulletMode)
		record.Cadence = diagnostics.Cadence
		record.PacketOverTarget = diagnostics.PacketOverTarget
		record.PacketOverHardCap = diagnostics.PacketOverHardCap
		records = append(records, record)
	}
	return records
}

func ShadowRealtimeSummaryFields(result ShadowRealtimeResult) []any {
	fields := []any{
		"lane_packet_families", laneFamilySummary(result.PlannedRecords),
		"encoded_bytes", result.TotalEncodedBytes,
	}
	if pendingEvents := len(result.Snapshot.PendingEvents); pendingEvents > 0 {
		fields = append(fields, "event_batch_count", pendingEvents)
	}
	return fields
}
