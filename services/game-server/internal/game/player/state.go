package player

type Status string

const (
	StatusActive         Status = "active"
	StatusPendingRespawn Status = "pending_respawn"
	StatusEliminated     Status = "eliminated"
)

type WorldState struct {
	ID              string  `json:"id"`
	Status          Status  `json:"status"`
	HasActiveShip   bool    `json:"has_active_ship"`
	Targetable      bool    `json:"targetable"`
	Damageable      bool    `json:"damageable"`
	Collidable      bool    `json:"collidable"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Lives           int     `json:"lives"`
	RespawnCooldown float64 `json:"respawn_cooldown"`
}

type BuildWorldStateInput struct {
	ID              string
	HasActiveShip   bool
	X               float64
	Y               float64
	Lives           int
	RespawnCooldown float64
}

func BuildWorldState(input BuildWorldStateInput) WorldState {
	status := StatusEliminated
	if input.HasActiveShip {
		status = StatusActive
	} else if input.Lives > 0 {
		status = StatusPendingRespawn
	}

	isActive := status == StatusActive

	return WorldState{
		ID:              input.ID,
		Status:          status,
		HasActiveShip:   input.HasActiveShip,
		Targetable:      isActive,
		Damageable:      isActive,
		Collidable:      isActive,
		X:               input.X,
		Y:               input.Y,
		Lives:           input.Lives,
		RespawnCooldown: input.RespawnCooldown,
	}
}
