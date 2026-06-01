package player

type TargetStatus string

const (
	TargetStatusMissing  TargetStatus = "missing"
	TargetStatusInactive TargetStatus = "inactive"
	TargetStatusActive   TargetStatus = "active"
)

func TargetStatusForWorldState(state WorldState, exists bool) TargetStatus {
	if !exists {
		return TargetStatusMissing
	}
	if state.Targetable {
		return TargetStatusActive
	}
	return TargetStatusInactive
}
