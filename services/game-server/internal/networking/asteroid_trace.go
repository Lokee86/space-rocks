package networking

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

const asteroidTraceEnv = "SPACE_ROCKS_ASTEROID_TRACE"

func asteroidTraceEnabled() bool {
	configured := strings.ToLower(strings.TrimSpace(os.Getenv(asteroidTraceEnv)))
	if configured != "" {
		return configured == "1" || configured == "true" || configured == "yes" || configured == "on"
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ENVIRONMENT")), "development")
}

func traceAsteroidPacketWrite(
	session *webSocketSession,
	context SessionContext,
	matchID string,
	candidate realtime.RealtimeLaneCandidate,
	encodedBytes int,
	bufferedBytes uint64,
) {
	if !asteroidTraceEnabled() {
		return
	}
	lane := candidate.Lane()
	if lane != realtime.LaneAsteroids && lane != realtime.LaneAsteroidsLifecycle {
		return
	}
	metadata, ok := candidate.Metadata()
	if !ok {
		return
	}

	creates, updates, deletes, createIDs, updateIDs, deleteIDs := asteroidCandidateCounts(candidate)
	if lane == realtime.LaneAsteroids && metadata.Sequence%30 != 0 {
		return
	}
	fields := observability.Fields{
		"lane":           string(lane),
		"sequence":       metadata.Sequence,
		"baseline_id":    metadata.BaselineID,
		"chunk_index":    metadata.ChunkIndex,
		"chunk_count":    metadata.ChunkCount,
		"is_final_chunk": metadata.IsFinalChunk,
		"create_count":   creates,
		"update_count":   updates,
		"delete_count":   deletes,
		"create_ids":     jsonString(createIDs),
		"update_ids":     jsonString(updateIDs),
		"delete_ids":     jsonString(deleteIDs),
		"encoded_bytes":  encodedBytes,
		"buffered_bytes": bufferedBytes,
	}
	level := observability.LevelDebug
	if lane == realtime.LaneAsteroidsLifecycle {
		level = observability.LevelInfo
	}
	logging.EmitLegacy(observability.LegacyRequest{
		Level:       level,
		Category:    "realtime_asteroid_trace",
		LegacyEvent: "asteroid_packet_sent",
		Context: observability.Context{
			TraceID:    session.connectionTraceID,
			SessionID:  session.sessionID,
			RoomID:     context.RoomID,
			PlayerID:   context.GamePlayerID,
			MatchID:    matchID,
			PacketType: candidate.PacketFamily(),
		},
		Fields: fields,
	})
}

func asteroidCandidateCounts(candidate realtime.RealtimeLaneCandidate) (int, int, int, []string, []string, []string) {
	packet, ok := candidate.Payload.(realtime.AsteroidWireDeltaPacket)
	if !ok {
		if pointer, pointerOK := candidate.Payload.(*realtime.AsteroidWireDeltaPacket); pointerOK && pointer != nil {
			packet = *pointer
			ok = true
		}
	}
	if !ok {
		return 0, 0, 0, nil, nil, nil
	}
	createIDs := make([]string, 0, min(len(packet.AsteroidCreates), 12))
	for _, record := range packet.AsteroidCreates {
		if len(createIDs) >= 12 {
			break
		}
		createIDs = append(createIDs, record.ID)
	}
	updateIDs := make([]string, 0, min(len(packet.AsteroidUpdates), 12))
	for _, record := range packet.AsteroidUpdates {
		if len(updateIDs) >= 12 {
			break
		}
		if id, ok := record["id"].(string); ok {
			updateIDs = append(updateIDs, id)
		}
	}
	deleteIDs := packet.AsteroidDeletes
	if len(deleteIDs) > 12 {
		deleteIDs = deleteIDs[:12]
	}
	return len(packet.AsteroidCreates), len(packet.AsteroidUpdates), len(packet.AsteroidDeletes), createIDs, updateIDs, deleteIDs
}

func jsonString(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
