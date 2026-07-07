package realtime

import (
	"strings"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type EncodedRealtimeLanePacket struct {
	Candidate    RealtimeLaneCandidate
	Encoded      []byte
	EncodedBytes int
}

type ActiveRealtimeResult struct {
	Snapshot           game.GameplayPresentationSnapshot
	SessionState       RealtimeSessionState
	Candidates         []RealtimeLaneCandidate
	SelectedCandidates []RealtimeLaneCandidate
	PlannedRecords     []ScheduleRecord
	SendPlan           SendPlan
	MetricRecord       packetmetrics.PacketMetricRecord
	MetricSummaries    []packetmetrics.PacketMetricRecord
	EncodedLanePackets []EncodedRealtimeLanePacket
	EncodedPackets     map[Lane][]byte
	EncodedBytes       map[Lane]int
	EventBatchEventIDs []string
	TotalEncodedBytes  int
	Mode               string
}

func BuildActiveRealtimeResultForGame(gameInstance *game.Game, playerID string, state RealtimeSessionState) (ActiveRealtimeResult, error) {
	snapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	logActivePendingPresentationEvents(playerID, snapshot)
	return BuildActiveRealtimeResult(snapshot, state), nil
}

func BuildActiveRealtimeResult(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ActiveRealtimeResult {
	preparedState := state
	candidatePlan := assembleRealtimeLaneCandidates(snapshot, state, &preparedState)
	candidatePlan.Candidates = ExpandHotLaneCandidateChunks(candidatePlan.Candidates)

	records := make([]ScheduleRecord, 0, len(candidatePlan.Candidates))
	for i, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(i, candidate))
	}
	sendPlan := SelectSendPlan(records)
	selectedCandidates := IncludedRealtimeLaneCandidates(candidatePlan.Candidates, sendPlan.Included)
	encodedPackets := make(map[Lane][]byte, len(selectedCandidates))
	encodedBytes := make(map[Lane]int, len(selectedCandidates))
	encodedLanePackets := make([]EncodedRealtimeLanePacket, 0, len(selectedCandidates))
	for _, candidate := range selectedCandidates {
		encodedPacket, recordedBytes := encodeLanePacket(candidate)
		if recordedBytes > 0 {
			encodedLanePackets = append(encodedLanePackets, EncodedRealtimeLanePacket{
				Candidate:    candidate,
				Encoded:      encodedPacket,
				EncodedBytes: recordedBytes,
			})
			if _, exists := encodedPackets[candidate.Lane]; !exists {
				encodedPackets[candidate.Lane] = encodedPacket
			}
			encodedBytes[candidate.Lane] += recordedBytes
		}
	}

	result := ActiveRealtimeResult{
		Snapshot:           snapshot,
		SessionState:       preparedState,
		Candidates:         candidatePlan.Candidates,
		SelectedCandidates: selectedCandidates,
		PlannedRecords:     records,
		SendPlan:           sendPlan,
		MetricRecord:       packetmetrics.PacketMetricRecord{},
		MetricSummaries:    nil,
		EncodedLanePackets: encodedLanePackets,
		EncodedPackets:     encodedPackets,
		EncodedBytes:       encodedBytes,
		EventBatchEventIDs: activeEventBatchEventIDs(snapshot.PendingEvents),
		Mode:               "active",
	}
	result.MetricRecord = result.SendPlan.Summary.ToPacketMetricRecord("active", LaneWorld)
	totalEncodedBytes := 0
	for _, recorded := range encodedLanePackets {
		totalEncodedBytes += recorded.EncodedBytes
	}
	result.TotalEncodedBytes = totalEncodedBytes
	if len(encodedBytes) > 0 {
		result.MetricRecord.Bytes = encodedBytes[LaneWorld]
	}
	result.MetricSummaries = ActiveLaneMetricRecords(result)
	return result
}

func IncludedRealtimeLaneCandidates(candidates []RealtimeLaneCandidate, included []ScheduleRecord) []RealtimeLaneCandidate {
	if len(candidates) == 0 || len(included) == 0 {
		return nil
	}

	selected := make([]RealtimeLaneCandidate, 0, len(included))
	seen := make(map[int]struct{}, len(included))
	for _, record := range included {
		index := record.CandidateIndex
		if index < 0 || index >= len(candidates) {
			continue
		}
		if _, ok := seen[index]; ok {
			continue
		}
		seen[index] = struct{}{}
		selected = append(selected, candidates[index])
	}

	if len(selected) == 0 {
		return nil
	}
	return selected
}

