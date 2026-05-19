package devtools

type PlayerOptions struct {
	Invincible    bool
	InfiniteLives bool
}

func (options *PlayerOptions) ToggleInvincible() bool {
	options.Invincible = !options.Invincible
	return options.Invincible
}

func (options *PlayerOptions) ToggleInfiniteLives() bool {
	options.InfiniteLives = !options.InfiniteLives
	return options.InfiniteLives
}

func (options PlayerOptions) CanTakeDamage() bool {
	return !options.Invincible
}

func (options PlayerOptions) CanLoseLives() bool {
	return !options.InfiniteLives
}
