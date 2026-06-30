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
			"record_count", record.RecordCount,
			"create_count", record.CreateCount,
			"update_count", record.UpdateCount,
			"delete_count", record.DeleteCount,
			"priority_band", record.PriorityBand,
			"deferred_count", record.DeferredCount,
			"superseded_count", record.SupersededCount,
			"required_count", record.RequiredCount,
			"budget_target", record.BudgetTarget,
			"budget_status", record.BudgetStatus,
		)
	}
}
