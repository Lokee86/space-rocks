package realtime

const (
	TargetBytes   = 500
	WarningBytes  = 800
	DangerBytes   = 1100
	HardCapBytes  = 1200
)

func EstimatePacketBytes(packetFamily string, recordCount int, payloadBytes int) int {
	overhead := 64
	switch packetFamily {
	case PacketFamilyWorldFull, PacketFamilyWorldDelta:
		overhead = 96
	case PacketFamilyOverlayFull, PacketFamilyOverlayDelta:
		overhead = 80
	case PacketFamilySessionFull, PacketFamilySessionDelta:
		overhead = 88
	case PacketFamilyEventBatch:
		overhead = 72
	case PacketFamilyResyncRequest, PacketFamilyResyncRequired:
		overhead = 48
	}

	return overhead + (recordCount * 24) + payloadBytes
}
