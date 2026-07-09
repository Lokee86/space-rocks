package devtools

import "sync"

type ObserverRegistry struct {
	mu         sync.Mutex
	registered map[any]struct{}
}

func NewObserverRegistry() *ObserverRegistry {
	return &ObserverRegistry{
		registered: make(map[any]struct{}),
	}
}

func (registry *ObserverRegistry) RegisterOnce(target StreamTarget, observer func(float64)) {
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
