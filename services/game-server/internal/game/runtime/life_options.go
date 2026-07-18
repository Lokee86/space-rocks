package runtime

type LifeOptions struct {
	InfiniteLives bool
}

func (options *LifeOptions) SetInfiniteLives(infinite bool) {
	options.InfiniteLives = infinite
}
