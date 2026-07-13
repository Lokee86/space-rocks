package grid

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

type Index struct {
	bounds          space.Bounds
	cellWidth       float64
	cellHeight      float64
	cellsX          int
	cellsY          int
	buckets         [][]spatial.Entry
	touched         []int
	querySeen       map[spatial.Ref]uint64
	queryGeneration uint64
}

var _ spatial.Index = (*Index)(nil)

func New(bounds space.Bounds, cellSize float64) *Index {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		panic("spatial grid requires positive bounds")
	}
	if cellSize <= 0 {
		panic("spatial grid requires a positive cell size")
	}

	cellsX := int(math.Ceil(bounds.Width / cellSize))
	cellsY := int(math.Ceil(bounds.Height / cellSize))
	return &Index{
		bounds:     bounds,
		cellWidth:  bounds.Width / float64(cellsX),
		cellHeight: bounds.Height / float64(cellsY),
		cellsX:     cellsX,
		cellsY:     cellsY,
		buckets:    make([][]spatial.Entry, cellsX*cellsY),
		querySeen:  make(map[spatial.Ref]uint64),
	}
}

func (index *Index) Rebuild(entries []spatial.Entry) {
	clear(index.querySeen)
	index.queryGeneration = 0

	for _, bucketIndex := range index.touched {
		index.buckets[bucketIndex] = index.buckets[bucketIndex][:0]
	}
	index.touched = index.touched[:0]

	for _, entry := range entries {
		entry.Position = space.WrapPosition(entry.Position, index.bounds)
		if entry.Radius < 0 {
			entry.Radius = 0
		}
		index.insert(entry)
	}
}

