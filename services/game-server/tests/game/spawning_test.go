package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
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