func encodedMetricPackets(result ActiveRealtimeResult) []EncodedRealtimeLanePacket {
	if len(result.EncodedLanePackets) > 0 {
		return result.EncodedLanePackets
	}

	if len(result.SelectedCandidates) == 0 || len(result.EncodedBytes) == 0 {
		return nil
	}

	packets := make([]EncodedRealtimeLanePacket, 0, len(result.SelectedCandidates))
	for _, candidate := range result.SelectedCandidates {
		encodedBytes := result.EncodedBytes[candidate.Lane]
		if encodedBytes <= 0 {
			continue
		}
		packets = append(packets, EncodedRealtimeLanePacket{
			Candidate:    candidate,
			EncodedBytes: encodedBytes,
		})
	}
	return packets
}

func ActiveLaneMetricRecords(result ActiveRealtimeResult) []packetmetrics.PacketMetricRecord {
	encodedPackets := encodedMetricPackets(result)
	records := make([]packetmetrics.PacketMetricRecord, 0, len(encodedPackets))
	for _, encoded := range encodedPackets {
		candidate := encoded.Candidate
		record := result.SendPlan.Summary.ToPacketMetricRecord(packetFamilyForCandidate(candidate), candidate.Lane)
		diagnostics := CandidateWriteDiagnosticsFor(candidate, result.SessionState, encoded.EncodedBytes)
		record.Bytes = encoded.EncodedBytes
		record.Channel = diagnostics.Channel
		record.EncodedBytes = diagnostics.EncodedBytes
		record.WorldHotCount = diagnostics.WorldHotCount
		record.AsteroidHotCount = diagnostics.AsteroidHotCount
		record.BulletHotCount = diagnostics.BulletHotCount
		record.AsteroidOffloadedCount = diagnostics.AsteroidOffloadedCount
		record.BulletOffloadedCount = diagnostics.BulletOffloadedCount
		record.AsteroidMode = string(diagnostics.AsteroidMode)
		record.BulletMode = string(diagnostics.BulletMode)
		record.Cadence = diagnostics.Cadence
		record.PacketOverTarget = diagnostics.PacketOverTarget
		record.PacketOverHardCap = diagnostics.PacketOverHardCap
		records = append(records, record)
	}
	return records
}

func ActiveRealtimeSummaryFields(result ActiveRealtimeResult) []any {
	fields := []any{
		"lane_packet_families", laneFamilySummary(result.PlannedRecords),
		"encoded_bytes", result.TotalEncodedBytes,
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

func logActivePendingPresentationEvents(playerID string, snapshot game.GameplayPresentationSnapshot) {
	if len(snapshot.PendingEvents) == 0 {
		return
	}

	eventTypes := make([]string, 0, len(snapshot.PendingEvents))
	eventIDs := make([]string, 0, len(snapshot.PendingEvents))
	for _, event := range snapshot.PendingEvents {
		eventTypes = append(eventTypes, event.Event.Type)
		eventIDs = append(eventIDs, event.EventID)
		if event.Event.Type == game.PacketTypeShipDeath {
			logging.Network.Debug("ship death pending in active snapshot",
				logging.FieldPlayerID, playerID,
				"event_id", event.EventID,
				"event_type", event.Event.Type,
				"event_player_id", event.Event.PlayerID,
				"lives", event.Event.Lives,
				"respawn_delay", event.Event.RespawnDelay,
			)
		}
	}
	if playerID == "" {
		playerID = "unknown"
	}
	logging.Network.Debug("pending presentation events in active snapshot",
		logging.FieldPlayerID, playerID,
		"pending_event_count", len(snapshot.PendingEvents),
		"event_types", strings.Join(eventTypes, ","),
		"event_ids", strings.Join(eventIDs, ","),
	)
}

func encodeLanePacketUnchecked(candidate RealtimeLaneCandidate) ([]byte, int) {
	packet := WireLanePacket(candidate)
	if packet == nil {
		return nil, 0
	}
	if candidate.Lane == LaneWorld || candidate.Lane == LaneSession || candidate.Lane == LaneOverlay || candidate.Lane == LaneEvent || candidate.Lane == LaneAsteroids || candidate.Lane == LaneBullets {
		packet = CompactWirePacket(packet)
	}
	encoded, err := packetcodec.Encode(packet)
	if err != nil {
		return nil, 0
	}
	return encoded, len(encoded)
}

func encodeLanePacket(candidate RealtimeLaneCandidate) ([]byte, int) {
	encoded, recordedBytes := encodeLanePacketUnchecked(candidate)
	if recordedBytes <= 0 {
		return nil, 0
	}
	if candidate.Lane == LaneAsteroids || candidate.Lane == LaneBullets {
		if !hotPacketSendAllowed(recordedBytes) {
			return nil, 0
		}
	}
	return encoded, recordedBytes
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


