package realtime

type RealtimeLaneState struct {
	Lane         Lane
	SnapshotKind SnapshotKind
	Sequence     int
	BaselineID   string
	SnapshotID   string
	ChunkIndex   int
	ChunkCount   int
	IsFinalChunk  bool
}

type RealtimeSessionState struct {
	ReceiverID string
	Lanes      map[Lane]RealtimeLaneState
}

func NewRealtimeSessionState(receiverID string) RealtimeSessionState {
	return RealtimeSessionState{
		ReceiverID: receiverID,
		Lanes:      make(map[Lane]RealtimeLaneState),
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

func (state RealtimeSessionState) LaneState(lane Lane) (RealtimeLaneState, bool) {
	laneState, ok := state.Lanes[lane]
	return laneState, ok
}

func (state RealtimeSessionState) SharedWorldSnapshotID(snapshotID string, payloadsIdentical bool) string {
	if payloadsIdentical {
		return snapshotID
	}
	return ""
}
