package outbound

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/logging"
)

const (
	gameplayPresentationLargePacketThreshold = 4 * 1024
	gameplayPresentationSlowWriteThreshold    = 20 * time.Millisecond
)

func logGameplayPresentationPacketSize(packetSize int, roomID string, playerID string, remoteAddr string) {
	if packetSize <= gameplayPresentationLargePacketThreshold {
		return
	}

	logging.Network.Warn("gameplay presentation packet large",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
		"packet_size", packetSize,
	)
}

func LogSlowGameplayPresentationWrite(duration time.Duration, roomID string, playerID string, remoteAddr string) {
	if duration <= gameplayPresentationSlowWriteThreshold {
		return
	}

	logging.Network.Warn("gameplay presentation packet write slow",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
		"write_duration_ms", duration.Milliseconds(),
	)
}
