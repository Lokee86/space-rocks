package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) ConfigureMatchRules(resolved modes.ResolvedMatchRules) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if len(game.participantRecords) > 0 {
		return fmt.Errorf("match rules must be configured before participation begins")
	}
	if err := teams.ValidateConfig(resolved.TeamConfig); err != nil {
		return fmt.Errorf("invalid resolved team configuration: %w", err)
	}
	if resolved.ModeID != modes.ModeArcadeSurvival && resolved.ModeID != modes.ModeScoreAttack && resolved.ModeID != modes.ModeDeathmatch {
		return fmt.Errorf("unsupported resolved mode %q", resolved.ModeID)
	}
	if resolved.LivesPolicy.InfiniteLives {
		if resolved.LivesPolicy.StartingLives != 0 {
			return fmt.Errorf("infinite lives cannot define starting lives")
		}
	} else if resolved.LivesPolicy.StartingLives <= 0 {
		return fmt.Errorf("starting lives must be positive")
	}
	if resolved.ModeID == modes.ModeScoreAttack && resolved.ObjectivePolicy.TargetScore <= 0 {
		return fmt.Errorf("score attack target must be positive")
	}
	if resolved.ModeID == modes.ModeDeathmatch && resolved.ObjectivePolicy.TargetKills <= 0 {
		return fmt.Errorf("deathmatch kill target must be positive")
	}

	game.resolvedMatchRules = modes.CloneResolvedMatchRules(resolved)
	game.modeID = string(resolved.ModeID)
	game.teamStructure = resolved.TeamConfig.Structure
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

func (game *Game) applyResolvedMatchRulesToSessionLocked(session *playerSession) {
	if session == nil {
		return
	}
	policy := game.resolvedMatchRules.LivesPolicy
	session.LifeOptions.SetInfiniteLives(policy.InfiniteLives)
	if policy.InfiniteLives {
		session.Lives = constants.PlayerStartingLives
		return
	}
	session.Lives = policy.StartingLives
}

func (game *Game) recordModeScoreSuccessLocked(playerID string, score int) {
	target := 0
	switch game.resolvedMatchRules.ModeID {
	case modes.ModeScoreAttack:
		target = game.resolvedMatchRules.ObjectivePolicy.TargetScore
	case modes.ModeDeathmatch:
		target = game.resolvedMatchRules.ObjectivePolicy.TargetKills
	default:
		return
	}
	if score < target {
		return
	}
	if _, exists := game.scoreSuccessOrders[playerID]; exists {
		return
	}
	game.nextScoreSuccessOrder++
	game.scoreSuccessOrders[playerID] = game.nextScoreSuccessOrder
	game.scoreCompletionTimes[playerID] = game.matchElapsed
}
