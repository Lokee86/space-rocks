package realtime

type CandidateWriteDiagnostics struct {
	PacketFamily           string
	Lane                   Lane
	Kind                   RealtimeLaneCandidateKind
	Sequence               int
	BaselineID             string
	SnapshotID             string
	SnapshotKind           SnapshotKind
	ChunkIndex             int
	ChunkCount             int
	IsFinalChunk           bool
	Channel                string
	EncodedBytes           int
	WorldHotCount          int
	ShipHotCount           int
	AsteroidHotCount       int
	BulletHotCount         int
	ShipOffloadedCount     int
	AsteroidOffloadedCount int
	BulletOffloadedCount   int
	ShipMode               HotLaneMode
	AsteroidMode           HotLaneMode
	BulletMode             HotLaneMode
	Cadence                string
	PacketOverTarget       bool
	PacketOverHardCap      bool
}

func CandidateWriteDiagnosticsFor(candidate RealtimeLaneCandidate, state RealtimeSessionState, encodedBytes int) CandidateWriteDiagnostics {
	lane, kind := candidate.Lane(), candidate.Kind()
	diagnostics := CandidateWriteDiagnostics{
		PacketFamily: candidate.PacketFamily(),
		Lane:         lane,
		Kind:         kind,
		Channel:      string(lane),
		EncodedBytes: encodedBytes,
	}
	if lane == LaneShips || lane == LaneAsteroids || lane == LaneBullets {
		diagnostics.Cadence = hotPacketCadenceForDiagnostics(candidate, state)
		diagnostics.WorldHotCount, diagnostics.ShipHotCount, diagnostics.AsteroidHotCount, diagnostics.BulletHotCount, diagnostics.ShipOffloadedCount, diagnostics.AsteroidOffloadedCount, diagnostics.BulletOffloadedCount = hotLaneCountsForDiagnostics(candidate)
		diagnostics.ShipMode, diagnostics.AsteroidMode, diagnostics.BulletMode = hotLaneModesForDiagnostics(state)
		diagnostics.PacketOverTarget = encodedBytes > WarningBytes && encodedBytes < HardCapBytes
		diagnostics.PacketOverHardCap = encodedBytes >= HardCapBytes
	}
	metadata, ok := candidate.Metadata()
	if !ok {
		return diagnostics
	}
	diagnostics.Sequence = metadata.Sequence
	diagnostics.BaselineID = metadata.BaselineID
	diagnostics.SnapshotID = metadata.SnapshotID
	diagnostics.SnapshotKind = metadata.SnapshotKind
	diagnostics.ChunkIndex = metadata.ChunkIndex
	diagnostics.ChunkCount = metadata.ChunkCount
	diagnostics.IsFinalChunk = metadata.IsFinalChunk
	return diagnostics
}

func hotPacketCadenceForDiagnostics(candidate RealtimeLaneCandidate, state RealtimeSessionState) string {
	lane := candidate.Lane()
	laneState, ok := state.LaneState(LaneWorld)
	if !ok {
		return "inline"
	}
	if lane == LaneShips {
		return hotPacketCadenceLabel(state.HotLaneCohorts.ShipMode, laneState.Sequence)
	}
	if lane == LaneAsteroids {
		return hotPacketCadenceLabel(state.HotLaneCohorts.AsteroidMode, laneState.Sequence)
	}
	if lane == LaneBullets {
		return hotPacketCadenceLabel(state.HotLaneCohorts.BulletMode, laneState.Sequence)
	}
	return ""
}

func hotPacketCadenceLabel(mode HotLaneMode, sequence int) string {
	switch mode {
	case HotLaneModeFullOwned60Hz:
		return "60hz"
	case HotLaneModeFullOwned30Hz:
		return "30hz"
	case HotLaneModeFullOwned20Hz:
		return "20hz"
	case HotLaneModeNeedsChunking:
		return "chunking"
	case HotLaneModeOverflow:
		return "overflow"
	default:
		return "inline"
	}
}

func hotLaneModesForDiagnostics(state RealtimeSessionState) (HotLaneMode, HotLaneMode, HotLaneMode) {
	return state.HotLaneCohorts.ShipMode, state.HotLaneCohorts.AsteroidMode, state.HotLaneCohorts.BulletMode
}

func hotLaneCountsForDiagnostics(candidate RealtimeLaneCandidate) (int, int, int, int, int, int, int) {
	switch delta := candidate.Payload.(type) {
	case ShipWireDeltaPacket:
		return 0, len(delta.ShipUpdates), 0, 0, len(delta.ShipUpdates), 0, 0
	case AsteroidWireDeltaPacket:
		return 0, 0, len(delta.AsteroidUpdates), 0, 0, len(delta.AsteroidUpdates), 0
	case BulletWireDeltaPacket:
		return 0, 0, 0, len(delta.BulletUpdates), 0, 0, len(delta.BulletUpdates)
	default:
		return 0, 0, 0, 0, 0, 0, 0
	}
}
