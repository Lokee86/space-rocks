package packetmetrics

import "github.com/Lokee86/space-rocks/server/internal/logging"

func LogShadowLaneMetrics(records []PacketMetricRecord, roomID string, playerID string, remoteAddr string) {
	logLaneMetrics(records, roomID, playerID, remoteAddr)
}

func LogSentLaneMetrics(records []PacketMetricRecord, roomID string, playerID string, remoteAddr string) {
	logLaneMetrics(records, roomID, playerID, remoteAddr)
}

func logLaneMetrics(records []PacketMetricRecord, roomID string, playerID string, remoteAddr string) {
	for _, record := range records {
		logging.Network.Debug("realtime lane metric",
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
			"packet_family", record.PacketFamily,
			"lane", record.Lane,
			"bytes", record.Bytes,
		)
	}
}
