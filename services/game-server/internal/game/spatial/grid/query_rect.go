package grid

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func (index *Index) QueryRect(dst []spatial.Ref, rect spatial.Rect, mask spatial.KindMask) []spatial.Ref {
	if mask == 0 {
		return dst
	}

	rect.Center = space.WrapPosition(rect.Center, index.bounds)
	rect.HalfExtents.X = math.Max(0, rect.HalfExtents.X)
	rect.HalfExtents.Y = math.Max(0, rect.HalfExtents.Y)
	xCells := index.axisRange(rect.Center.X-rect.HalfExtents.X, rect.Center.X+rect.HalfExtents.X, index.bounds.Width, index.cellWidth, index.cellsX)
	yCells := index.axisRange(rect.Center.Y-rect.HalfExtents.Y, rect.Center.Y+rect.HalfExtents.Y, index.bounds.Height, index.cellHeight, index.cellsY)
	generation := index.beginQuery()

	for yOffset := 0; yOffset < yCells.count; yOffset++ {
		y := (yCells.start + yOffset) % index.cellsY
		for xOffset := 0; xOffset < xCells.count; xOffset++ {
			x := (xCells.start + xOffset) % index.cellsX
			for _, entry := range index.buckets[y*index.cellsX+x] {
				if !mask.Allows(entry.Ref.Kind) || !index.markQueryRef(entry.Ref, generation) {
					continue
				}
				if circleOverlapsRect(rect, entry.Position, entry.Radius, index.bounds) {
					dst = append(dst, entry.Ref)
				}
			}
		}
	}
	return dst
}

func circleOverlapsRect(rect spatial.Rect, position physics.Vector2, radius float64, bounds space.Bounds) bool {
	delta := space.ShortestDelta(rect.Center, position, bounds)
	distanceX := math.Max(0, math.Abs(delta.X)-rect.HalfExtents.X)
	distanceY := math.Max(0, math.Abs(delta.Y)-rect.HalfExtents.Y)
	return distanceX*distanceX+distanceY*distanceY <= radius*radius
}
