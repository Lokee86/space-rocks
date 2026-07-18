package lives

import (
	"fmt"
	"sort"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type Runtime struct {
	policy       Policy
	participants map[string]*participantRecord
	teamPools    map[teams.ID]*TeamPoolState
}

type participantRecord struct {
	playerID         string
	teamID           teams.ID
	status           playerstate.Status
	infiniteOverride bool
	remainingLives   int
	deathCount       int
	respawnCount     int
	respawnCooldown  float64
	deathExhausted   bool
	deathHistory     []DeathFact
}

func (runtime *Runtime) DeathHistory(playerID string) ([]DeathFact, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return nil, false
	}
	history := make([]DeathFact, len(participant.deathHistory))
	for index, fact := range participant.deathHistory {
		history[index] = fact.clone()
	}
	return history, true
}

func NewRuntime(policy Policy) (*Runtime, error) {
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lives policy: %w", err)
	}

	return &Runtime{
		policy:       clonePolicy(policy),
		participants: make(map[string]*participantRecord),
		teamPools:    make(map[teams.ID]*TeamPoolState),
	}, nil
}

func (runtime *Runtime) Policy() Policy {
	return clonePolicy(runtime.policy)
}

func (runtime *Runtime) RegisterParticipant(registration ParticipantRegistration) error {
	if registration.PlayerID == "" {
		return fmt.Errorf("participant player ID is required")
	}
	if _, exists := runtime.participants[registration.PlayerID]; exists {
		return fmt.Errorf("participant %q is already registered", registration.PlayerID)
	}
	if runtime.policy.Model == LifeModelSharedTeamPool {
		if registration.TeamID == teams.NoTeam {
			return fmt.Errorf("shared team-pool participants require a team ID")
		}
		if _, exists := runtime.teamPools[registration.TeamID]; !exists {
			runtime.teamPools[registration.TeamID] = &TeamPoolState{
				TeamID:         registration.TeamID,
				StartingLives:  runtime.policy.TeamPool.StartingLives,
				RemainingLives: runtime.policy.TeamPool.StartingLives,
			}
		}
	}

	remainingLives := runtime.policy.StartingLives
	if runtime.policy.Model == LifeModelSharedTeamPool {
		remainingLives = 0
	}
	if runtime.policy.Model == LifeModelInfinite {
		remainingLives = InfiniteLives
	}
	runtime.participants[registration.PlayerID] = &participantRecord{
		playerID:       registration.PlayerID,
		teamID:         registration.TeamID,
		status:         playerstate.StatusActive,
		remainingLives: remainingLives,
	}
	return nil
}

func (runtime *Runtime) RollbackParticipant(playerID string) bool {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return false
	}
	delete(runtime.participants, playerID)
	if runtime.policy.Model == LifeModelSharedTeamPool {
		for _, other := range runtime.participants {
			if other.teamID == participant.teamID {
				return true
			}
		}
		delete(runtime.teamPools, participant.teamID)
	}
	return true
}

func (runtime *Runtime) ParticipantSnapshot(playerID string) (ParticipantState, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return ParticipantState{}, false
	}
	return runtime.snapshotParticipant(participant), true
}

func (runtime *Runtime) ParticipantSnapshots() []ParticipantState {
	playerIDs := make([]string, 0, len(runtime.participants))
	for playerID := range runtime.participants {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)

	snapshots := make([]ParticipantState, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		snapshots = append(snapshots, runtime.snapshotParticipant(runtime.participants[playerID]))
	}
	return snapshots
}

func (runtime *Runtime) TeamPoolSnapshot(teamID teams.ID) (TeamPoolState, bool) {
	pool, ok := runtime.teamPools[teamID]
	if !ok {
		return TeamPoolState{}, false
	}
	return *pool, true
}

func (runtime *Runtime) TeamPoolSnapshots() []TeamPoolState {
	teamIDs := make([]teams.ID, 0, len(runtime.teamPools))
	for teamID := range runtime.teamPools {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(i, j int) bool { return teamIDs[i] < teamIDs[j] })

	snapshots := make([]TeamPoolState, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		snapshots = append(snapshots, *runtime.teamPools[teamID])
	}
	return snapshots
}

func (runtime *Runtime) snapshotParticipant(participant *participantRecord) ParticipantState {
	startingLives := runtime.policy.StartingLives
	if runtime.policy.Model == LifeModelSharedTeamPool {
		pool := runtime.teamPools[participant.teamID]
		if pool != nil {
			startingLives = pool.StartingLives
		}
	}
	if runtime.policy.Model == LifeModelInfinite {
		startingLives = InfiniteLives
	}

	return ParticipantState{
		PlayerID:         participant.playerID,
		TeamID:           participant.teamID,
		Status:           participant.status,
		InfiniteOverride: participant.infiniteOverride,
		StartingLives:    startingLives,
		RemainingLives:   participant.remainingLives,
		EffectiveLives:   runtime.effectiveLives(participant),
		DeathCount:       participant.deathCount,
		RespawnCount:     participant.respawnCount,
		RespawnCooldown:  participant.respawnCooldown,
	}
}

func clonePolicy(policy Policy) Policy {
	if policy.TeamPool != nil {
		teamPool := *policy.TeamPool
		policy.TeamPool = &teamPool
	}
	return policy
}
