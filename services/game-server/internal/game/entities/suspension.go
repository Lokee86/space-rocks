package entities

type SuspensionState struct {
	Paused    bool
	DevFrozen bool
}

func (state SuspensionState) IsSuspended() bool {
	return state.Paused || state.DevFrozen
}

func (state *SuspensionState) SetPaused(paused bool) {
	state.Paused = paused
}

func (state *SuspensionState) SetDevFrozen(frozen bool) {
	state.DevFrozen = frozen
}
