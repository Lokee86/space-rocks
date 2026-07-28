package devtools

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

const continuousBulletStreamObserverKey = "devtools.continuous_bullet_streams"

// ObserverRegistry keeps the registration seam while leaving observer ownership
// with each target. It must not retain target keys across match lifetimes.
type ObserverRegistry struct{}

func NewObserverRegistry() *ObserverRegistry {
	return &ObserverRegistry{}
}

func (registry *ObserverRegistry) RegisterOnce(target StreamTarget, observer func(float64, func() bool, func(string, physics.Vector2, physics.Vector2) bool)) {
	if registry == nil || target == nil || observer == nil {
		return
	}
	target.RegisterSimulationStepObserverOnce(continuousBulletStreamObserverKey, observer)
}
