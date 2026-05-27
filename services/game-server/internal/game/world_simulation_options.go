package game

type WorldSimulationOptions struct {
	FreezeWorld bool
}

func (options *WorldSimulationOptions) SetFreezeWorld(frozen bool) {
	options.FreezeWorld = frozen
}

func (options WorldSimulationOptions) IsWorldFrozen() bool {
	return options.FreezeWorld
}

func (options WorldSimulationOptions) AsteroidsCanMove() bool {
	return !options.FreezeWorld
}

func (options WorldSimulationOptions) BulletsCanMove() bool {
	return !options.FreezeWorld
}

func (options WorldSimulationOptions) CanSpawnAsteroids() bool {
	return !options.FreezeWorld
}

func (options WorldSimulationOptions) CanRunCollisions() bool {
	return !options.FreezeWorld
}
