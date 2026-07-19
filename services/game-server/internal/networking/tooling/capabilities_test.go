package tooling

import "testing"

func TestCapabilitySetEmptySetDeniesCapabilities(t *testing.T) {
	var capabilities CapabilitySet

	for _, capability := range []string{
		CapabilityToolingRead,
		CapabilityToolingControl,
		CapabilityAdminControl,
	} {
		if capabilities.Has(capability) {
			t.Fatalf("empty capability set grants %q", capability)
		}
	}
}

func TestNewTemporaryCapabilitySetGrantsToolingCapabilitiesOnly(t *testing.T) {
	capabilities := NewTemporaryCapabilitySet()

	for _, capability := range []string{
		CapabilityToolingRead,
		CapabilityToolingControl,
	} {
		if !capabilities.Has(capability) {
			t.Errorf("temporary capability set does not grant %q", capability)
		}
	}

	for _, capability := range []string{
		CapabilityAdminControl,
		"unknown.capability",
	} {
		if capabilities.Has(capability) {
			t.Errorf("temporary capability set grants %q", capability)
		}
	}
}
