package realtime

import (
	"math"
	"strings"
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestBuildActiveRealtimeResultSurfacesWorldQuantizationFailure(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", X: math.NaN()}},
	}
	assertActiveQuantizationFailure(t, snapshot, "world")
}

func TestAssembleRealtimeLaneCandidatesReturnsWorldQuantizationFailure(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", X: math.NaN()}},
	}
	_, err := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1", "match-1"))
	if err == nil {
		t.Fatal("expected exported planner to return world quantization failure")
	}
	if !strings.Contains(err.Error(), "quantize world full packet") {
		t.Fatalf("error = %q, want world quantization context", err)
	}
}

func TestBuildActiveRealtimeResultSurfacesOverlayQuantizationFailure(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", RespawnCooldown: math.NaN()}},
	}
	assertActiveQuantizationFailure(t, snapshot, "overlay")
}

func TestBuildActiveRealtimeResultSurfacesSessionQuantizationFailure(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", SpawnX: math.NaN()}},
	}
	assertActiveQuantizationFailure(t, snapshot, "session")
}

func assertActiveQuantizationFailure(t *testing.T, snapshot game.GameplayPresentationSnapshot, lane string) {
	t.Helper()
	_, err := BuildActiveRealtimeResult(snapshot, NewRealtimeSessionState("player-1", "match-1"))
	if err == nil {
		t.Fatalf("expected %s quantization failure", lane)
	}
	if !strings.Contains(err.Error(), "quantize "+lane+" full packet") {
		t.Fatalf("error = %q, want %s quantization context", err, lane)
	}
}
