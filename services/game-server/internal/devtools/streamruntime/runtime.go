package streamruntime

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type Runtime struct {
	continuousBulletStreams *ContinuousBulletStreams
}

var DefaultRuntime = NewRuntime()

func NewRuntime() *Runtime {
	return &Runtime{
		continuousBulletStreams: NewContinuousBulletStreams(),
	}
}

func (runtime *Runtime) BeginContinuousBulletStream(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	return runtime.continuousBulletStreams.Begin(ownerPlayerID, origin, direction)
}

func (runtime *Runtime) ActiveContinuousBulletStreams() []ContinuousBulletStream {
	return runtime.continuousBulletStreams.Active()
}

func (runtime *Runtime) ClearContinuousBulletStreams() {
	runtime.continuousBulletStreams.Clear()
}

func (runtime *Runtime) StepContinuousBulletStreams(delta float64, bulletsCanMove bool, spawn func(string, physics.Vector2, physics.Vector2) bool) {
	runtime.continuousBulletStreams.Step(delta, bulletsCanMove, spawn)
}
