package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestAssembleRealtimeLaneCandidatesUsesNextSessionSequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", ShipType: "v_wing"}}}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneSession)
	if !ok {
		t.Fatalf("expected session candidate")
	}
	full, ok := candidate.Payload.(SessionWireFullPacket)
	if !ok {
		t.Fatalf("session candidate full type = %T, want SessionWireFullPacket", candidate.Payload)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("session full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesOmitsSessionWhenStoredBaselineMatches(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 5, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		TotalAsteroids:  5,
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{Sequence: 7, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, snapshot, 7))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneSession); ok {
		t.Fatalf("expected no session candidate when stored baseline matches, got %#v", plan.Candidates)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsSessionDeltaWhenStoredBaselineDiffers(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 9, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
		PlayerLifecycle: map[string]string{"player-1": "respawning"},
		TotalAsteroids:  8,
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{Sequence: 8, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", ShipType: "v_wing", Score: 5, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"}},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		TotalAsteroids:  5,
	}, 7))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	session, ok := findCandidateByLane(plan.Candidates, LaneSession)
	if !ok {
		t.Fatal("expected session delta candidate when stored baseline differs")
	}
	if session.Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected session delta candidate, got kind=%q", session.Kind())
	}
	if _, ok := session.Payload.(SessionWireLaneDelta); !ok {
		t.Fatalf("expected session delta packet, got %T", session.Payload)
	}
	if _, ok := session.Projection.(SessionWireFullPacket); !ok {
		t.Fatalf("expected current session full projection to be carried, got %T", session.Projection)
	}
}