package servicelog

// Status is the small operational state exposed by Runtime.
type Status struct {
	Closed       bool
	Degraded     bool
	FailureCount int
	LastError    string
}
