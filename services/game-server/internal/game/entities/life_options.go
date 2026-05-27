package entities

type LifeOptions struct {
	InfiniteLives bool
}

func (options *LifeOptions) SetInfiniteLives(infinite bool) {
	options.InfiniteLives = infinite
}

func (options LifeOptions) CanLoseLives() bool {
	return !options.InfiniteLives
}
