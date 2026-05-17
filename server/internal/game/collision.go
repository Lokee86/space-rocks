package game

import "math"

type CollisionShapeType string

const (
	CollisionShapeCircle    CollisionShapeType = "circle"
	CollisionShapeCapsule   CollisionShapeType = "capsule"
	CollisionShapeRectangle CollisionShapeType = "rectangle"
	CollisionShapePolygon   CollisionShapeType = "polygon"
)

type CollisionShape struct {
	Type   CollisionShapeType
	Radius float64
	Height float64
	Size   Vector2
	Points []Vector2
}

type CollisionBody struct {
	ID       string
	Position Vector2
	Rotation float64
	Shape    CollisionShape
}

type Collision struct {
	A            CollisionBody
	B            CollisionBody
	ContactPoint Vector2
}

func NewCircleShape(radius float64) CollisionShape {
	return CollisionShape{Type: CollisionShapeCircle, Radius: radius}
}

func NewCapsuleShape(radius float64, height float64) CollisionShape {
	return CollisionShape{Type: CollisionShapeCapsule, Radius: radius, Height: height}
}

func NewRectangleShape(width float64, height float64) CollisionShape {
	return CollisionShape{Type: CollisionShapeRectangle, Size: Vector2{X: width, Y: height}}
}

func NewPolygonShape(points []Vector2) CollisionShape {
	return CollisionShape{Type: CollisionShapePolygon, Points: points}
}

func DetectCollision(a CollisionBody, b CollisionBody) (Collision, bool) {
	ok := shapesIntersect(a, b)
	if !ok {
		return Collision{}, false
	}

	return Collision{
		A:            a,
		B:            b,
		ContactPoint: a.Position.add(b.Position).multiply(0.5),
	}, true
}

func shapesIntersect(a CollisionBody, b CollisionBody) bool {
	aKind := primitiveKind(a.Shape.Type)
	bKind := primitiveKind(b.Shape.Type)

	switch {
	case aKind == "circle" && bKind == "circle":
		return circlesIntersect(circlePrimitive(a), circlePrimitive(b))
	case aKind == "capsule" && bKind == "capsule":
		return capsulesIntersect(capsulePrimitive(a), capsulePrimitive(b))
	case aKind == "circle" && bKind == "capsule":
		return circleCapsuleIntersects(circlePrimitive(a), capsulePrimitive(b))
	case aKind == "capsule" && bKind == "circle":
		return circleCapsuleIntersects(circlePrimitive(b), capsulePrimitive(a))
	case aKind == "polygon" && bKind == "polygon":
		return polygonsIntersect(polygonPoints(a), polygonPoints(b))
	case aKind == "circle" && bKind == "polygon":
		return circlePolygonIntersects(circlePrimitive(a), polygonPoints(b))
	case aKind == "polygon" && bKind == "circle":
		return circlePolygonIntersects(circlePrimitive(b), polygonPoints(a))
	case aKind == "capsule" && bKind == "polygon":
		return capsulePolygonIntersects(capsulePrimitive(a), polygonPoints(b))
	case aKind == "polygon" && bKind == "capsule":
		return capsulePolygonIntersects(capsulePrimitive(b), polygonPoints(a))
	default:
		return false
	}
}

func primitiveKind(shapeType CollisionShapeType) string {
	if shapeType == CollisionShapeRectangle || shapeType == CollisionShapePolygon {
		return "polygon"
	}

	return string(shapeType)
}

type collisionCircle struct {
	Center Vector2
	Radius float64
}

type collisionCapsule struct {
	Start  Vector2
	End    Vector2
	Radius float64
}

func circlePrimitive(body CollisionBody) collisionCircle {
	return collisionCircle{
		Center: body.Position,
		Radius: body.Shape.Radius,
	}
}

func capsulePrimitive(body CollisionBody) collisionCapsule {
	segmentLength := math.Max(0, body.Shape.Height-2*body.Shape.Radius)
	offset := Vector2{Y: segmentLength * 0.5}.rotated(body.Rotation)
	return collisionCapsule{
		Start:  body.Position.subtract(offset),
		End:    body.Position.add(offset),
		Radius: body.Shape.Radius,
	}
}

func polygonPoints(body CollisionBody) []Vector2 {
	var localPoints []Vector2
	if body.Shape.Type == CollisionShapeRectangle {
		half := body.Shape.Size.multiply(0.5)
		localPoints = []Vector2{
			{X: -half.X, Y: -half.Y},
			{X: half.X, Y: -half.Y},
			{X: half.X, Y: half.Y},
			{X: -half.X, Y: half.Y},
		}
	} else {
		localPoints = body.Shape.Points
	}

	points := make([]Vector2, 0, len(localPoints))
	for _, point := range localPoints {
		points = append(points, body.Position.add(point.rotated(body.Rotation)))
	}

	return points
}

func circlesIntersect(a collisionCircle, b collisionCircle) bool {
	radius := a.Radius + b.Radius
	return a.Center.subtract(b.Center).lengthSquared() <= radius*radius
}

