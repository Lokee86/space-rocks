package devtools

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

type ObserverRegistry struct {
	mu         sync.Mutex
	registered map[any]struct{}
}

func NewObserverRegistry() *ObserverRegistry {
	return &ObserverRegistry{
		registered: make(map[any]struct{}),
	}
}

func (registry *ObserverRegistry) RegisterOnce(target StreamTarget, observer func(float64, func() bool, func(string, physics.Vector2, physics.Vector2) bool)) {
	if registry == nil || target == nil || observer == nil {
		return
	}

	key := target.ObserverKey()
	registry.mu.Lock()
	if registry.registered == nil {
		registry.registered = make(map[any]struct{})
	}
	if _, exists := registry.registered[key]; exists {
		registry.mu.Unlock()
		return
	}
	registry.registered[key] = struct{}{}
	registry.mu.Unlock()

	target.RegisterSimulationStepObserver(observer)
}
