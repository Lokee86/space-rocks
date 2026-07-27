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
	MatchID             string
	Lanes               map[Lane]RealtimeLaneState
	BaselineReady       map[Lane]bool
	baselineProjections map[Lane]any
	packetSequences     map[string]int
	packetProjections   map[string]any
	HotLaneCohorts      HotLaneCohortState
	HotLaneTick         int
}

func NewRealtimeSessionState(receiverID string, matchID string) RealtimeSessionState {
	return RealtimeSessionState{
		ReceiverID:          receiverID,
		MatchID:             matchID,
		Lanes:               make(map[Lane]RealtimeLaneState),
		BaselineReady:       make(map[Lane]bool),
		baselineProjections: make(map[Lane]any),
		packetSequences:     make(map[string]int),
		packetProjections:   make(map[string]any),
		HotLaneCohorts:      NewHotLaneCohortState(),
	}
}

func (state *RealtimeSessionState) AdvanceHotLaneTick() int {
	state.HotLaneTick++
	if state.HotLaneTick <= 0 {
		state.HotLaneTick = 1
	}
	return state.HotLaneTick
}

func (state RealtimeSessionState) IdentityMatches(receiverID, matchID string) bool {
	return state.ReceiverID == receiverID && state.MatchID == matchID
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

func (state RealtimeSessionState) PacketSequence(packetFamily string) int {
	if state.packetSequences == nil {
		return 0
	}
	return state.packetSequences[packetFamily]
}

func (state *RealtimeSessionState) UpdatePacketSequence(packetFamily string, sequence int) {
	if packetFamily == "" {
		return
	}
	if state.packetSequences == nil {
		state.packetSequences = make(map[string]int)
	}
	if sequence < state.packetSequences[packetFamily] {
		return
	}
	state.packetSequences[packetFamily] = sequence
}

func (state RealtimeSessionState) PacketProjection(packetFamily string) (any, bool) {
	if state.packetProjections == nil {
		return nil, false
	}
	projection, ok := state.packetProjections[packetFamily]
	return projection, ok
}

func (state *RealtimeSessionState) StorePacketProjection(packetFamily string, projection any) {
	if packetFamily == "" || projection == nil {
		return
	}
	if state.packetProjections == nil {
		state.packetProjections = make(map[string]any)
	}
	state.packetProjections[packetFamily] = projection
}

func (state *RealtimeSessionState) RequireFullBaseline(lane Lane) bool {
	if !IsBaselineLane(lane) {
		return false
	}
	if state.BaselineReady == nil {
		state.BaselineReady = make(map[Lane]bool)
	}
	state.BaselineReady[lane] = false
	state.ClearBaselineProjection(lane)
	return true
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

// CommitSuccessfulCandidate advances only the candidate that completed its
// transport write. Chunked candidates commit their projection on the final
// chunk so an incomplete same-sequence burst remains supersedable.
func CommitSuccessfulCandidate(state *RealtimeSessionState, candidate RealtimeLaneCandidate) {
	if state == nil {
		return
	}
	metadata, ok := candidate.Metadata()
	if !ok {
		return
	}
	persistedMetadata := AdvanceMetadataForSuccessfulWrite(candidate.Lane(), metadata)
	if candidate.PacketFamily() == PacketFamilyPlayerLocator {
		state.UpdatePacketSequence(PacketFamilyPlayerLocator, persistedMetadata.Sequence)
		if metadata.IsFinalChunk {
			if projection, ok := CandidateProjection(candidate); ok {
				state.StorePacketProjection(PacketFamilyPlayerLocator, projection)
			}
		}
		return
	}
	state.UpdateLane(candidate.Lane(), persistedMetadata)
	if !metadata.IsFinalChunk {
		return
	}
	projection, hasProjection := CandidateProjection(candidate)
	if hasProjection {
		state.StoreBaselineProjection(candidate.Lane(), projection)
	}
	if candidate.Kind() == RealtimeLaneCandidateKindFull {
		state.MarkBaselineReady(candidate.Lane())
	}
	if candidate.Lane() != LaneWorld || !hasProjection {
		return
	}
	world, ok := projection.(WorldWireFullPacket)
	if !ok {
		return
	}
	if candidate.Kind() == RealtimeLaneCandidateKindFull {
		seedHotLaneProjections(state, world)
		return
	}
	syncHotLaneProjectionMembership(state, world)
}
