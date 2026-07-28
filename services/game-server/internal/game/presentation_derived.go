package game

// presentationDerivedEntry holds immutable data derived from one published
// presentation generation. Values are owned by the caller and must not be
// mutated after publication.
type presentationDerivedEntry struct {
	generation uint64
	value      any
}

// PresentationDerived returns one cached immutable value for a published
// presentation generation. Concurrent receivers share the first successful
// build instead of repeating generation-wide projection work. The current and
// immediately previous generations are retained because receiver write loops
// can briefly overlap adjacent published frames.
func (game *Game) PresentationDerived(generation uint64, key string, build func() (any, error)) (any, error) {
	game.presentationDerivedMu.Lock()
	defer game.presentationDerivedMu.Unlock()

	entries := game.presentationDerived[key]
	for _, entry := range entries {
		if entry.generation == generation {
			return entry.value, nil
		}
	}

	value, err := build()
	if err != nil {
		return nil, err
	}
	entries = append(entries, presentationDerivedEntry{
		generation: generation,
		value:      value,
	})
	if len(entries) > 2 {
		oldest := 0
		for index := 1; index < len(entries); index++ {
			if entries[index].generation < entries[oldest].generation {
				oldest = index
			}
		}
		entries = append(entries[:oldest], entries[oldest+1:]...)
	}
	game.presentationDerived[key] = entries
	return value, nil
}
