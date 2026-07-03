package packetmetrics

// PacketMetricRecord captures generic realtime packet plan metrics for logging and later aggregation.
type PacketMetricRecord struct {
	PacketFamily   string
	Lane           string
	Bytes          int
	SendReason     string
	ChunkDecision  string
	ResyncDecision string
}

// Clone returns a copy of the record for safe logging/aggregation handoff.
func (r PacketMetricRecord) Clone() PacketMetricRecord {
	return r
}

func NewPacketMetricRecord() PacketMetricRecord {
	return PacketMetricRecord{}
}
