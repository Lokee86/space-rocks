package physics

import "math"

const collisionOutlineCircleSegments = 24

func CollisionBodyOutlinePoints(body CollisionBody) []Vector2 {
	switch body.Shape.Type {
	case CollisionShapeCircle:
		return circleOutlinePoints(body)
	case CollisionShapeCapsule:
		return capsuleOutlinePoints(body)
	case CollisionShapeRectangle:
		return polygonPoints(body)
	case CollisionShapePolygon:
		return polygonPoints(body)
	default:
		return nil
	}
}

func circleOutlinePoints(body CollisionBody) []Vector2 {
	points := make([]Vector2, 0, collisionOutlineCircleSegments)
	for index := 0; index < collisionOutlineCircleSegments; index++ {
		angle := float64(index) * 2 * math.Pi / float64(collisionOutlineCircleSegments)
		offset := Vector2{X: math.Cos(angle), Y: math.Sin(angle)}.Multiply(body.Shape.Radius)
		points = append(points, body.Position.Add(offset))
	}

	return points
}

func capsuleOutlinePoints(body CollisionBody) []Vector2 {
	radius := body.Shape.Radius
	segmentLength := math.Max(0, body.Shape.Height-2*radius)
	halfSegment := segmentLength * 0.5
	topCenter := body.Position.Add(Vector2{Y: halfSegment}.Rotated(body.Rotation))
	bottomCenter := body.Position.Subtract(Vector2{Y: halfSegment}.Rotated(body.Rotation))

	points := make([]Vector2, 0, collisionOutlineCircleSegments)
	segmentsPerCap := collisionOutlineCircleSegments / 2
	if segmentsPerCap < 1 {
		segmentsPerCap = 1
	}

	for index := 0; index <= segmentsPerCap; index++ {
		angle := math.Pi + float64(index)*math.Pi/float64(segmentsPerCap)
		local := Vector2{X: math.Cos(angle) * radius, Y: math.Sin(angle) * radius}
		points = append(points, topCenter.Add(local.Rotated(body.Rotation)))
	}

	for index := 0; index <= segmentsPerCap; index++ {
		angle := float64(index) * math.Pi / float64(segmentsPerCap)
		local := Vector2{X: math.Cos(angle) * radius, Y: math.Sin(angle) * radius}
		points = append(points, bottomCenter.Add(local.Rotated(body.Rotation)))
	}

	return points
}
