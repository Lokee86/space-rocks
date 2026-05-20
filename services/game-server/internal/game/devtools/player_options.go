package devtools

type PlayerOptions struct {
	Invincible    bool
	InfiniteLives bool
	FreezePlayer  bool
}

func (options *PlayerOptions) ToggleInvincible() bool {
	options.Invincible = !options.Invincible
	return options.Invincible
}

func (options *PlayerOptions) ToggleInfiniteLives() bool {
	options.InfiniteLives = !options.InfiniteLives
	return options.InfiniteLives
}

func (options *PlayerOptions) ToggleFreezePlayer() bool {
	options.FreezePlayer = !options.FreezePlayer
	return options.FreezePlayer
}

func (options PlayerOptions) IsPlayerFrozen() bool {
	return options.FreezePlayer
}

func (options PlayerOptions) CanTakeDamage() bool {
	return !options.Invincible
}

func (options PlayerOptions) CanLoseLives() bool {
	return !options.InfiniteLives
}
