package devtools

import "github.com/Lokee86/space-rocks/server/internal/game/entities"

const DummyPlayerVisibleWorldWidth = 1280
const DummyPlayerVisibleWorldHeight = 720

func DummyPlayerCameraConfig() entities.ClientConfig {
	return entities.ClientConfig{
		VisibleWorldWidth:  DummyPlayerVisibleWorldWidth,
		VisibleWorldHeight: DummyPlayerVisibleWorldHeight,
	}
}
