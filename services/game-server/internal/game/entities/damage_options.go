package entities

type DamageOptions struct {
	Invincible bool
}

func (options *DamageOptions) SetInvincible(invincible bool) {
	options.Invincible = invincible
}

func (options DamageOptions) CanTakeDamage() bool {
	return !options.Invincible
}
