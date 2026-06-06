package weapons

type SlotState struct {
	CooldownRemaining float64
	AmmoRemaining     int
}

type State struct {
	Primary   SlotState
	Secondary SlotState
}

func StepSlotState(state SlotState, delta float64) SlotState {
	state.CooldownRemaining -= delta
	if state.CooldownRemaining < 0 {
		state.CooldownRemaining = 0
	}
	return state
}

func StepState(state State, delta float64) State {
	state.Primary = StepSlotState(state.Primary, delta)
	state.Secondary = StepSlotState(state.Secondary, delta)
	return state
}
