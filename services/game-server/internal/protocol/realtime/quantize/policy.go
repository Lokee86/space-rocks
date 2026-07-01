package quantize

type PolicyName string

const (
	PolicyFloatGeneric    PolicyName = "float_generic"
	PolicyRatio01         PolicyName = "ratio_0_1"
	PolicyPercent0100     PolicyName = "percent_0_100"
	PolicySeconds         PolicyName = "seconds"
	PolicySignedSeconds   PolicyName = "signed_seconds"
	PolicyAngleTurn       PolicyName = "angle_turn"
	PolicyPosition        PolicyName = "position"
	PolicyVelocity        PolicyName = "velocity"
	PolicyAngularVelocity PolicyName = "angular_velocity"
)

type Kind string

const (
	KindSigned   Kind = "signed"
	KindUnsigned Kind = "unsigned"
)

type Mode string

const (
	ModeScaled Mode = ""
)

const (
	ModeRegularScale Mode = "regular_scale"
	ModeRatio        Mode = "ratio"
	ModeAngleTurn    Mode = "angle_turn"
)

type Policy struct {
	Name     PolicyName
	Kind     Kind
	Mode     Mode
	Scale    int64
	Min      int64
	Max      int64
}

var policies = map[PolicyName]Policy{
	PolicyFloatGeneric: {
		Name:  PolicyFloatGeneric,
		Kind:  KindSigned,
		Mode:  ModeRegularScale,
		Scale: 1000,
		Min:   -2147483648,
		Max:   2147483647,
	},
	PolicyRatio01: {
		Name:  PolicyRatio01,
		Kind:  KindUnsigned,
		Mode:  ModeRatio,
		Min:   0,
		Max:   65535,
	},
	PolicyPercent0100: {
		Name:  PolicyPercent0100,
		Kind:  KindUnsigned,
		Mode:  ModeRegularScale,
		Scale: 100,
		Min:   0,
		Max:   10000,
	},
	PolicySeconds: {
		Name:  PolicySeconds,
		Kind:  KindUnsigned,
		Mode:  ModeRegularScale,
		Scale: 1000,
		Min:   0,
		Max:   4294967295,
	},
	PolicySignedSeconds: {
		Name:  PolicySignedSeconds,
		Kind:  KindSigned,
		Mode:  ModeRegularScale,
		Scale: 1000,
		Min:   -2147483648,
		Max:   2147483647,
	},
	PolicyAngleTurn: {
		Name:  PolicyAngleTurn,
		Kind:  KindUnsigned,
		Mode:  ModeAngleTurn,
		Min:   0,
		Max:   65535,
	},
	PolicyPosition: {
		Name:  PolicyPosition,
		Kind:  KindSigned,
		Mode:  ModeRegularScale,
		Scale: 10,
		Min:   -2147483648,
		Max:   2147483647,
	},
	PolicyVelocity: {
		Name:  PolicyVelocity,
		Kind:  KindSigned,
		Mode:  ModeRegularScale,
		Scale: 10,
		Min:   -2147483648,
		Max:   2147483647,
	},
	PolicyAngularVelocity: {
		Name:  PolicyAngularVelocity,
		Kind:  KindSigned,
		Mode:  ModeRegularScale,
		Scale: 1000,
		Min:   -2147483648,
		Max:   2147483647,
	},
}

func PolicyByName(name PolicyName) (Policy, bool) {
	policy, ok := policies[name]
	return policy, ok
}
