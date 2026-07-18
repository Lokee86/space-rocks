package participation

import (
	"fmt"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
)

const DefaultAFKTimeout = 35.0

type AFKPolicy struct {
	Timeout float64
}

func NewDefaultAFKPolicy() AFKPolicy {
	return AFKPolicy{Timeout: DefaultAFKTimeout}
}

type ExpiryRequest struct {
	PlayerID   string
	ReasonCode string
}

type Runtime struct {
	policy       AFKPolicy
	participants map[string]*afkParticipant
}

type afkParticipant struct {
	remaining float64
	expired   bool
}

func NewRuntime(policy AFKPolicy) (*Runtime, error) {
	if policy.Timeout <= 0 {
		return nil, fmt.Errorf("AFK timeout must be positive")
	}
	return &Runtime{policy: policy, participants: make(map[string]*afkParticipant)}, nil
}

func (runtime *Runtime) RegisterParticipant(playerID string) error {
	if playerID == "" {
		return fmt.Errorf("participant player ID is required")
	}
	if _, exists := runtime.participants[playerID]; exists {
		return fmt.Errorf("participant %q is already registered", playerID)
	}
	runtime.participants[playerID] = &afkParticipant{remaining: runtime.policy.Timeout}
	return nil
}

func (runtime *Runtime) UnregisterParticipant(playerID string) bool {
	if _, exists := runtime.participants[playerID]; !exists {
		return false
	}
	delete(runtime.participants, playerID)
	return true
}

func (runtime *Runtime) RecordAction(playerID string) bool {
	participant, ok := runtime.participants[playerID]
	if !ok || participant.expired {
		return false
	}
	participant.remaining = runtime.policy.Timeout
	return true
}

func (runtime *Runtime) Step(delta float64, status func(string) (playerstate.Status, bool)) []ExpiryRequest {
	if delta <= 0 {
		return nil
	}
	requests := make([]ExpiryRequest, 0)
	for playerID, participant := range runtime.participants {
		currentStatus, ok := status(playerID)
		if !ok || (currentStatus != playerstate.StatusActive && currentStatus != playerstate.StatusPendingRespawn) {
			continue
		}
		participant.remaining = max(0, participant.remaining-delta)
		if participant.remaining == 0 && !participant.expired {
			participant.expired = true
			requests = append(requests, ExpiryRequest{PlayerID: playerID, ReasonCode: "afk_forfeit"})
		}
	}
	return requests
}
