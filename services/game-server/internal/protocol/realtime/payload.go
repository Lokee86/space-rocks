package realtime

import "fmt"

type RealtimeLanePayload interface {
	Lane() Lane
	CandidateKind() RealtimeLaneCandidateKind
	PacketFamily() string
	LaneMetadata() (Metadata, bool)
	WirePacket() map[string]any
	realtimeLanePayload()
}

var (
	_ RealtimeLanePayload = WorldFullPacket{}
	_ RealtimeLanePayload = WorldWireFullPacket{}
	_ RealtimeLanePayload = WorldDeltaPacket{}
	_ RealtimeLanePayload = WorldWireDeltaPacket{}
	_ RealtimeLanePayload = OverlayFullPacket{}
	_ RealtimeLanePayload = OverlayWireFullPacket{}
	_ RealtimeLanePayload = OverlayLaneDelta{}
	_ RealtimeLanePayload = OverlayWireLaneDelta{}
	_ RealtimeLanePayload = SessionFullPacket{}
	_ RealtimeLanePayload = SessionWireFullPacket{}
	_ RealtimeLanePayload = SessionLaneDelta{}
	_ RealtimeLanePayload = SessionWireLaneDelta{}
	_ RealtimeLanePayload = AsteroidWireDeltaPacket{}
	_ RealtimeLanePayload = BulletWireDeltaPacket{}
	_ RealtimeLanePayload = EventBatchPacket{}
)

func NewRealtimeLaneCandidate(payload RealtimeLanePayload, projection any) (RealtimeLaneCandidate, error) {
	if err := ValidateRealtimeLanePayload(payload); err != nil {
		return RealtimeLaneCandidate{}, err
	}
	return RealtimeLaneCandidate{
		Payload:    payload,
		Projection: projection,
	}, nil
}

func mustRealtimeLaneCandidate(payload RealtimeLanePayload, projection any) RealtimeLaneCandidate {
	candidate, err := NewRealtimeLaneCandidate(payload, projection)
	if err != nil {
		panic(fmt.Sprintf("invalid realtime lane payload: %v", err))
	}
	return candidate
}
