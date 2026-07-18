package objectives

func conditionStatesEqual(left, right conditionState) bool {
	if left.Progress != right.Progress || left.Satisfied != right.Satisfied ||
		left.SequenceIndex != right.SequenceIndex || left.MaintainEnabled != right.MaintainEnabled ||
		left.Attribution != right.Attribution || len(left.Members) != len(right.Members) {
		return false
	}
	for member := range left.Members {
		if _, ok := right.Members[member]; !ok {
			return false
		}
	}
	return true
}

func cloneMembers(source map[string]struct{}) map[string]struct{} {
	if source == nil {
		return nil
	}
	clone := make(map[string]struct{}, len(source))
	for member := range source {
		clone[member] = struct{}{}
	}
	return clone
}
