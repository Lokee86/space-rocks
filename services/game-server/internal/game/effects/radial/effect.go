package radial

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type SpawnRequest struct {
	ID            string
	SourceID      string
	SourcePlayerID string
	Origin        physics.Vector2
	Spec          Spec
}

type Effect struct {
	ID             string
	SourceID       string
	SourcePlayerID string
	Origin         physics.Vector2
	Spec           Spec
	AgeSeconds     float64
	Zones          []Zone
}

func NewEffect(req SpawnRequest) Effect {
	return Effect{
		ID:             req.ID,
		SourceID:       req.SourceID,
		SourcePlayerID: req.SourcePlayerID,
		Origin:         req.Origin,
		Spec:           req.Spec,
		Zones:          buildZones(req.Spec),
	}
}
