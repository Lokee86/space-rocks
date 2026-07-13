package grid

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"

func (index *Index) beginQuery() uint64 {
	index.queryGeneration++
	if index.queryGeneration == 0 {
		clear(index.querySeen)
		index.queryGeneration = 1
	}
	return index.queryGeneration
}

func (index *Index) markQueryRef(ref spatial.Ref, generation uint64) bool {
	if index.querySeen[ref] == generation {
		return false
	}
	index.querySeen[ref] = generation
	return true
}