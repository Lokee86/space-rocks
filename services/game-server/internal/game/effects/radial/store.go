package radial

type Store struct {
	effects map[string]*Effect
}

func NewStore() Store {
	return Store{
		effects: make(map[string]*Effect),
	}
}

func (s Store) Add(effect Effect) {
	if s.effects == nil {
		s.effects = make(map[string]*Effect)
	}

	effectCopy := effect
	s.effects[effect.ID] = &effectCopy
}

func (s Store) All() map[string]*Effect {
	return s.effects
}

func (s Store) Remove(id string) {
	delete(s.effects, id)
}

func (s Store) Len() int {
	return len(s.effects)
}
