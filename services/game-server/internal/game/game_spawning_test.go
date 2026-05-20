package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
)

func TestAsteroidSpawningUsesCameraViewsWithoutPlayer(t *testing.T) {
	game := New()
	game.cameraViews["player-1"] = &entities.CameraView{
		X: 100,
		Y: 100,
		Config: entities.ClientConfig{
			VisibleWorldWidth:  200,
			VisibleWorldHeight: 200,
		},
	}

	game.Step(constants.AsteroidSpawnInterval)

	if len(game.state.Asteroids) != constants.AsteroidSpawnBatchSize {
		t.Fatalf("expected %d asteroids spawned for camera view, got %d", constants.AsteroidSpawnBatchSize, len(game.state.Asteroids))
	}
}
