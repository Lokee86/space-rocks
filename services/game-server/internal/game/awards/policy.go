package awards

const StandardPolicyID = "standard_awards_v1"

type Policy struct {
	ID             string
	Assists        AssistPolicy
	AssistScore    float64
	ComboEnabled   bool
	ComboCounter   CounterID
	KillStreakName string
}

func NewStandardPolicy() Policy {
	assist := DefaultAssistPolicy()
	assist.Enabled = false
	return Policy{
		ID:             StandardPolicyID,
		Assists:        assist,
		AssistScore:    0,
		ComboEnabled:   true,
		ComboCounter:   CounterScore,
		KillStreakName: "pvp_kills",
	}
}

func (policy Policy) Normalize() Policy {
	if policy.ID != StandardPolicyID {
		return NewStandardPolicy()
	}
	if policy.Assists.WindowSeconds <= 0 {
		policy.Assists.WindowSeconds = 5
	}
	if policy.Assists.ThresholdFraction <= 0 {
		policy.Assists.ThresholdFraction = 0.05
	}
	if policy.Assists.RetentionBufferFraction < 0 {
		policy.Assists.RetentionBufferFraction = 0.10
	}
	if policy.ComboCounter == "" {
		policy.ComboCounter = CounterScore
	}
	if policy.KillStreakName == "" {
		policy.KillStreakName = "pvp_kills"
	}
	return policy
}
