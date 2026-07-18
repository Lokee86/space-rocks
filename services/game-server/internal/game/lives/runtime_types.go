package lives

import (
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

const InfiniteLives = -1

const (
	StatusRemoved playerstate.Status = "removed"
)

type ParticipantRegistration struct {
	PlayerID string
	TeamID   teams.ID
}

type ParticipantState struct {
	PlayerID         string
	TeamID           teams.ID
	Status           playerstate.Status
	InfiniteOverride bool
	StartingLives    int
	RemainingLives   int
	EffectiveLives   int
	DeathCount       int
	RespawnCount     int
	RespawnCooldown  float64
}

type TeamPoolState struct {
	TeamID         teams.ID
	StartingLives  int
	RemainingLives int
}

type LifeMutation struct {
	Accepted        bool
	PlayerID        string
	PreviousLives   int
	CurrentLives    int
	Delta           int
	ReasonCode      string
	PreviousStatus  playerstate.Status
	ResultingStatus playerstate.Status
}

type TransitionFact struct {
	Accepted        bool
	PlayerID        string
	PreviousStatus  playerstate.Status
	ResultingStatus playerstate.Status
	ReasonCode      string
}

type RemovalFact struct {
	TransitionFact
}
