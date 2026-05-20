package physics

import "math"

type Vector2 struct {
	X float64
	Y float64
}

func (vector Vector2) DirectionTo(target Vector2) Vector2 {
	return Vector2{
		X: target.X - vector.X,
		Y: target.Y - vector.Y,
	}.Normalized()
}

func (vector Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: vector.X + other.X, Y: vector.Y + other.Y}
}

func (vector Vector2) Subtract(other Vector2) Vector2 {
	return Vector2{X: vector.X - other.X, Y: vector.Y - other.Y}
}

func (vector Vector2) Multiply(scalar float64) Vector2 {
	return Vector2{X: vector.X * scalar, Y: vector.Y * scalar}
}

func (vector Vector2) Dot(other Vector2) float64 {
	return vector.X*other.X + vector.Y*other.Y
}

func (vector Vector2) Length() float64 {
	return math.Hypot(vector.X, vector.Y)
}

func (vector Vector2) LengthSquared() float64 {
	return vector.X*vector.X + vector.Y*vector.Y
}

func (vector Vector2) Normalized() Vector2 {
	length := vector.Length()
	if length == 0 {
		return Vector2{}
	}

	return Vector2{
		X: vector.X / length,
		Y: vector.Y / length,
	}
}

func (vector Vector2) Rotated(angle float64) Vector2 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	return Vector2{
		X: vector.X*cos - vector.Y*sin,
		Y: vector.X*sin + vector.Y*cos,
	}
}

func (vector Vector2) LimitLength(maxLength float64) Vector2 {
	length := vector.Length()
	if length <= maxLength {
		return vector
	}

	scale := maxLength / length
	return Vector2{
		X: vector.X * scale,
		Y: vector.Y * scale,
	}
}
