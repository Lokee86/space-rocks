package realtime

type RealtimeLaneCandidateKind string

const (
	RealtimeLaneCandidateKindFull       RealtimeLaneCandidateKind = "full"
	RealtimeLaneCandidateKindDelta      RealtimeLaneCandidateKind = "delta"
	RealtimeLaneCandidateKindEventBatch RealtimeLaneCandidateKind = "event_batch"
)

type RealtimeLaneCandidate struct {
	Payload    RealtimeLanePayload
	Projection any
	MatchID    string
}

func (candidate RealtimeLaneCandidate) hasPayload() bool {
	return candidate.Payload != nil && !isTypedNil(candidate.Payload)
}

func (candidate RealtimeLaneCandidate) Lane() Lane {
	if !candidate.hasPayload() {
		return ""
	}
	return candidate.Payload.Lane()
}

func (candidate RealtimeLaneCandidate) Kind() RealtimeLaneCandidateKind {
	if !candidate.hasPayload() {
		return ""
	}
	return candidate.Payload.CandidateKind()
}

func (candidate RealtimeLaneCandidate) PacketFamily() string {
	if !candidate.hasPayload() {
		return ""
	}
	return candidate.Payload.PacketFamily()
}

func (candidate RealtimeLaneCandidate) Metadata() (Metadata, bool) {
	if !candidate.hasPayload() {
		return Metadata{}, false
	}
	metadata, ok := candidate.Payload.LaneMetadata()
	if ok {
		metadata.MatchID = candidate.MatchID
	}
	return metadata, ok
}

type RealtimeLanePlan struct {
	Candidates []RealtimeLaneCandidate
}

type RealtimeSendPrepared struct {
	CandidatePlan RealtimeLanePlan
	Records       []ScheduleRecord
	SendPlan      SendPlan
	SessionState  RealtimeSessionState
}
