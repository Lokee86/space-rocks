package packetmetrics

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/logging"
)

const gameplayPresentationSlowWriteThreshold = 20 * time.Millisecond

type GameplayPresentationPacketMetrics struct {
	PacketSize     int
	Contributors   GameplayPacketContributors
	BuildDuration  time.Duration
	EncodeDuration time.Duration
}

func NewGameplayPresentationPacketMetrics(packetSize int, contributors GameplayPacketContributors, buildDuration time.Duration, encodeDuration time.Duration) GameplayPresentationPacketMetrics {
	return GameplayPresentationPacketMetrics{
		PacketSize:     packetSize,
		Contributors:   contributors,
		BuildDuration:  buildDuration,
		EncodeDuration: encodeDuration,
	}
}

func LogGameplayPresentationPacketSize(metrics GameplayPresentationPacketMetrics, roomID string, playerID string, remoteAddr string) {
	_ = metrics
	_ = roomID
	_ = playerID
	_ = remoteAddr
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
