package realtime

import "testing"

type stage4UnknownPayload struct{}

func (stage4UnknownPayload) Lane() Lane { return LaneWorld }
func (stage4UnknownPayload) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (stage4UnknownPayload) PacketFamily() string { return PacketFamilyWorldFull }
func (stage4UnknownPayload) LaneMetadata() (Metadata, bool) {
	return Metadata{Lane: LaneWorld}, true
}
func (stage4UnknownPayload) WirePacket() map[string]any {
	return map[string]any{"type": PacketFamilyWorldFull}
}
func (stage4UnknownPayload) realtimeLanePayload() {}

type stage4EmptyWirePayload struct{}

func (stage4EmptyWirePayload) Lane() Lane { return LaneWorld }
func (stage4EmptyWirePayload) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (stage4EmptyWirePayload) PacketFamily() string { return PacketFamilyWorldFull }
func (stage4EmptyWirePayload) LaneMetadata() (Metadata, bool) {
	return Metadata{Lane: LaneWorld}, true
}
func (stage4EmptyWirePayload) WirePacket() map[string]any { return nil }
func (stage4EmptyWirePayload) realtimeLanePayload()       {}

type stage4MissingTypePayload struct{}

func (stage4MissingTypePayload) Lane() Lane { return LaneWorld }
func (stage4MissingTypePayload) CandidateKind() RealtimeLaneCandidateKind {
	return RealtimeLaneCandidateKindFull
}
func (stage4MissingTypePayload) PacketFamily() string { return PacketFamilyWorldFull }
func (stage4MissingTypePayload) LaneMetadata() (Metadata, bool) {
	return Metadata{Lane: LaneWorld}, true
}
func (stage4MissingTypePayload) WirePacket() map[string]any {
	return map[string]any{"ships": []any{}}
}
func (stage4MissingTypePayload) realtimeLanePayload() {}

func stage4WorldCandidate() RealtimeLaneCandidate {
	return mustRealtimeLaneCandidate(WorldFullPacket{
		Type:     PacketFamilyWorldFull,
		Metadata: Metadata{Lane: LaneWorld},
	}, nil)
}

func TestSupportedRealtimeLanePayloadsValidateAndWire(t *testing.T) {
	metadata := func(lane Lane) Metadata {
		return Metadata{Lane: lane, Sequence: 1}
	}
	cases := []struct {
		name    string
		payload RealtimeLanePayload
		lane    Lane
		kind    RealtimeLaneCandidateKind
		family  string
	}{
		{
			name:    "world-full",
			payload: WorldFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata(LaneWorld)},
			lane:    LaneWorld,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilyWorldFull,
		},
		{
			name:    "world-wire-full",
			payload: WorldWireFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata(LaneWorld)},
			lane:    LaneWorld,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilyWorldFull,
		},
		{
			name:    "world-delta",
			payload: WorldDeltaPacket{Type: PacketFamilyWorldDelta, Metadata: metadata(LaneWorld)},
			lane:    LaneWorld,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyWorldDelta,
		},
		{
			name:    "world-wire-delta",
			payload: WorldWireDeltaPacket{Type: PacketFamilyWorldDelta, Metadata: metadata(LaneWorld)},
			lane:    LaneWorld,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyWorldDelta,
		},
		{
			name:    "overlay-full",
			payload: OverlayFullPacket{Type: PacketFamilyOverlayFull, Metadata: metadata(LaneOverlay)},
			lane:    LaneOverlay,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilyOverlayFull,
		},
		{
			name:    "overlay-wire-full",
			payload: OverlayWireFullPacket{Type: PacketFamilyOverlayFull, Metadata: metadata(LaneOverlay)},
			lane:    LaneOverlay,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilyOverlayFull,
		},
		{
			name:    "overlay-delta",
			payload: OverlayLaneDelta{Metadata: metadata(LaneOverlay)},
			lane:    LaneOverlay,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyOverlayDelta,
		},
		{
			name:    "overlay-wire-delta",
			payload: OverlayWireLaneDelta{Metadata: metadata(LaneOverlay)},
			lane:    LaneOverlay,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyOverlayDelta,
		},
		{
			name:    "session-full",
			payload: SessionFullPacket{Type: PacketFamilySessionFull, Metadata: metadata(LaneSession)},
			lane:    LaneSession,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilySessionFull,
		},
		{
			name:    "session-wire-full",
			payload: SessionWireFullPacket{Type: PacketFamilySessionFull, Metadata: metadata(LaneSession)},
			lane:    LaneSession,
			kind:    RealtimeLaneCandidateKindFull,
			family:  PacketFamilySessionFull,
		},
		{
			name:    "session-delta",
			payload: SessionLaneDelta{Metadata: metadata(LaneSession)},
			lane:    LaneSession,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilySessionDelta,
		},
		{
			name:    "session-wire-delta",
			payload: SessionWireLaneDelta{Metadata: metadata(LaneSession)},
			lane:    LaneSession,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilySessionDelta,
		},
		{
			name:    "asteroid-hot-delta",
			payload: AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: metadata(LaneAsteroids)},
			lane:    LaneAsteroids,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyAsteroidDelta,
		},
		{
			name:    "asteroid-lifecycle-delta",
			payload: AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: metadata(LaneAsteroidsLifecycle)},
			lane:    LaneAsteroidsLifecycle,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyAsteroidsLifecycle,
		},
		{
			name:    "bullet-hot-delta",
			payload: BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: metadata(LaneBullets)},
			lane:    LaneBullets,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyBulletDelta,
		},
		{
			name:    "bullet-lifecycle-delta",
			payload: BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: metadata(LaneBulletsLifecycle)},
			lane:    LaneBulletsLifecycle,
			kind:    RealtimeLaneCandidateKindDelta,
			family:  PacketFamilyBulletsLifecycle,
		},
		{
			name:    "event-batch",
			payload: EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: metadata(LaneEvent)},
			lane:    LaneEvent,
			kind:    RealtimeLaneCandidateKindEventBatch,
			family:  PacketFamilyEventBatch,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate, err := NewRealtimeLaneCandidate(tc.payload, nil)
			if err != nil {
				t.Fatalf("validate payload: %v", err)
			}
			if candidate.Lane() != tc.lane || candidate.Kind() != tc.kind || candidate.PacketFamily() != tc.family {
				t.Fatalf("candidate identity mismatch")
			}
			metadata, ok := candidate.Metadata()
			if !ok || metadata.Lane != tc.lane {
				t.Fatalf("metadata = %#v, ok=%v", metadata, ok)
			}
			wire, err := WireLanePacket(candidate)
			if err != nil {
				t.Fatalf("wire: %v", err)
			}
			if len(wire) == 0 || wire["type"] != tc.family {
				t.Fatalf("invalid wire: %#v", wire)
			}
		})
	}
}

