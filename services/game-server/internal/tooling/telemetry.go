package tooling

import (
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	networkingtooling "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
)

type livePacketMetrics struct {
	PacketCount         uint64
	EncodedBytesTotal   uint64
	LastEncodedBytes    int
	MaximumEncodedBytes int
	LastPacketFamily    string
}

type liveTelemetryState struct {
	mu       sync.Mutex
	roomID   string
	sequence int
	packets  map[string]livePacketMetrics
}

func newLiveTelemetryState(roomID string) *liveTelemetryState {
	return &liveTelemetryState{roomID: roomID, packets: make(map[string]livePacketMetrics)}
}

func (controller *Controller) observeTelemetryPacket(sessionID string, roomID string, lane string, packetFamily string, encodedBytes int) {
	if encodedBytes < 0 {
		encodedBytes = 0
	}
	controller.mu.Lock()
	state := controller.telemetry[sessionID]
	if state == nil || state.roomID != roomID {
		state = newLiveTelemetryState(roomID)
		controller.telemetry[sessionID] = state
	}
	controller.mu.Unlock()

	state.mu.Lock()
	metric := state.packets[lane]
	metric.PacketCount++
	metric.EncodedBytesTotal += uint64(encodedBytes)
	metric.LastEncodedBytes = encodedBytes
	if encodedBytes > metric.MaximumEncodedBytes {
		metric.MaximumEncodedBytes = encodedBytes
	}
	metric.LastPacketFamily = packetFamily
	state.packets[lane] = metric
	state.mu.Unlock()
}

func (controller *Controller) TelemetrySnapshot(context networkingtooling.Context) (protocol.TelemetrySnapshot, error) {
	room, gameInstance, err := controller.resolveRoomGame(context)
	if err != nil {
		return protocol.TelemetrySnapshot{}, err
	}
	counts := gameInstance.RuntimeEntityCounts()
	now := time.Now()
	process := measurement.ProcessSample{}
	if controller.processSampler != nil {
		process = controller.processSampler.Sample(now)
	}
	sequence, laneMetrics, packetsOut, bytesOut, maxPacketBytes := controller.telemetryPacketSnapshot(context.SessionID, context.RoomID)
	metrics := map[string]any{
		"server_room_count":              controller.rooms.RoomCount(),
		"server_match_id":                room.CurrentMatchID(),
		"server_players":                 counts.Players,
		"server_player_sessions":         counts.PlayerSessions,
		"server_enemies":                 counts.Enemies,
		"server_asteroids":               counts.Asteroids,
		"server_projectiles":             counts.Projectiles,
		"server_pickups":                 counts.Pickups,
		"server_radial_effects":          counts.RadialEffects,
		"server_total_asteroids_spawned": counts.AsteroidsSpawnedTotal,
		"server_heap_allocated_bytes":    process.HeapAllocatedBytes,
		"server_heap_in_use_bytes":       process.HeapInUseBytes,
		"server_system_bytes":            process.SystemBytes,
		"server_goroutines":              process.Goroutines,
		"server_gc_cycles":               process.GCCycles,
		"server_packets_out":             packetsOut,
		"server_bytes_out":               bytesOut,
		"server_max_packet_bytes":        maxPacketBytes,
		"server_lane_metrics":            laneMetrics,
	}
	return protocol.TelemetrySnapshot{
		Type:          protocol.PacketTypeTelemetrySnapshot,
		TimestampMsec: int(now.UnixMilli()),
		Sequence:      sequence,
		Metrics:       metrics,
	}, nil
}

func (controller *Controller) telemetryPacketSnapshot(sessionID string, roomID string) (int, map[string]any, uint64, uint64, int) {
	controller.mu.Lock()
	state := controller.telemetry[sessionID]
	if state == nil || state.roomID != roomID {
		state = newLiveTelemetryState(roomID)
		controller.telemetry[sessionID] = state
	}
	controller.mu.Unlock()

	state.mu.Lock()
	defer state.mu.Unlock()
	state.sequence++
	lanes := make(map[string]any, len(state.packets))
	var packetCount uint64
	var encodedBytes uint64
	maximumEncodedBytes := 0
	for lane, metric := range state.packets {
		lanes[lane] = map[string]any{
			"packet_count":          metric.PacketCount,
			"encoded_bytes_total":   metric.EncodedBytesTotal,
			"last_encoded_bytes":    metric.LastEncodedBytes,
			"maximum_encoded_bytes": metric.MaximumEncodedBytes,
			"last_packet_family":    metric.LastPacketFamily,
		}
		packetCount += metric.PacketCount
		encodedBytes += metric.EncodedBytesTotal
		if metric.MaximumEncodedBytes > maximumEncodedBytes {
			maximumEncodedBytes = metric.MaximumEncodedBytes
		}
	}
	return state.sequence, lanes, packetCount, encodedBytes, maximumEncodedBytes
}

func (controller *Controller) CloseTelemetry(context networkingtooling.Context) {
	if context.SessionID == "" {
		return
	}
	controller.mu.Lock()
	delete(controller.telemetry, context.SessionID)
	controller.mu.Unlock()
}
