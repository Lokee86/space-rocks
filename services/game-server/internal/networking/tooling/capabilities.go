package tooling

// CapabilitySet contains the capabilities granted to a tooling connection.
type CapabilitySet map[string]struct{}

// Has reports whether the capability is granted.
func (set CapabilitySet) Has(capability string) bool {
	_, ok := set[capability]
	return ok
}

// NewTemporaryCapabilitySet returns the current temporary all-connection policy.
func NewTemporaryCapabilitySet() CapabilitySet {
	return CapabilitySet{
		CapabilityToolingRead:    {},
		CapabilityToolingControl: {},
	}
}
