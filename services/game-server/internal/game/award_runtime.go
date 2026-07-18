package game

import (
	"fmt"
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
)

type GameplayAwardSnapshot struct {
	Counters   []awards.CounterSnapshot
	TeamTotals []awards.CounterSnapshot
	Combos     []awards.ComboSnapshot
	Streaks    []awards.StreakSnapshot
}

func (game *Game) awardsRuntime() *awards.Runtime {
	if game.awardRuntime == nil {
		game.awardRuntime = awards.NewRuntime()
	}
	if game.awardPolicy.ID == "" {
		game.awardPolicy = awards.NewStandardPolicy()
	}
	return game.awardRuntime
}

func (game *Game) SetAwardPolicy(policy awards.Policy) {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.awardPolicy = policy.Normalize()
}

func (game *Game) AwardPolicy() awards.Policy {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.awardPolicy
}

func (game *Game) RegisterCustomCounter(id awards.CounterID) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.awardsRuntime().RegisterCustomCounter(id)
}

func (game *Game) RegisterCustomCounterWithVisibility(id awards.CounterID, visibility awards.Visibility) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.awardsRuntime().RegisterCustomCounterWithVisibility(id, visibility)
}

func (game *Game) SetGameplayCounterVisibility(id awards.CounterID, visibility awards.Visibility) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.awardsRuntime().SetCounterVisibility(id, visibility)
}

func (game *Game) ResetGameplayAwards() {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.awardsRuntime().Reset()
	game.awardClock = 0
	for playerID := range game.participantRecords {
		game.syncProjectedPlayerScoreLocked(playerID, 0)
	}
}

func (game *Game) GameplayAwardSnapshot() GameplayAwardSnapshot {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.gameplayAwardSnapshotLocked()
}

func (game *Game) AddGameplayAwardObserver(observer func(awards.EventResult)) {
	if observer == nil {
		return
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	game.awardEventObservers = append(game.awardEventObservers, observer)
}

func (game *Game) ApplyGameplayAwardEvent(eventID string, mutations []awards.Mutation) (awards.EventResult, error) {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return awards.EventResult{}, fmt.Errorf("match results are locked")
	}
	for _, mutation := range mutations {
		if mutation.Owner.Scope == awards.ScopePlayer && !game.playerCanReceiveAwardsLocked(mutation.Owner.ID) {
			return awards.EventResult{}, fmt.Errorf("player %q cannot receive new awards", mutation.Owner.ID)
		}
	}
	return game.applyAwardMutationsLocked(eventID, mutations)
}

func (game *Game) GameplayCounter(owner awards.Owner, id awards.CounterID) (float64, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.awardsRuntime().Counter(owner, id)
}

func (game *Game) gameplayAwardSnapshotLocked() GameplayAwardSnapshot {
	runtime := game.awardsRuntime()
	memberships := make(map[string]string, len(game.participantRecords))
	for playerID, record := range game.participantRecords {
		if record != nil {
			memberships[playerID] = string(record.TeamID)
		}
	}
	return GameplayAwardSnapshot{
		Counters: runtime.Snapshot(),
		TeamTotals: runtime.DerivedTeamTotals(memberships, []awards.CounterID{
			awards.CounterScore, awards.CounterKills, awards.CounterAssists, awards.CounterDeaths,
			awards.CounterDamageDealt, awards.CounterDamageTaken, awards.CounterObjectiveProgress,
			awards.CounterResourcesCollected, awards.CounterCompletionTime,
		}),
		Combos:  runtime.ComboSnapshots(game.awardClock),
		Streaks: runtime.StreakSnapshot(),
	}
}

func (game *Game) nextAwardEventIDLocked(prefix string) string {
	game.nextAwardEventID++
	return fmt.Sprintf("%s:%d", prefix, game.nextAwardEventID)
}

func (game *Game) applyAwardMutationsLocked(eventID string, mutations []awards.Mutation) (awards.EventResult, error) {
	if game.lockedFinalMatchState != nil {
		return awards.EventResult{}, fmt.Errorf("match results are locked")
	}
	result, err := game.awardsRuntime().ApplyEvent(eventID, mutations)
	if err != nil || !result.Applied {
		return result, err
	}
	for _, change := range result.Changes {
		if change.Owner.Scope == awards.ScopePlayer && change.CounterID == awards.CounterScore {
			game.syncProjectedPlayerScoreLocked(change.Owner.ID, change.After)
		}
	}
	game.applyAwardResultToObjectivesLocked(result)
	if decision := game.evaluateMatchDecisionLocked(); decision.IsOver {
		_, _ = game.lockFinalMatchStateForDecisionLocked(decision)
	}
	for _, observer := range game.awardEventObservers {
		observer(result)
	}
	return result, nil
}

func (game *Game) syncProjectedPlayerScoreLocked(playerID string, value float64) {
	score := int(math.Round(value))
	if score < 0 {
		score = 0
	}
	if session, ok := game.playerSessions[playerID]; ok && session != nil {
		session.Score = score
	}
	if record, ok := game.participantRecords[playerID]; ok && record != nil {
		record.Score = score
		game.recordScoreAttackSuccessLocked(playerID, score)
	}
}

func (game *Game) playerCanReceiveAwardsLocked(playerID string) bool {
	_, ok := game.playerSessions[playerID]
	return ok
}

func (game *Game) eligibleAwardPlayersLocked() map[string]bool {
	eligible := make(map[string]bool, len(game.playerSessions))
	for playerID := range game.playerSessions {
		eligible[playerID] = true
	}
	return eligible
}
