package devtools

import "github.com/Lokee86/space-rocks/server/internal/game/runtime"

const DummyPlayerVisibleWorldWidth = 1280
const DummyPlayerVisibleWorldHeight = 720

func DummyPlayerCameraConfig() runtime.ClientConfig {
	return runtime.ClientConfig{
		VisibleWorldWidth:  DummyPlayerVisibleWorldWidth,
		VisibleWorldHeight: DummyPlayerVisibleWorldHeight,
	}
}
