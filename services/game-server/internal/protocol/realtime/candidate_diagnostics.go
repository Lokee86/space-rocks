package realtime

type CandidateWriteDiagnostics struct {
	PacketFamily          string
	Lane                  Lane
	Kind                  RealtimeLaneCandidateKind
	Sequence              int
	BaselineID            string
	SnapshotID            string
	SnapshotKind          SnapshotKind
	ChunkIndex            int
	ChunkCount            int
	IsFinalChunk          bool
	Channel               string
	EncodedBytes          int
	WorldHotCount         int
	AsteroidHotCount      int
	BulletHotCount        int
	AsteroidOffloadedCount int
	BulletOffloadedCount  int
	AsteroidMode          HotLaneMode
	BulletMode            HotLaneMode
	Cadence               string
	PacketOverTarget      bool
	PacketOverHardCap     bool
}

func CandidateWriteDiagnosticsFor(candidate RealtimeLaneCandidate, state RealtimeSessionState, encodedBytes int) CandidateWriteDiagnostics {
	diagnostics := CandidateWriteDiagnostics{
		PacketFamily: packetFamilyForCandidate(candidate),
		Lane:         candidate.Lane,
		Kind:         candidate.Kind,
		Channel:      string(candidate.Lane),
		EncodedBytes: encodedBytes,
	}
	if candidate.Lane == LaneAsteroids || candidate.Lane == LaneBullets {
		diagnostics.Cadence = hotPacketCadenceForDiagnostics(candidate, state)
		diagnostics.WorldHotCount, diagnostics.AsteroidHotCount, diagnostics.BulletHotCount, diagnostics.AsteroidOffloadedCount, diagnostics.BulletOffloadedCount = hotLaneCountsForDiagnostics(candidate)
		diagnostics.AsteroidMode, diagnostics.BulletMode = hotLaneModesForDiagnostics(state)
		diagnostics.PacketOverTarget = encodedBytes > WarningBytes && encodedBytes < HardCapBytes
		diagnostics.PacketOverHardCap = encodedBytes >= HardCapBytes
	}
	metadata, ok := CandidateMetadata(candidate, state)
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
	laneState, ok := state.LaneState(LaneWorld)
	if !ok {
		return "inline"
	}
	if candidate.Lane == LaneAsteroids {
		return hotPacketCadenceLabel(state.HotLaneCohorts.AsteroidMode, laneState.Sequence)
	}
	if candidate.Lane == LaneBullets {
		return hotPacketCadenceLabel(state.HotLaneCohorts.BulletMode, laneState.Sequence)
	}
	return ""
}

func hotPacketCadenceLabel(mode HotLaneMode, sequence int) string {
	switch mode {
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

func hotLaneModesForDiagnostics(state RealtimeSessionState) (HotLaneMode, HotLaneMode) {
	return state.HotLaneCohorts.AsteroidMode, state.HotLaneCohorts.BulletMode
}

func hotLaneCountsForDiagnostics(candidate RealtimeLaneCandidate) (int, int, int, int, int) {
	switch delta := candidate.Delta.(type) {
	case AsteroidWireDeltaPacket:
		return 0, len(delta.AsteroidUpdates), 0, len(delta.AsteroidUpdates), 0
	case *AsteroidWireDeltaPacket:
		return 0, len(delta.AsteroidUpdates), 0, len(delta.AsteroidUpdates), 0
	case BulletWireDeltaPacket:
		return 0, 0, len(delta.BulletUpdates), 0, len(delta.BulletUpdates)
	case *BulletWireDeltaPacket:
		return 0, 0, len(delta.BulletUpdates), 0, len(delta.BulletUpdates)
	default:
		return 0, 0, 0, 0, 0
	}
}