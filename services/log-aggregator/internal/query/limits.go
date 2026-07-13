package query

const (
	defaultQueryLimit = 100
	maximumQueryLimit = 1000
)

type LimitPolicy struct {
	Default int
	Maximum int
}

func (policy LimitPolicy) normalized() LimitPolicy {
	if policy.Default <= 0 {
		policy.Default = defaultQueryLimit
	}
	if policy.Maximum <= 0 {
		policy.Maximum = maximumQueryLimit
	}
	if policy.Default > policy.Maximum {
		policy.Default = policy.Maximum
	}
	return policy
}