func capsulesIntersect(a collisionCapsule, b collisionCapsule) bool {
	radius := a.Radius + b.Radius
	return segmentSegmentDistanceSquared(a.Start, a.End, b.Start, b.End) <= radius*radius
}

func circleCapsuleIntersects(circle collisionCircle, capsule collisionCapsule) bool {
	radius := circle.Radius + capsule.Radius
	return pointSegmentDistanceSquared(circle.Center, capsule.Start, capsule.End) <= radius*radius
}

func circlePolygonIntersects(circle collisionCircle, polygon []Vector2) bool {
	if len(polygon) < 3 {
		return false
	}
	if pointInPolygon(circle.Center, polygon) {
		return true
	}

	radiusSquared := circle.Radius * circle.Radius
	for index, point := range polygon {
		next := polygon[(index+1)%len(polygon)]
		if pointSegmentDistanceSquared(circle.Center, point, next) <= radiusSquared {
			return true
		}
	}

	return false
}

func capsulePolygonIntersects(capsule collisionCapsule, polygon []Vector2) bool {
	if len(polygon) < 3 {
		return false
	}
	if pointInPolygon(capsule.Start, polygon) || pointInPolygon(capsule.End, polygon) {
		return true
	}

	radiusSquared := capsule.Radius * capsule.Radius
	for index, point := range polygon {
		next := polygon[(index+1)%len(polygon)]
		if segmentSegmentDistanceSquared(capsule.Start, capsule.End, point, next) <= radiusSquared {
			return true
		}
	}
	for _, point := range polygon {
		if pointSegmentDistanceSquared(point, capsule.Start, capsule.End) <= radiusSquared {
			return true
		}
	}

	return false
}

func polygonsIntersect(a []Vector2, b []Vector2) bool {
	if len(a) < 3 || len(b) < 3 {
		return false
	}

	for _, point := range a {
		if pointInPolygon(point, b) {
			return true
		}
	}
	for _, point := range b {
		if pointInPolygon(point, a) {
			return true
		}
	}
	for i, point := range a {
		next := a[(i+1)%len(a)]
		for j, other := range b {
			otherNext := b[(j+1)%len(b)]
			if segmentsIntersect(point, next, other, otherNext) {
				return true
			}
		}
	}

	return false
}

func pointSegmentDistanceSquared(point Vector2, start Vector2, end Vector2) float64 {
	closest := closestPointOnSegment(point, start, end)
	return point.subtract(closest).lengthSquared()
}

func segmentSegmentDistanceSquared(aStart Vector2, aEnd Vector2, bStart Vector2, bEnd Vector2) float64 {
	if segmentsIntersect(aStart, aEnd, bStart, bEnd) {
		return 0
	}

	return minFloat(
		pointSegmentDistanceSquared(aStart, bStart, bEnd),
		pointSegmentDistanceSquared(aEnd, bStart, bEnd),
		pointSegmentDistanceSquared(bStart, aStart, aEnd),
		pointSegmentDistanceSquared(bEnd, aStart, aEnd),
	)
}

func closestPointOnSegment(point Vector2, start Vector2, end Vector2) Vector2 {
	segment := end.subtract(start)
	lengthSquared := segment.lengthSquared()
	if lengthSquared == 0 {
		return start
	}

	t := point.subtract(start).dot(segment) / lengthSquared
	t = math.Max(0, math.Min(1, t))
	return start.add(segment.multiply(t))
}

func pointInPolygon(point Vector2, polygon []Vector2) bool {
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		current := polygon[i]
		previous := polygon[j]
		if (current.Y > point.Y) != (previous.Y > point.Y) {
			x := (previous.X-current.X)*(point.Y-current.Y)/(previous.Y-current.Y) + current.X
			if point.X < x {
				inside = !inside
			}
		}
		j = i
	}

	return inside
}

func segmentsIntersect(a Vector2, b Vector2, c Vector2, d Vector2) bool {
	aSide := orientation(a, b, c)
	bSide := orientation(a, b, d)
	cSide := orientation(c, d, a)
	dSide := orientation(c, d, b)

	if aSide == 0 && pointOnSegment(c, a, b) {
		return true
	}
	if bSide == 0 && pointOnSegment(d, a, b) {
		return true
	}
	if cSide == 0 && pointOnSegment(a, c, d) {
		return true
	}
	if dSide == 0 && pointOnSegment(b, c, d) {
		return true
	}

	return (aSide > 0) != (bSide > 0) && (cSide > 0) != (dSide > 0)
}

func orientation(a Vector2, b Vector2, c Vector2) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func pointOnSegment(point Vector2, start Vector2, end Vector2) bool {
	return point.X >= math.Min(start.X, end.X) &&
		point.X <= math.Max(start.X, end.X) &&
		point.Y >= math.Min(start.Y, end.Y) &&
		point.Y <= math.Max(start.Y, end.Y)
}

func minFloat(values ...float64) float64 {
	result := values[0]
	for _, value := range values[1:] {
		result = math.Min(result, value)
	}

	return result
}
