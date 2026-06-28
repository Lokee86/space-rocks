package packetmetrics

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/logging"
)

const (
	serverSnapshotTargetBytes  = 500
	serverSnapshotWarningBytes = 800
	serverSnapshotDangerBytes  = 1100
	serverSnapshotHardCapBytes = 1200

	gameplayPresentationSlowWriteThreshold = 20 * time.Millisecond
)

type GameplayPresentationPacketMetrics struct {
	PacketSize     int
	PacketSeverity string
	Contributors   GameplayPacketContributors
	BuildDuration  time.Duration
	EncodeDuration time.Duration
}

func NewGameplayPresentationPacketMetrics(packetSize int, contributors GameplayPacketContributors, buildDuration time.Duration, encodeDuration time.Duration) GameplayPresentationPacketMetrics {
	return GameplayPresentationPacketMetrics{
		PacketSize:     packetSize,
		PacketSeverity: classifyGameplaySnapshotPacketSize(packetSize),
		Contributors:   contributors,
		BuildDuration:  buildDuration,
		EncodeDuration: encodeDuration,
	}
}

func classifyGameplaySnapshotPacketSize(packetSize int) string {
	switch {
	case packetSize <= serverSnapshotTargetBytes:
		return "target"
	case packetSize <= serverSnapshotWarningBytes:
		return "above_target"
	case packetSize < serverSnapshotDangerBytes:
		return "warning"
	case packetSize < serverSnapshotHardCapBytes:
		return "danger"
	default:
		return "hard_cap"
	}
}

func LogGameplayPresentationPacketSize(metrics GameplayPresentationPacketMetrics, roomID string, playerID string, remoteAddr string) {
	if metrics.PacketSeverity == "target" || metrics.PacketSeverity == "above_target" {
		return
	}

	logging.Network.Warn("gameplay presentation packet over budget",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
		"packet_size", metrics.PacketSize,
		"packet_category", "server_snapshot",
		"packet_severity", metrics.PacketSeverity,
		"snapshot_target_bytes", serverSnapshotTargetBytes,
		"snapshot_warning_bytes", serverSnapshotWarningBytes,
		"snapshot_danger_bytes", serverSnapshotDangerBytes,
		"snapshot_hard_cap_bytes", serverSnapshotHardCapBytes,
		"room_state", metrics.Contributors.RoomState,
		"players", metrics.Contributors.Players,
		"player_sessions", metrics.Contributors.PlayerSessions,
		"player_lifecycle", metrics.Contributors.PlayerLifecycle,
		"asteroids", metrics.Contributors.Asteroids,
		"bullets", metrics.Contributors.Bullets,
		"pickups", metrics.Contributors.Pickups,
		"enemies", metrics.Contributors.Enemies,
		"events", metrics.Contributors.Events,
		"total_asteroids", metrics.Contributors.TotalAsteroids,
		"build_duration_ms", metrics.BuildDuration.Milliseconds(),
		"encode_duration_ms", metrics.EncodeDuration.Milliseconds(),
	)
}

func LogSlowGameplayPresentationWrite(duration time.Duration, metrics GameplayPresentationPacketMetrics, roomID string, playerID string, remoteAddr string) {
	if duration <= gameplayPresentationSlowWriteThreshold {
		return
	}

	logging.Network.Warn("gameplay presentation packet write slow",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
		"write_duration_ms", duration.Milliseconds(),
		"packet_size", metrics.PacketSize,
		"packet_category", "server_snapshot",
		"packet_severity", metrics.PacketSeverity,
		"snapshot_target_bytes", serverSnapshotTargetBytes,
		"snapshot_warning_bytes", serverSnapshotWarningBytes,
		"snapshot_danger_bytes", serverSnapshotDangerBytes,
		"snapshot_hard_cap_bytes", serverSnapshotHardCapBytes,
		"room_state", metrics.Contributors.RoomState,
		"players", metrics.Contributors.Players,
		"player_sessions", metrics.Contributors.PlayerSessions,
		"player_lifecycle", metrics.Contributors.PlayerLifecycle,
		"asteroids", metrics.Contributors.Asteroids,
		"bullets", metrics.Contributors.Bullets,
		"pickups", metrics.Contributors.Pickups,
		"enemies", metrics.Contributors.Enemies,
		"events", metrics.Contributors.Events,
		"total_asteroids", metrics.Contributors.TotalAsteroids,
		"build_duration_ms", metrics.BuildDuration.Milliseconds(),
		"encode_duration_ms", metrics.EncodeDuration.Milliseconds(),
	)
}