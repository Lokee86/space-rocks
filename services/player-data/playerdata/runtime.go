package playerdata

import "errors"

type Config struct {
	Store Store
}

type Runtime struct {
	dispatcher *Dispatcher
}

func NewRuntime(config Config) (*Runtime, error) {
	if config.Store == nil {
		return nil, errors.New("store is required")
	}

	return &Runtime{
		dispatcher: NewDispatcher(config.Store),
	}, nil
}

func (r *Runtime) Handle(payload []byte) ([]byte, error) {
	return r.dispatcher.Handle(payload)
}
