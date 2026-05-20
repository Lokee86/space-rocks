package devtools

type WorldOptions struct {
	FreezeAsteroids  bool
	FreezeBullets    bool
	FreezeSpawning   bool
	FreezeCollisions bool
}

func (options *WorldOptions) ToggleFreezeWorld() bool {
	enabled := !options.IsWorldFrozen()
	options.FreezeAsteroids = enabled
	options.FreezeBullets = enabled
	options.FreezeSpawning = enabled
	options.FreezeCollisions = enabled
	return enabled
}

func (options WorldOptions) IsWorldFrozen() bool {
	return options.FreezeAsteroids &&
		options.FreezeBullets &&
		options.FreezeSpawning &&
		options.FreezeCollisions
}

func (options WorldOptions) AsteroidsCanMove() bool {
	return !options.FreezeAsteroids
}

func (options WorldOptions) BulletsCanMove() bool {
	return !options.FreezeBullets
}

func (options WorldOptions) CanSpawnAsteroids() bool {
	return !options.FreezeSpawning
}

func (options WorldOptions) CanRunCollisions() bool {
	return !options.FreezeCollisions
}
