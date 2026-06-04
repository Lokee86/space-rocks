package game

import "github.com/Lokee86/space-rocks/server/internal/game/runtime"

const (
	DummyPlayerVisibleWorldWidth  = 1920
	DummyPlayerVisibleWorldHeight = 1080
)

func DummyPlayerCameraConfig() runtime.ClientConfig {
	return runtime.ClientConfig{
		VisibleWorldWidth:  DummyPlayerVisibleWorldWidth,
		VisibleWorldHeight: DummyPlayerVisibleWorldHeight,
	}
}
