package targeting

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type TargetKind string

const (
	TargetKindPlayer   TargetKind = "player"
	TargetKindEnemy    TargetKind = "enemy"
	TargetKindAsteroid TargetKind = "asteroid"
	TargetKindBullet   TargetKind = "bullet"
)

type TargetRef struct {
	Kind TargetKind
	ID   string
}

type TargetCandidate struct {
	Ref  TargetRef
	Body physics.CollisionBody
}

func EmptyTarget() TargetRef {
	return TargetRef{}
}

func (target TargetRef) IsEmpty() bool {
	return target.Kind == "" || target.ID == ""
}

func TargetKindPriority(kind TargetKind) int {
	switch kind {
	case TargetKindPlayer:
		return 4
	case TargetKindEnemy:
		return 3
	case TargetKindAsteroid:
		return 2
	case TargetKindBullet:
		return 1
	default:
		return 0
	}
}

// ValidateRequestedTarget returns the accepted target_player_id and whether
// the request is valid.
func ValidateRequestedTarget(
	requesterPlayerID string,
	requestedTargetPlayerID string,
	playerExists func(playerID string) bool,
) (string, bool) {
	if playerExists == nil {
		return "", false
	}

	if !playerExists(requesterPlayerID) {
		return "", false
	}

	if requestedTargetPlayerID == "" {
		return "", true
	}

	if !playerExists(requestedTargetPlayerID) {
		return "", false
	}

	return requestedTargetPlayerID, true
}
