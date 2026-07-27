package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestApplyNetworkInterestFiltersByRecipientCameraAndWraps(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:        "player-self",
		HasCameraView: true,
		CameraView: runtime.CameraView{
			X: 10,
			Y: 100,
			Config: runtime.ClientConfig{
				VisibleWorldWidth:  100,
				VisibleWorldHeight: 100,
			},
		},
		Players: map[string]runtime.ShipState{
			"player-self": {ID: "player-self", X: 5000, Y: 5000},
			"player-near": {ID: "player-near", X: 200, Y: 100},
			"player-far":  {ID: "player-far", X: 3000, Y: 3000},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-wrapped": {ID: "asteroid-wrapped", X: 17190, Y: 100},
			"asteroid-far":     {ID: "asteroid-far", X: 3000, Y: 3000},
		},
		Bullets: map[string]runtime.BulletState{
			"bullet-near": {ID: "bullet-near", X: 500, Y: 100},
			"bullet-far":  {ID: "bullet-far", X: 3000, Y: 3000},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-near": {ID: "pickup-near", X: 200, Y: 100},
			"pickup-far":  {ID: "pickup-far", X: 3000, Y: 3000},
		},
	}

	filtered := applyNetworkInterest(snapshot, NewRealtimeSessionState("player-self", "match-1"), "")

	assertMapContains(t, filtered.Players, "player-self")
	assertMapContains(t, filtered.Players, "player-near")
	assertMapMissing(t, filtered.Players, "player-far")
	assertMapContains(t, filtered.Asteroids, "asteroid-wrapped")
	assertMapMissing(t, filtered.Asteroids, "asteroid-far")
	assertMapContains(t, filtered.Bullets, "bullet-near")
	assertMapMissing(t, filtered.Bullets, "bullet-far")
	assertMapContains(t, filtered.Pickups, "pickup-near")
	assertMapMissing(t, filtered.Pickups, "pickup-far")
}

func TestApplyNetworkInterestUsesExitMarginForPreviouslyRelevantEntity(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:        "player-self",
		HasCameraView: true,
		CameraView: runtime.CameraView{
			X: 0,
			Y: 0,
			Config: runtime.ClientConfig{
				VisibleWorldWidth:  100,
				VisibleWorldHeight: 100,
			},
		},
		Players: map[string]runtime.ShipState{
			"player-self": {ID: "player-self"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-edge": {ID: "asteroid-edge", X: 450, Y: 0},
		},
	}

	withoutHistory := applyNetworkInterest(snapshot, NewRealtimeSessionState("player-self", "match-1"), "")
	assertMapMissing(t, withoutHistory.Asteroids, "asteroid-edge")

	state := NewRealtimeSessionState("player-self", "match-1")
	state.StoreBaselineProjection(LaneWorld, WorldWireFullPacket{
		Asteroids: []WorldAsteroidWireRecord{{ID: "asteroid-edge"}},
	})
	withHistory := applyNetworkInterest(snapshot, state, "")
	assertMapContains(t, withHistory.Asteroids, "asteroid-edge")
}

func TestApplyNetworkInterestUsesSpectateTargetAsCameraAnchor(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:        "player-self",
		HasCameraView: true,
		CameraView: runtime.CameraView{
			X: 0,
			Y: 0,
			Config: runtime.ClientConfig{
				VisibleWorldWidth:  100,
				VisibleWorldHeight: 100,
			},
		},
		PlayerLocators: map[string]game.PlayerLocatorState{
			"player-target": {ID: "player-target", X: 4000, Y: 4000, Active: false},
		},
		Players: map[string]runtime.ShipState{
			"player-self":   {ID: "player-self", X: 0, Y: 0},
			"player-target": {ID: "player-target", X: 4000, Y: 4000},
			"player-near":   {ID: "player-near", X: 4100, Y: 4000},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-target-area": {ID: "asteroid-target-area", X: 4100, Y: 4000},
		},
	}

	filtered := applyNetworkInterest(snapshot, NewRealtimeSessionState("player-self", "match-1"), "player-target")
	assertMapContains(t, filtered.Players, "player-self")
	assertMapContains(t, filtered.Players, "player-target")
	assertMapContains(t, filtered.Players, "player-near")
	assertMapContains(t, filtered.Asteroids, "asteroid-target-area")
}

func TestApplyNetworkInterestUsesSelfLocatorWhenCameraConfigHasNotArrived(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-self",
		PlayerLocators: map[string]game.PlayerLocatorState{
			"player-self": {ID: "player-self", X: 2000, Y: 2000, Active: true},
		},
		Players: map[string]runtime.ShipState{
			"player-self": {ID: "player-self", X: 2000, Y: 2000},
			"player-near": {ID: "player-near", X: 2100, Y: 2000},
			"player-far":  {ID: "player-far", X: 6000, Y: 6000},
		},
	}

	filtered := applyNetworkInterest(snapshot, NewRealtimeSessionState("player-self", "match-1"), "")
	assertMapContains(t, filtered.Players, "player-self")
	assertMapContains(t, filtered.Players, "player-near")
	assertMapMissing(t, filtered.Players, "player-far")
}

func TestPlayerLocatorUsesShipsLaneWithoutAdvancingShipDeltaState(t *testing.T) {
	state := NewRealtimeSessionState("player-self", "match-1")
	state.HotLaneTick = 1
	snapshot := game.GameplayPresentationSnapshot{
		ServerSentMsec: 1234,
		PlayerLocators: map[string]game.PlayerLocatorState{
			"player-2": {ID: "player-2", X: 10, Y: 20, VelocityX: 3, VelocityY: 4, Active: true},
		},
	}

	candidates := buildPlayerLocatorCandidate(snapshot, state)
	if len(candidates) != 1 {
		t.Fatalf("locator candidates = %d, want 1", len(candidates))
	}
	candidate := candidates[0]
	if candidate.Lane() != LaneShips {
		t.Fatalf("locator lane = %q, want %q", candidate.Lane(), LaneShips)
	}
	if candidate.PacketFamily() != PacketFamilyPlayerLocator {
		t.Fatalf("locator family = %q", candidate.PacketFamily())
	}
	metadata, ok := candidate.Metadata()
	if !ok || metadata.Sequence != 1 {
		t.Fatalf("unexpected locator metadata: %#v", metadata)
	}
	payload := candidate.Payload.(PlayerLocatorPacket)
	if payload.Metadata.MatchID != "match-1" {
		t.Fatalf("locator payload match id = %q", payload.Metadata.MatchID)
	}

	CommitSuccessfulCandidate(&state, candidate)
	if _, ok := state.LaneState(LaneShips); ok {
		t.Fatal("locator write advanced ship-delta lane state")
	}
	if state.PacketSequence(PacketFamilyPlayerLocator) != 1 {
		t.Fatalf("locator sequence = %d, want 1", state.PacketSequence(PacketFamilyPlayerLocator))
	}
	if _, ok := state.PacketProjection(PacketFamilyPlayerLocator); !ok {
		t.Fatal("locator projection was not committed")
	}

	state.HotLaneTick = 2
	if got := buildPlayerLocatorCandidate(snapshot, state); len(got) != 0 {
		t.Fatalf("locator emitted before cadence: %d candidates", len(got))
	}
}

func assertMapContains[T any](t *testing.T, values map[string]T, id string) {
	t.Helper()
	if _, ok := values[id]; !ok {
		t.Fatalf("expected %q to be present", id)
	}
}

func assertMapMissing[T any](t *testing.T, values map[string]T, id string) {
	t.Helper()
	if _, ok := values[id]; ok {
		t.Fatalf("expected %q to be absent", id)
	}
}
