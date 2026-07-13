package grid

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func (index *Index) QueryCircle(dst []spatial.Ref, center physics.Vector2, radius float64, mask spatial.KindMask) []spatial.Ref {
	if mask == 0 || radius < 0 {
		return dst
	}

	center = space.WrapPosition(center, index.bounds)
	xCells := index.axisRange(center.X-radius, center.X+radius, index.bounds.Width, index.cellWidth, index.cellsX)
	yCells := index.axisRange(center.Y-radius, center.Y+radius, index.bounds.Height, index.cellHeight, index.cellsY)
	generation := index.beginQuery()
	for yOffset := 0; yOffset < yCells.count; yOffset++ {
		y := (yCells.start + yOffset) % index.cellsY
		for xOffset := 0; xOffset < xCells.count; xOffset++ {
			x := (xCells.start + xOffset) % index.cellsX
			for _, entry := range index.buckets[y*index.cellsX+x] {
				if !mask.Allows(entry.Ref.Kind) || !index.markQueryRef(entry.Ref, generation) {
					continue
				}
				delta := space.ShortestDelta(center, entry.Position, index.bounds)
				combinedRadius := radius + entry.Radius
				if delta.LengthSquared() <= combinedRadius*combinedRadius {
					dst = append(dst, entry.Ref)
				}
			}
		}
	}
	return dst
}