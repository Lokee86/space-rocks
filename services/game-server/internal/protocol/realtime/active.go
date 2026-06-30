package realtime

import (
	"strings"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type ActiveRealtimeResult struct {
	Snapshot           game.GameplayPresentationSnapshot
	SessionState       RealtimeSessionState
	Candidates         []RealtimeLaneCandidate
	SelectedCandidates []RealtimeLaneCandidate
	PlannedRecords     []ScheduleRecord
	SendPlan           SendPlan
	MetricRecord       packetmetrics.PacketMetricRecord
	MetricSummaries    []packetmetrics.PacketMetricRecord
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
	prepared := prepareRealtimeSendPlan(snapshot, state)
	selectedCandidates := IncludedRealtimeLaneCandidates(prepared.CandidatePlan.Candidates, prepared.SendPlan.Included)
	encodedPackets := make(map[Lane][]byte, len(selectedCandidates))
	encodedBytes := make(map[Lane]int, len(selectedCandidates))
	for _, candidate := range selectedCandidates {
		encodedPacket, recordedBytes := encodeLanePacket(candidate)
		if recordedBytes > 0 {
			encodedPackets[candidate.Lane] = encodedPacket
			encodedBytes[candidate.Lane] = recordedBytes
		}
	}

	result := ActiveRealtimeResult{
		Snapshot:           snapshot,
		SessionState:       state,
		Candidates:         prepared.CandidatePlan.Candidates,
		SelectedCandidates: selectedCandidates,
		PlannedRecords:     prepared.Records,
		SendPlan:           prepared.SendPlan,
		EncodedPackets:     encodedPackets,
		EncodedBytes:       encodedBytes,
		EventBatchEventIDs: activeEventBatchEventIDs(snapshot.PendingEvents),
		Mode:               "active",
	}
	result.MetricRecord = result.SendPlan.Summary.ToPacketMetricRecord("active", LaneWorld, "unspecified", HardCapBytes, "sent")
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

func ActiveLaneMetricRecords(result ActiveRealtimeResult) []packetmetrics.PacketMetricRecord {
	records := make([]packetmetrics.PacketMetricRecord, 0, len(result.SelectedCandidates))
	for _, candidate := range result.SelectedCandidates {
		record := result.SendPlan.Summary.ToPacketMetricRecord(string(candidate.Lane), candidate.Lane, "unspecified", HardCapBytes, "sent")
		record.Bytes = result.EncodedBytes[candidate.Lane]
		records = append(records, record)
	}
	return records
}

func ActiveRealtimeSummaryFields(result ActiveRealtimeResult) []any {
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
