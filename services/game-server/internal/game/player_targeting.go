package game

import (
	targetpolicy "github.com/Lokee86/space-rocks/server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

type PlayerTargeting struct {
	Kind     string
	ID       string
	PlayerID string
}

func EmptyPlayerTargeting() PlayerTargeting {
	return PlayerTargeting{}
}

func PlayerTargetingFromRef(target targetpolicy.TargetRef) PlayerTargeting {
	targeting := PlayerTargeting{
		Kind: string(target.Kind),
		ID:   target.ID,
	}
	if target.Kind == targetpolicy.TargetKindPlayer {
		targeting.PlayerID = target.ID
	}
	return targeting
}

func (targeting PlayerTargeting) TargetRef() targetpolicy.TargetRef {
	return targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKind(targeting.Kind),
		ID:   targeting.ID,
	}
}

func (targeting PlayerTargeting) ApplyToShip(ship *runtime.Ship) {
	if ship == nil {
		return
	}

	ship.TargetKind = targeting.Kind
	ship.TargetID = targeting.ID
	ship.TargetPlayerID = targeting.PlayerID
}
