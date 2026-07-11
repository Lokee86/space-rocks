package realtime

import (
	"reflect"
	"testing"
)

type invalidRealtimeLanePayload struct {
	lane     Lane
	kind     RealtimeLaneCandidateKind
	family   string
	metadata Metadata
	present  bool
}

func (payload invalidRealtimeLanePayload) Lane() Lane { return payload.lane }
func (payload invalidRealtimeLanePayload) CandidateKind() RealtimeLaneCandidateKind {
	return payload.kind
}
func (payload invalidRealtimeLanePayload) PacketFamily() string { return payload.family }
func (payload invalidRealtimeLanePayload) LaneMetadata() (Metadata, bool) {
	return payload.metadata, payload.present
}
func (invalidRealtimeLanePayload) WirePacket() map[string]any { return map[string]any{} }
func (invalidRealtimeLanePayload) realtimeLanePayload()        {}

func TestRealtimeLanePayloadImplementationsExposeExpectedContract(t *testing.T) {
	metadata := Metadata{Lane: LaneWorld, Sequence: 1}
	payloads := []RealtimeLanePayload{
		WorldFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata},
		WorldWireFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata},
		WorldDeltaPacket{Type: PacketFamilyWorldDelta, Metadata: metadata},
		WorldWireDeltaPacket{Type: PacketFamilyWorldDelta, Metadata: metadata},
		OverlayFullPacket{Type: PacketFamilyOverlayFull, Metadata: Metadata{Lane: LaneOverlay}},
		OverlayWireFullPacket{Type: PacketFamilyOverlayFull, Metadata: Metadata{Lane: LaneOverlay}},
		OverlayLaneDelta{Metadata: Metadata{Lane: LaneOverlay}},
		OverlayWireLaneDelta{Metadata: Metadata{Lane: LaneOverlay}},
		SessionFullPacket{Type: PacketFamilySessionFull, Metadata: Metadata{Lane: LaneSession}},
		SessionWireFullPacket{Type: PacketFamilySessionFull, Metadata: Metadata{Lane: LaneSession}},
		SessionLaneDelta{Metadata: Metadata{Lane: LaneSession}},
		SessionWireLaneDelta{Metadata: Metadata{Lane: LaneSession}},
		AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroids}},
		BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{Lane: LaneBulletsLifecycle}},
		EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}},
	}

	for _, payload := range payloads {
		if err := ValidateRealtimeLanePayload(payload); err != nil {
			t.Errorf("%T rejected: %v", payload, err)
		}
		if payload.WirePacket() == nil {
			t.Errorf("%T returned nil wire packet", payload)
		}
	}
}

func TestValidateRealtimeLanePayloadRejectsInvalidPayloads(t *testing.T) {
	valid := invalidRealtimeLanePayload{
		lane: LaneWorld, kind: RealtimeLaneCandidateKindFull, family: PacketFamilyWorldFull,
		metadata: Metadata{Lane: LaneWorld}, present: true,
	}
	tests := []struct {
		name    string
		payload RealtimeLanePayload
	}{
		{"nil", nil},
		{"typed nil", (*WorldFullPacket)(nil)},
		{"empty lane", invalidRealtimeLanePayload{kind: valid.kind, family: valid.family, metadata: valid.metadata, present: true}},
		{"empty kind", invalidRealtimeLanePayload{lane: valid.lane, family: valid.family, metadata: valid.metadata, present: true}},
		{"empty family", invalidRealtimeLanePayload{lane: valid.lane, kind: valid.kind, metadata: valid.metadata, present: true}},
		{"missing metadata", invalidRealtimeLanePayload{lane: valid.lane, kind: valid.kind, family: valid.family}},
		{"metadata lane mismatch", invalidRealtimeLanePayload{lane: valid.lane, kind: valid.kind, family: valid.family, metadata: Metadata{Lane: LaneOverlay}, present: true}},
		{"unsupported matrix combination", invalidRealtimeLanePayload{lane: LaneWorld, kind: RealtimeLaneCandidateKindEventBatch, family: PacketFamilyEventBatch, metadata: Metadata{Lane: LaneWorld}, present: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateRealtimeLanePayload(test.payload); err == nil {
				t.Fatal("expected payload validation error")
			}
			if _, err := NewRealtimeLaneCandidate(test.payload, nil); err == nil {
				t.Fatal("expected constructor validation error")
			}
		})
	}
}

func TestNewRealtimeLaneCandidateStoresTypedPayloadAndProjection(t *testing.T) {
	payload := WorldFullPacket{Type: PacketFamilyWorldFull, Metadata: Metadata{Lane: LaneWorld, Sequence: 3}}
	projection := "projection"

	candidate, err := NewRealtimeLaneCandidate(payload, projection)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	if !reflect.DeepEqual(candidate.Payload, payload) || candidate.Lane() != LaneWorld || candidate.Kind() != RealtimeLaneCandidateKindFull || candidate.Projection != projection {
		t.Fatalf("candidate = %#v", candidate)
	}
}

func TestRealtimeLaneCandidateMethodsAreNilSafe(t *testing.T) {
	var payload *WorldFullPacket
	candidate := RealtimeLaneCandidate{Payload: payload}
	if candidate.Lane() != "" || candidate.Kind() != "" || candidate.PacketFamily() != "" {
		t.Fatalf("typed-nil candidate methods returned data: %#v", candidate)
	}
	if metadata, ok := candidate.Metadata(); ok || metadata != (Metadata{}) {
		t.Fatalf("typed-nil candidate metadata = %#v, %v", metadata, ok)
	}
}
