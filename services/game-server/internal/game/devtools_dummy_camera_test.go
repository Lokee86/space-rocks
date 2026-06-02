package game

import "github.com/Lokee86/space-rocks/server/internal/game/entities"

const (
	DummyPlayerVisibleWorldWidth  = 1920
	DummyPlayerVisibleWorldHeight = 1080
)

func DummyPlayerCameraConfig() entities.ClientConfig {
	return entities.ClientConfig{
		VisibleWorldWidth:  DummyPlayerVisibleWorldWidth,
		VisibleWorldHeight: DummyPlayerVisibleWorldHeight,
	}
}
