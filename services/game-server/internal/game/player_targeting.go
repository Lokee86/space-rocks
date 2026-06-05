package game

import (
	targetpolicy "github.com/Lokee86/space-rocks/server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

type PlayerTargeting struct {
	Kind string
	ID   string
}

func EmptyPlayerTargeting() PlayerTargeting {
	return PlayerTargeting{}
}

func PlayerTargetingFromRef(target targetpolicy.TargetRef) PlayerTargeting {
	return PlayerTargeting{
		Kind: string(target.Kind),
		ID:   target.ID,
	}
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
}
