package playerdata

import "errors"

type RuntimeSink struct {
	runtime *Runtime
}

func NewRuntimeSink(runtime *Runtime) *RuntimeSink {
	return &RuntimeSink{runtime: runtime}
}

func (s *RuntimeSink) HandlePlayerDataCommand(payload []byte) ([]byte, error) {
	if s.runtime == nil {
		return nil, errors.New("player-data runtime is required")
	}
	return s.runtime.Handle(payload)
}
