package gametests

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func seededAsteroidSnapshot(t *testing.T, seed int64) map[string]runtime.AsteroidState {
	t.Helper()

	scenario := newScenarioWithSeed(t, seed)
	playerID := scenario.addPlayer()
	scenario.send(playerID, servergame.ClientPacket{
		Type: servergame.PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  200,
			VisibleWorldHeight: 200,
		},
	})

	scenario.step(constants.AsteroidSpawnInterval)
	return scenario.presentationSnapshot(playerID).Asteroids
}

func TestAsteroidSpawnDeterminismSameSeedProducesIdenticalAsteroidStates(t *testing.T) {
	first := seededAsteroidSnapshot(t, 1)
	second := seededAsteroidSnapshot(t, 1)

	if len(first) == 0 || len(second) == 0 {
		t.Fatalf("expected seeded asteroid snapshots to be non-empty, got %d and %d", len(first), len(second))
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected identical asteroid states for same seed, got first=%v second=%v", first, second)
	}
}

func TestAsteroidSpawnDeterminismDifferentSeedsProduceDifferentAsteroidStates(t *testing.T) {
	first := seededAsteroidSnapshot(t, 1)
	second := seededAsteroidSnapshot(t, 2)

	if len(first) == 0 || len(second) == 0 {
		t.Fatalf("expected seeded asteroid snapshots to be non-empty, got %d and %d", len(first), len(second))
	}
	if reflect.DeepEqual(first, second) {
		t.Fatalf("expected different asteroid states for different seeds, got first=%v second=%v", first, second)
	}
}
