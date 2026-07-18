package modes

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func DefaultRoomModeConfig() RoomModeConfig {
	return RoomModeConfig{
		PresetID:      PresetArcadeSurvival,
		StartingLives: lives.NewBaselinePolicy().StartingLives,
	}
}

func NormalizeRoomModeConfig(config RoomModeConfig) RoomModeConfig {
	if config.PresetID == "" {
		config.PresetID = PresetArcadeSurvival
	}
	if !config.InfiniteLives && config.StartingLives == 0 {
		config.StartingLives = lives.NewBaselinePolicy().StartingLives
	}
	if config.PresetID == PresetScoreAttack && config.TargetScore == 0 {
		config.TargetScore = DefaultScoreAttackTarget
	}
	return config
}

func ValidateRoomModeConfig(config RoomModeConfig) error {
	config = NormalizeRoomModeConfig(config)
	switch config.PresetID {
	case PresetArcadeSurvival:
		if config.TargetScore != 0 {
			return fmt.Errorf("arcade_survival does not accept target_score")
		}
	case PresetScoreAttack:
		if config.TargetScore <= 0 {
			return fmt.Errorf("score_attack target_score must be positive")
		}
	default:
		return fmt.Errorf("unknown mode preset %q", config.PresetID)
	}
	if config.InfiniteLives {
		if config.StartingLives != 0 {
			return fmt.Errorf("infinite lives does not accept starting_lives")
		}
	} else if config.StartingLives <= 0 {
		return fmt.Errorf("starting_lives must be positive")
	}
	return nil
}

func Resolve(config RoomModeConfig, teamConfig teams.Config) (ResolvedMatchRules, error) {
	config = NormalizeRoomModeConfig(config)
	if err := ValidateRoomModeConfig(config); err != nil {
		return ResolvedMatchRules{}, err
	}
	if err := teams.ValidateConfig(teamConfig); err != nil {
		return ResolvedMatchRules{}, fmt.Errorf("invalid team configuration: %w", err)
	}

	lifePolicy := lives.NewBaselinePolicy()
	if config.InfiniteLives {
		lifePolicy.Model = lives.LifeModelInfinite
		lifePolicy.StartingLives = 0
	} else {
		lifePolicy.StartingLives = config.StartingLives
	}

	resolved := ResolvedMatchRules{
		PresetID:                 config.PresetID,
		ModeID:                   ModeArcadeSurvival,
		TeamConfig:               teamConfig,
		LivesPolicy:              lifePolicy,
		AwardPolicyID:            awards.StandardPolicyID,
		RankingMetric:            RankingNone,
		MatchEndPrecedence:       []MatchEndCondition{EndNoActivePlayers},
		ResultPolicy:             ResultFinalFacts,
		PlayerDamageEnabled:      false,
		PlayerSpawnProfileID:     lifePolicy.SpawnProfileID,
		EncounterSpawnProfileIDs: []string{EncounterAsteroidsV1},
		InGameJoiningAllowed:     false,
		ProgressionEligible:      true,
		FreezeGameplayOnEnd:      true,
	}
	if config.PresetID == PresetScoreAttack {
		resolved.ModeID = ModeScoreAttack
		resolved.ObjectivePolicy = ObjectivePolicy{DefinitionID: "score_attack_target_v1", TargetScore: config.TargetScore}
		resolved.RankingMetric = RankingCompletionTime
		resolved.MatchEndPrecedence = []MatchEndCondition{EndTargetScoreReached, EndNoActivePlayers}
		resolved.ResultPolicy = ResultScoreAttack
	}
	return CloneResolvedMatchRules(resolved), nil
}
