package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestAsteroidSpawningUsesClientCameraView(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{
		Type: servergame.PacketTypeClientConfig,
		Config: entities.ClientConfig{
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
