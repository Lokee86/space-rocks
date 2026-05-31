package game

type WorldSimulationOptions struct {
	FreezeAsteroids  bool
	FreezeBullets    bool
	FreezeSpawning   bool
	FreezeCollisions bool
}

func (options *WorldSimulationOptions) SetFreezeWorld(frozen bool) {
	options.FreezeAsteroids = frozen
	options.FreezeBullets = frozen
	options.FreezeSpawning = frozen
	options.FreezeCollisions = frozen
}

func (options WorldSimulationOptions) IsWorldFrozen() bool {
	return options.FreezeAsteroids &&
		options.FreezeBullets &&
		options.FreezeSpawning &&
		options.FreezeCollisions
}

func (options WorldSimulationOptions) AsteroidsCanMove() bool {
	return !options.FreezeAsteroids
}

func (options WorldSimulationOptions) BulletsCanMove() bool {
	return !options.FreezeBullets
}

func (options WorldSimulationOptions) CanSpawnAsteroids() bool {
	return !options.FreezeSpawning
}

func (options WorldSimulationOptions) CanRunCollisions() bool {
	return !options.FreezeCollisions
}

func (options *WorldSimulationOptions) ToggleFreezeWorld() bool {
	enabled := !options.IsWorldFrozen()
	options.SetFreezeWorld(enabled)
	return enabled
}

func (options *WorldSimulationOptions) ToggleFreezeAsteroids() bool {
	options.FreezeAsteroids = !options.FreezeAsteroids
	return options.FreezeAsteroids
}

func (options *WorldSimulationOptions) ToggleFreezeBullets() bool {
	options.FreezeBullets = !options.FreezeBullets
	return options.FreezeBullets
}

func (options *WorldSimulationOptions) ToggleFreezeSpawning() bool {
	options.FreezeSpawning = !options.FreezeSpawning
	return options.FreezeSpawning
}

func (options *WorldSimulationOptions) ToggleFreezeCollisions() bool {
	options.FreezeCollisions = !options.FreezeCollisions
	return options.FreezeCollisions
}
