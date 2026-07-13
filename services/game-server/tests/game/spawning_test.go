package gametests

import (
	"math/rand"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func TestAsteroidSpawningUsesClientCameraView(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{
		Type: servergame.PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  200,
			VisibleWorldHeight: 200,
		},
	})
	scenario.step(constants.AsteroidSpawnInterval)

	snapshot := scenario.presentationSnapshot(playerID)
	if len(snapshot.Asteroids) != constants.AsteroidSpawnBatchSize {
		t.Fatalf("expected %d asteroids spawned for camera view, got %d", constants.AsteroidSpawnBatchSize, len(snapshot.Asteroids))
	}
}

func TestAsteroidSpawningNearBoundaryStoresWrappedPosition(t *testing.T) {
	rand.Seed(1)
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.setPlayerPosition(playerID, physics.Vector2{
		X: constants.WorldWidth - 1,
		Y: constants.WorldHeight - 1,
	})
	scenario.send(playerID, servergame.ClientPacket{
		Type: servergame.PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  200,
			VisibleWorldHeight: 200,
		},
	})

	scenario.step(constants.AsteroidSpawnInterval)

	snapshot := scenario.presentationSnapshot(playerID)
	for id, asteroid := range snapshot.Asteroids {
		if asteroid.X < 0 || asteroid.X >= constants.WorldWidth ||
			asteroid.Y < 0 || asteroid.Y >= constants.WorldHeight {
			t.Fatalf("expected asteroid %s to be stored inside world bounds, got (%v, %v)", id, asteroid.X, asteroid.Y)
		}
	}
}

func TestAsteroidStateIncludesResolvedScale(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	const asteroidSize = 3
	asteroid := scenario.control.ApplyAsteroidSpawnPlan(spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   physics.Vector2{X: 100, Y: 100},
		Size:       asteroidSize,
	})

	asteroidSnapshot, ok := scenario.asteroidSnapshot(playerID, asteroid.ID)
	if !ok {
		t.Fatal("expected gameplay snapshot world projection to include asteroid")
	}
	if asteroidSnapshot.Size != asteroidSize {
		t.Fatalf("expected asteroid size %d, got %d", asteroidSize, asteroidSnapshot.Size)
	}
	if asteroidSnapshot.Health != constants.AsteroidHealth {
		t.Fatalf("expected asteroid health %d, got %d", constants.AsteroidHealth, asteroidSnapshot.Health)
	}

	expectedScale := float64(asteroidSize) * constants.AsteroidSizeScale
	if asteroidSnapshot.Scale != expectedScale {
		t.Fatalf("expected asteroid scale %v, got %v", expectedScale, asteroidSnapshot.Scale)
	}
}
