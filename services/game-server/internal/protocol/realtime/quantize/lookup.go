package quantize

import "github.com/Lokee86/space-rocks/server/internal/protocol/realtimewire"

func LookupPolicy(fieldPath string) (Policy, bool) {
	policyName, ok := realtimewire.RealtimeWireQuantizations[fieldPath]
	if !ok {
		return mustPolicy(PolicyFloatGeneric), false
	}
	policy, ok := PolicyByName(PolicyName(policyName))
	if !ok {
		return mustPolicy(PolicyFloatGeneric), false
	}
	return policy, true
}

func mustPolicy(name PolicyName) Policy {
	policy, ok := PolicyByName(name)
	if !ok {
		panic("quantize: missing policy " + string(name))
	}
	return policy
}