func TestValidateWireLaneMap(t *testing.T) {
	candidate := stage4WorldCandidate()
	cases := []struct {
		name    string
		wire    map[string]any
		wantErr bool
	}{
		{name: "valid", wire: map[string]any{"type": PacketFamilyWorldFull}},
		{name: "empty", wire: map[string]any{}, wantErr: true},
		{name: "missing-type", wire: map[string]any{"ships": []any{}}, wantErr: true},
		{name: "non-string-type", wire: map[string]any{"type": 1}, wantErr: true},
		{name: "blank-type", wire: map[string]any{"type": "  "}, wantErr: true},
		{name: "mismatched-type", wire: map[string]any{"type": PacketFamilyWorldDelta}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if (validateWireLaneMap(candidate, tc.wire) != nil) != tc.wantErr {
				t.Fatalf("unexpected validation result")
			}
		})
	}
}

func TestWireLanePacketRejectsInvalidPayloads(t *testing.T) {
	var typedNil *WorldFullPacket
	supportedPointer := &WorldFullPacket{
		Type:     PacketFamilyWorldFull,
		Metadata: Metadata{Lane: LaneWorld},
	}
	cases := []RealtimeLaneCandidate{
		{},
		{Payload: typedNil},
		{Payload: supportedPointer},
		{Payload: stage4UnknownPayload{}},
		{Payload: stage4EmptyWirePayload{}},
		{Payload: stage4MissingTypePayload{}},
	}

	for i, candidate := range cases {
		if _, err := WireLanePacket(candidate); err == nil {
			t.Fatalf("case %d: expected encoding failure", i)
		}
	}
}

func TestEncodeLanePacketRejectsNilAndTypedNil(t *testing.T) {
	var typedNil *WorldFullPacket
	cases := []RealtimeLaneCandidate{{}, {Payload: typedNil}}
	for i, candidate := range cases {
		if _, _, err := encodeLanePacket(candidate); err == nil {
			t.Fatalf("case %d: expected encoding failure", i)
		}
	}
}

func TestSchedulerAndDiagnosticsUseTypedPayloadIdentity(t *testing.T) {
	candidate := stage4WorldCandidate()
	record := scheduleRecordForCandidate(2, candidate)
	diagnostics := CandidateWriteDiagnosticsFor(candidate, NewRealtimeSessionState("player-1", "match-1"), 17)
	if record.Lane != candidate.Lane() || record.RecordKind != string(candidate.Kind()) || record.PacketFamily != candidate.PacketFamily() {
		t.Fatalf("schedule record diverged: %#v", record)
	}
	if diagnostics.Lane != candidate.Lane() || diagnostics.Kind != candidate.Kind() || diagnostics.PacketFamily != candidate.PacketFamily() {
		t.Fatalf("diagnostics diverged: %#v", diagnostics)
	}
	metadata, ok := candidate.Metadata()
	if !ok || diagnostics.Sequence != metadata.Sequence {
		t.Fatalf("diagnostics metadata diverged: %#v", diagnostics)
	}
}
