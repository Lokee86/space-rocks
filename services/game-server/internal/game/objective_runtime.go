package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
)

func (game *Game) objectivesRuntime() *objectives.Runtime {
	if game.objectiveRuntime == nil {
		game.objectiveRuntime = objectives.NewRuntime()
	}
	return game.objectiveRuntime
}

func (game *Game) RegisterObjectiveDefinition(definition objectives.Definition) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	return game.objectivesRuntime().RegisterDefinition(definition)
}

func (game *Game) CreateObjectiveInstance(
	definitionID objectives.DefinitionID,
	registration objectives.Registration,
) (objectives.InstanceID, error) {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return "", fmt.Errorf("match results are locked")
	}
	instanceID, events, err := game.objectivesRuntime().CreateInstance(definitionID, registration)
	game.publishObjectiveEventsLocked(events)
	return instanceID, err
}

func (game *Game) RetireObjectiveDefinition(definitionID objectives.DefinitionID) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := game.objectivesRuntime().RetireDefinition(definitionID)
	game.publishObjectiveEventsLocked(events)
	return err
}

func (game *Game) ApplyObjectiveFacts(instanceID objectives.InstanceID, facts []objectives.Fact) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := game.objectivesRuntime().ApplyFacts(instanceID, facts)
	game.publishObjectiveEventsLocked(events)
	return err
}

func (game *Game) ApplyObjectiveFactsToScope(scope objectives.Scope, ownerID string, facts []objectives.Fact) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := game.objectivesRuntime().ApplyFactsToScope(scope, ownerID, facts)
	game.publishObjectiveEventsLocked(events)
	return err
}

func (game *Game) ObjectiveSnapshot(instanceID objectives.InstanceID, viewer objectives.Viewer) (objectives.Snapshot, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.objectivesRuntime().Snapshot(instanceID, viewer)
}

func (game *Game) ObjectiveSnapshots(viewer objectives.Viewer) []objectives.Snapshot {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.objectivesRuntime().Snapshots(viewer)
}

func (game *Game) AddObjectiveEventObserver(observer func(objectives.Event)) {
	if observer == nil {
		return
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	game.objectiveEventObservers = append(game.objectiveEventObservers, observer)
}

func (game *Game) publishObjectiveEventsLocked(events []objectives.Event) {
	for _, event := range events {
		for _, observer := range game.objectiveEventObservers {
			observer(event)
		}
	}
}

func (game *Game) stepObjectives(delta float64) {
	if game.lockedFinalMatchState != nil {
		return
	}
	simulationPaused := len(game.playerSessions) == 0
	events := game.objectivesRuntime().Step(delta, simulationPaused)
	game.publishObjectiveEventsLocked(events)
}

func (game *Game) applyAwardResultToObjectivesLocked(result awards.EventResult) {
	if !result.Applied {
		return
	}
	for _, change := range result.Changes {
		scope, ok := objectiveScopeForAwardScope(change.Owner.Scope)
		if !ok {
			continue
		}
		attribution := objectives.Attribution{Kind: objectives.AttributionInGame}
		switch change.Owner.Scope {
		case awards.ScopePlayer:
			attribution.PlayerID = change.Owner.ID
		case awards.ScopeTeam:
			attribution.TeamID = change.Owner.ID
		}
		fact := objectives.Fact{
			Key:         fmt.Sprintf("counter:%s", change.CounterID),
			Operation:   objectives.FactSet,
			Number:      change.After,
			Attribution: attribution,
		}
		events, err := game.objectivesRuntime().ApplyFactsToScope(scope, change.Owner.ID, []objectives.Fact{fact})
		if err == nil {
			game.publishObjectiveEventsLocked(events)
		}
	}
}

func objectiveScopeForAwardScope(scope awards.Scope) (objectives.Scope, bool) {
	switch scope {
	case awards.ScopePlayer:
		return objectives.ScopePlayer, true
	case awards.ScopeTeam:
		return objectives.ScopeTeam, true
	case awards.ScopeMatch:
		return objectives.ScopeMatch, true
	case awards.ScopeObjective:
		return objectives.ScopeCustom, true
	default:
		return "", false
	}
}
