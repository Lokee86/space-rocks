package game

import "math"

type Vector2 struct {
	X float64
	Y float64
}

func (vector Vector2) directionTo(target Vector2) Vector2 {
	return Vector2{
		X: target.X - vector.X,
		Y: target.Y - vector.Y,
	}.normalized()
}

func (vector Vector2) normalized() Vector2 {
	length := math.Hypot(vector.X, vector.Y)
	if length == 0 {
		return Vector2{}
	}

	return Vector2{
		X: vector.X / length,
		Y: vector.Y / length,
	}
}

func (vector Vector2) rotated(angle float64) Vector2 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	return Vector2{
		X: vector.X*cos - vector.Y*sin,
		Y: vector.X*sin + vector.Y*cos,
	}
}

func (vector Vector2) limitLength(maxLength float64) Vector2 {
	length := math.Hypot(vector.X, vector.Y)
	if length <= maxLength {
		return vector
	}

	scale := maxLength / length
	return Vector2{
		X: vector.X * scale,
		Y: vector.Y * scale,
	}
}
