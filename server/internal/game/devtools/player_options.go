package devtools

type PlayerOptions struct {
	Invincible bool
}

func (options *PlayerOptions) ToggleInvincible() bool {
	options.Invincible = !options.Invincible
	return options.Invincible
}

func (options PlayerOptions) CanTakeDamage() bool {
	return !options.Invincible
}
