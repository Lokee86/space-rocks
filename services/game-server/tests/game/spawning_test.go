package gametests

import (
	"math/rand"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
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

	packet := scenario.state(playerID)
	if len(packet.Asteroids) != constants.AsteroidSpawnBatchSize {
		t.Fatalf("expected %d asteroids spawned for camera view, got %d", constants.AsteroidSpawnBatchSize, len(packet.Asteroids))
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

	for id, asteroid := range scenario.state(playerID).Asteroids {
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
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: 100, Y: 100}, asteroidSize)

	asteroid, ok := scenario.state(playerID).Asteroids["asteroid-1"]
	if !ok {
		t.Fatal("expected state packet to include asteroid")
	}
	if asteroid.Size != asteroidSize {
		t.Fatalf("expected asteroid size %d, got %d", asteroidSize, asteroid.Size)
	}

	expectedScale := float64(asteroidSize) * constants.AsteroidSizeScale
	if asteroid.Scale != expectedScale {
		t.Fatalf("expected asteroid scale %v, got %v", expectedScale, asteroid.Scale)
	}
}
