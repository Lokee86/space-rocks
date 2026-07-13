package grid

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func (index *Index) insert(entry spatial.Entry) {
	xCells := index.axisRange(entry.Position.X-entry.Radius, entry.Position.X+entry.Radius, index.bounds.Width, index.cellWidth, index.cellsX)
	yCells := index.axisRange(entry.Position.Y-entry.Radius, entry.Position.Y+entry.Radius, index.bounds.Height, index.cellHeight, index.cellsY)

	for yOffset := 0; yOffset < yCells.count; yOffset++ {
		y := (yCells.start + yOffset) % index.cellsY
		for xOffset := 0; xOffset < xCells.count; xOffset++ {
			x := (xCells.start + xOffset) % index.cellsX
			bucketIndex := y*index.cellsX + x
			if len(index.buckets[bucketIndex]) == 0 {
				index.touched = append(index.touched, bucketIndex)
			}
			index.buckets[bucketIndex] = append(index.buckets[bucketIndex], entry)
		}
	}
}

type axisRange struct {
	start int
	count int
}

func (index *Index) axisRange(minimum, maximum, size, cellExtent float64, count int) axisRange {
	if maximum-minimum >= size {
		return axisRange{count: count}
	}

	first := int(math.Floor(minimum / cellExtent))
	last := int(math.Floor(maximum / cellExtent))
	span := last - first + 1
	if span > count {
		span = count
	}
	start := first % count
	if start < 0 {
		start += count
	}
	return axisRange{start: start, count: span}
}
