package packetmetrics

// PacketMetricRecord captures generic realtime packet plan metrics for logging and later aggregation.
type PacketMetricRecord struct {
	PacketFamily        string
	Lane                string
	Bytes               int
	SendReason          string
	ChunkDecision       string
	ResyncDecision      string
	Channel             string
	EncodedBytes        int
	WorldHotCount       int
	AsteroidHotCount    int
	BulletHotCount      int
	AsteroidOffloadedCount int
	BulletOffloadedCount    int
	AsteroidMode        string
	BulletMode          string
	Cadence             string
	PacketOverTarget    bool
	PacketOverHardCap   bool
}

// Clone returns a copy of the record for safe logging/aggregation handoff.
func (r PacketMetricRecord) Clone() PacketMetricRecord {
	return r
}

func NewPacketMetricRecord() PacketMetricRecord {
	return PacketMetricRecord{}
}
