package game

import "math"

type Vector2 struct {
	X float64
	Y float64
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
