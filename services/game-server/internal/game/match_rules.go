package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) ConfigureMatchRules(resolved modes.ResolvedMatchRules) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if len(game.participantRecords) > 0 || game.lockedFinalMatchState != nil {
		return fmt.Errorf("match rules must be configured before participation begins")
	}
	if err := teams.ValidateConfig(resolved.TeamConfig); err != nil {
		return fmt.Errorf("invalid resolved team configuration: %w", err)
	}
	if resolved.ModeID != modes.ModeArcadeSurvival && resolved.ModeID != modes.ModeScoreAttack {
		return fmt.Errorf("unsupported resolved mode %q", resolved.ModeID)
	}
	if resolved.AwardPolicyID != awards.StandardPolicyID {
		return fmt.Errorf("unsupported award policy %q", resolved.AwardPolicyID)
	}
	if resolved.PlayerSpawnProfileID != lives.DefaultSpawnProfileID || resolved.LivesPolicy.SpawnProfileID != resolved.PlayerSpawnProfileID {
		return fmt.Errorf("unsupported player spawn profile %q", resolved.PlayerSpawnProfileID)
	}
	if len(resolved.EncounterSpawnProfileIDs) != 1 || resolved.EncounterSpawnProfileIDs[0] != string(encounterspawn.ProfilePlayercentricAsteroidsV1) {
		return fmt.Errorf("unsupported encounter spawn profile selection")
	}
	lifeRuntime, err := lives.NewRuntime(resolved.LivesPolicy)
	if err != nil {
		return err
	}
	objectiveRuntime := objectives.NewRuntime()
	if resolved.ModeID == modes.ModeScoreAttack {
		if err := objectiveRuntime.RegisterDefinition(scoreAttackObjectiveDefinition(resolved)); err != nil {
			return err
		}
	}

	game.resolvedMatchRules = modes.CloneResolvedMatchRules(resolved)
	game.modeID = string(resolved.ModeID)
	game.teamStructure = resolved.TeamConfig.Structure
	game.lifeRuntime = lifeRuntime
	game.objectiveRuntime = objectiveRuntime
	game.awardPolicy = awards.NewStandardPolicy()
	game.matchElapsed = 0
	game.scoreCompletionTimes = make(map[string]float64)
	game.scoreSuccessOrders = make(map[string]int)
	game.nextScoreSuccessOrder = 0
	return nil
}

func (game *Game) ResolvedMatchRules() modes.ResolvedMatchRules {
	game.mu.Lock()
	defer game.mu.Unlock()
	return modes.CloneResolvedMatchRules(game.resolvedMatchRules)
}

func scoreAttackObjectiveDefinition(resolved modes.ResolvedMatchRules) objectives.Definition {
	return objectives.Definition{
		ID:    objectives.DefinitionID(resolved.ObjectivePolicy.DefinitionID),
		Scope: objectives.ScopePlayer,
		Success: objectives.Condition{
			Kind: objectives.ConditionNumeric, FactKey: "counter:" + string(awards.CounterScore),
			Target: float64(resolved.ObjectivePolicy.TargetScore), AllowDecrease: true,
			Overflow: objectives.OverflowClamp, AllowedAttribution: []objectives.AttributionKind{objectives.AttributionInGame},
		},
		Lifecycle: objectives.LifecycleDefinition{InitiallyActive: true, Visibility: objectives.VisibilityPublic},
	}
}

func (game *Game) registerModeObjectivesForPlayerLocked(playerID string) error {
	if game.resolvedMatchRules.ModeID != modes.ModeScoreAttack {
		return nil
	}
	_, events, err := game.objectivesRuntime().CreateInstance(
		objectives.DefinitionID(game.resolvedMatchRules.ObjectivePolicy.DefinitionID),
		objectives.Registration{InstanceID: objectives.InstanceID("score-attack:" + playerID), OwnerID: playerID},
	)
	game.publishObjectiveEventsLocked(events)
	return err
}

func (game *Game) recordScoreAttackSuccessLocked(playerID string, score int) {
	if game.resolvedMatchRules.ModeID != modes.ModeScoreAttack || score < game.resolvedMatchRules.ObjectivePolicy.TargetScore {
		return
	}
	if _, exists := game.scoreSuccessOrders[playerID]; exists {
		return
	}
	game.nextScoreSuccessOrder++
	game.scoreSuccessOrders[playerID] = game.nextScoreSuccessOrder
	game.scoreCompletionTimes[playerID] = game.matchElapsed
}
