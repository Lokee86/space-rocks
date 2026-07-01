package quantize

import (
	"errors"
	"math"
)

var (
	ErrInvalidFloatValue = errors.New("quantize: invalid float value")
	ErrInvalidPolicy     = errors.New("quantize: invalid policy")
)

func EncodeFloat(policy Policy, value float64) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrInvalidFloatValue
	}

	switch policy.Name {
	case PolicyRatio01:
		return roundToInt64(clampFloat(value, 0, 1)*65535.0), nil
	case PolicyPercent0100:
		return roundToInt64(clampFloat(value, 0, 100)*100.0), nil
	case PolicySeconds:
		return roundToInt64(clampFloat(value, 0, math.MaxInt32/1000.0)*1000.0), nil
	case PolicySignedSeconds:
		return roundToInt64(clampFloat(value, float64(math.MinInt32)/1000.0, float64(math.MaxInt32)/1000.0)*1000.0), nil
	case PolicyAngleTurn:
		return roundToInt64(clampFloat(value, 0, 1)*65535.0), nil
	case PolicyFloatGeneric, PolicyPosition, PolicyVelocity, PolicyAngularVelocity:
		if policy.Scale <= 0 {
			return 0, ErrInvalidPolicy
		}
		return roundToInt64(value * float64(policy.Scale)), nil
	default:
		if policy.Scale <= 0 {
			return 0, ErrInvalidPolicy
		}
		return roundToInt64(value * float64(policy.Scale)), nil
	}
}

func DecodeFloat(policy Policy, encoded int64) (float64, error) {
	switch policy.Name {
	case PolicyRatio01:
		return clampFloat(float64(encoded)/65535.0, 0, 1), nil
	case PolicyPercent0100:
		return clampFloat(float64(encoded)/100.0, 0, 100), nil
	case PolicySeconds:
		return clampFloat(float64(encoded)/1000.0, 0, math.MaxInt32/1000.0), nil
	case PolicySignedSeconds:
		return float64(encoded) / 1000.0, nil
	case PolicyAngleTurn:
		return clampFloat(float64(encoded)/65535.0, 0, 1), nil
	case PolicyFloatGeneric, PolicyPosition, PolicyVelocity, PolicyAngularVelocity:
		if policy.Scale <= 0 {
			return 0, ErrInvalidPolicy
		}
		return float64(encoded) / float64(policy.Scale), nil
	default:
		if policy.Scale <= 0 {
			return 0, ErrInvalidPolicy
		}
		return float64(encoded) / float64(policy.Scale), nil
	}
}

func roundToInt64(value float64) int64 {
	return int64(math.Round(value))
}

func clampFloat(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
