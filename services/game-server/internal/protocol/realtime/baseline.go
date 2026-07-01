package realtime

type RealtimeLaneState struct {
	Lane         Lane
	SnapshotKind SnapshotKind
	Sequence     int
	BaselineID   string
	SnapshotID   string
	ChunkIndex   int
	ChunkCount   int
	IsFinalChunk bool
}

func (state RealtimeLaneState) Metadata() Metadata {
	return Metadata{
		Lane:         state.Lane,
		Sequence:     state.Sequence,
		BaselineID:   state.BaselineID,
		SnapshotID:   state.SnapshotID,
		SnapshotKind: state.SnapshotKind,
		ChunkIndex:   state.ChunkIndex,
		ChunkCount:   state.ChunkCount,
		IsFinalChunk: state.IsFinalChunk,
	}
}

type RealtimeSessionState struct {
	ReceiverID          string
	Lanes               map[Lane]RealtimeLaneState
	BaselineReady       map[Lane]bool
	baselineProjections map[Lane]any
}

func NewRealtimeSessionState(receiverID string) RealtimeSessionState {
	return RealtimeSessionState{
		ReceiverID:          receiverID,
		Lanes:               make(map[Lane]RealtimeLaneState),
		BaselineReady:       make(map[Lane]bool),
		baselineProjections: make(map[Lane]any),
	}
}

func (state *RealtimeSessionState) UpdateLane(lane Lane, metadata Metadata) {
	if existing, ok := state.Lanes[lane]; ok && metadata.Sequence < existing.Sequence {
		return
	}
	state.Lanes[lane] = RealtimeLaneState{
		Lane:         lane,
		SnapshotKind: metadata.SnapshotKind,
		Sequence:     metadata.Sequence,
		BaselineID:   metadata.BaselineID,
		SnapshotID:   metadata.SnapshotID,
		ChunkIndex:   metadata.ChunkIndex,
		ChunkCount:   metadata.ChunkCount,
		IsFinalChunk: metadata.IsFinalChunk,
	}
}

func (state *RealtimeSessionState) MarkBaselineReady(lane Lane) {
	if state.BaselineReady == nil {
		state.BaselineReady = make(map[Lane]bool)
	}
	state.BaselineReady[lane] = true
}

func (state *RealtimeSessionState) StoreBaselineProjection(lane Lane, projection any) {
	if projection == nil {
		return
	}
	if state.baselineProjections == nil {
		state.baselineProjections = make(map[Lane]any)
	}
	state.baselineProjections[lane] = projection
}

func (state RealtimeSessionState) BaselineProjection(lane Lane) (any, bool) {
	if state.baselineProjections == nil {
		return nil, false
	}
	projection, ok := state.baselineProjections[lane]
	return projection, ok
}

func (state *RealtimeSessionState) ClearBaselineProjection(lane Lane) {
	if state.baselineProjections == nil {
		return
	}
	delete(state.baselineProjections, lane)
}

func (state RealtimeSessionState) LaneState(lane Lane) (RealtimeLaneState, bool) {
	laneState, ok := state.Lanes[lane]
	return laneState, ok
}

func (state RealtimeSessionState) LaneBaselineReady(lane Lane) bool {
	return state.BaselineReady[lane]
}

func NextLaneSequence(state RealtimeLaneState, synced bool) int {
	if !synced || state.Sequence < 1 {
		return 1
	}
	return state.Sequence + 1
}

func CandidateMetadata(candidate RealtimeLaneCandidate, state RealtimeSessionState) (Metadata, bool) {
	switch packet := candidate.Full.(type) {
	case WorldFullPacket:
		return packet.Metadata, true
	case WorldWireFullPacket:
		return packet.Metadata, true
	case OverlayFullPacket:
		return packet.Metadata, true
	case OverlayWireFullPacket:
		return packet.Metadata, true
	case SessionFullPacket:
		return packet.Metadata, true
	case SessionWireFullPacket:
		return packet.Metadata, true
	case EventBatchPacket:
		return packet.Metadata, true
	}

	switch packet := candidate.Delta.(type) {
	case WorldDeltaPacket:
		return packet.Metadata, true
	case WorldWireDeltaPacket:
		return packet.Metadata, true
	case OverlayLaneDelta:
		return packet.Metadata, true
	case OverlayWireLaneDelta:
		return packet.Metadata, true
	case SessionLaneDelta:
		return packet.Metadata, true
	case SessionWireLaneDelta:
		return packet.Metadata, true
	}

	laneState, ok := state.LaneState(candidate.Lane)
	if !ok {
		return Metadata{}, false
	}
	return laneState.Metadata(), true
}

func (state RealtimeSessionState) SharedWorldSnapshotID(snapshotID string, payloadsIdentical bool) string {
	if payloadsIdentical {
		return snapshotID
	}
	return ""
}

func AdvanceMetadataForSuccessfulWrite(lane Lane, metadata Metadata) Metadata {
	if lane != LaneEvent {
		return metadata
	}
	metadata.Sequence += 1
	metadata.SnapshotID = sequenceBackedBatchID(metadata.Sequence)
	return metadata
}

