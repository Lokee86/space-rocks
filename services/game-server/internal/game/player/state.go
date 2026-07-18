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
	Status          Status
	HasActiveShip   bool
	X               float64
	Y               float64
	Lives           int
	RespawnCooldown float64
}

func BuildWorldState(input BuildWorldStateInput) WorldState {
	isActive := input.Status == StatusActive

	return WorldState{
		ID:              input.ID,
		Status:          input.Status,
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
