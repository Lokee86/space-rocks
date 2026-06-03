package streamruntime

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

func StepContinuousBulletStreams(delta float64, bulletsCanMove bool, spawn func(string, physics.Vector2, physics.Vector2) bool) {
	DefaultRuntime.StepContinuousBulletStreams(delta, bulletsCanMove, spawn)
}
