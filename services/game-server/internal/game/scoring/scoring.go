package scoring

import "github.com/Lokee86/space-rocks/server/internal/constants"

type EventKind string

const (
	EventAsteroidDestroyed EventKind = "asteroid_destroyed"
)

type Event struct {
	Kind         EventKind
	PlayerID     string
	TargetID     string
	AsteroidSize int
}

type Award struct {
	PlayerID string
	Points   int
	Reason   EventKind
}

type Policy struct {
	BaseAsteroidDestroyedPoints int
}

func NewDefaultPolicy() Policy {
	return Policy{
		BaseAsteroidDestroyedPoints: constants.BaseScore,
	}
}

func (policy Policy) Evaluate(event Event) []Award {
	switch event.Kind {
	case EventAsteroidDestroyed:
		return policy.evaluateAsteroidDestroyed(event)
	default:
		return nil
	}
}

func (policy Policy) evaluateAsteroidDestroyed(event Event) []Award {
	if event.PlayerID == "" || event.AsteroidSize <= 0 {
		return nil
	}

	return []Award{{
		PlayerID: event.PlayerID,
		Points:   policy.BaseAsteroidDestroyedPoints / event.AsteroidSize,
		Reason:   event.Kind,
	}}
}
