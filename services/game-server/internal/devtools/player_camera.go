package devtools

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"

const DummyPlayerVisibleWorldWidth = runtime.BaseVisibleWorldWidth
const DummyPlayerVisibleWorldHeight = runtime.BaseVisibleWorldHeight

func DummyPlayerCameraConfig() runtime.ClientConfig {
	return runtime.DefaultCameraConfig()
}
