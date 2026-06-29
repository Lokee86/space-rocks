package realtime

type ResyncDecisionKind string

const (
	ResyncDecisionNone           ResyncDecisionKind = "none"
	ResyncDecisionWrongBaseline  ResyncDecisionKind = "wrong_baseline"
	ResyncDecisionMissingBaseline ResyncDecisionKind = "missing_baseline"
)

type ResyncDecision struct {
	Kind           ResyncDecisionKind
	Lane           Lane
	Sequence       int
	BaselineID     string
	SnapshotID     string
	ServerSentMsec int
}

func DecideResync(state RealtimeSessionState, lane Lane, expectedBaselineID string, requiredBaselineID string, observed RealtimeLaneState, hasObserved bool) ResyncDecision {
	if !hasObserved {
		if requiredBaselineID == "" {
			return ResyncDecision{Kind: ResyncDecisionNone, Lane: lane}
		}
		return ResyncDecision{
			Kind:       ResyncDecisionMissingBaseline,
			Lane:       lane,
			BaselineID: requiredBaselineID,
		}
	}

	if expectedBaselineID != "" && observed.BaselineID != expectedBaselineID {
		return ResyncDecision{
			Kind:       ResyncDecisionWrongBaseline,
			Lane:       lane,
			Sequence:   observed.Sequence,
			BaselineID: observed.BaselineID,
			SnapshotID: observed.SnapshotID,
		}
	}

	if requiredBaselineID != "" && observed.BaselineID == "" {
		return ResyncDecision{
			Kind:       ResyncDecisionMissingBaseline,
			Lane:       lane,
			BaselineID: requiredBaselineID,
		}
	}

	return ResyncDecision{
		Kind:       ResyncDecisionNone,
		Lane:       lane,
		Sequence:   observed.Sequence,
		BaselineID: observed.BaselineID,
		SnapshotID: observed.SnapshotID,
	}
}
