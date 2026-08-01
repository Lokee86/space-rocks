package modes

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func DefaultRoomModeConfig() RoomModeConfig {
	return RoomModeConfig{
		PresetID:      PresetArcadeSurvival,
		StartingLives: constants.PlayerStartingLives,
	}
}

func NormalizeRoomModeConfig(config RoomModeConfig) RoomModeConfig {
	if config.PresetID == "" {
		config.PresetID = PresetArcadeSurvival
	}
	if config.PresetID == PresetDeathmatch || config.PresetID == PresetTeamDeathmatch {
		config.InfiniteLives = true
		config.StartingLives = 0
	} else if !config.InfiniteLives && config.StartingLives == 0 {
		config.StartingLives = constants.PlayerStartingLives
	}
	if config.PresetID == PresetScoreAttack && config.TargetScore == 0 {
		config.TargetScore = DefaultScoreAttackTarget
	}
	if (config.PresetID == PresetDeathmatch || config.PresetID == PresetTeamDeathmatch) && config.TargetKills == 0 {
		config.TargetKills = DefaultDeathmatchTargetKills
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
		if config.TargetKills != 0 {
			return fmt.Errorf("arcade_survival does not accept target_kills")
		}
	case PresetScoreAttack:
		if config.TargetScore <= 0 {
			return fmt.Errorf("score_attack target_score must be positive")
		}
		if config.TargetKills != 0 {
			return fmt.Errorf("score_attack does not accept target_kills")
		}
	case PresetDeathmatch, PresetTeamDeathmatch:
		if config.TargetKills <= 0 {
			return fmt.Errorf("deathmatch target_kills must be positive")
		}
		if config.TargetScore != 0 {
			return fmt.Errorf("deathmatch does not accept target_score")
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
	if config.PresetID == PresetDeathmatch && teamConfig.Structure != teams.StructureFFA {
		return ResolvedMatchRules{}, fmt.Errorf("deathmatch requires free-for-all teams")
	}
	if config.PresetID == PresetTeamDeathmatch && teamConfig.Structure != teams.StructureCustom && teamConfig.Structure != teams.StructureAutoBalanced {
		return ResolvedMatchRules{}, fmt.Errorf("team deathmatch requires custom or auto-balanced teams")
	}

	resolved := ResolvedMatchRules{
		PresetID:                 config.PresetID,
		ModeID:                   ModeArcadeSurvival,
		TeamConfig:               teamConfig,
		LivesPolicy:              LivesPolicy{StartingLives: config.StartingLives, InfiniteLives: config.InfiniteLives},
		AwardPolicyID:            StandardAwardPolicyID,
		RankingMetric:            RankingNone,
		MatchEndPrecedence:       []MatchEndCondition{EndNoActivePlayers},
		ResultPolicy:             ResultFinalFacts,
		PlayerDamageEnabled:      false,
		PlayerSpawnProfileID:     DefaultSpawnProfileID,
		EncounterSpawnProfileIDs: []string{EncounterAsteroidsV1},
		InGameJoiningAllowed:     false,
		ProgressionEligible:      true,
		FreezeGameplayOnEnd:      true,
	}
	switch config.PresetID {
	case PresetScoreAttack:
		resolved.ModeID = ModeScoreAttack
		resolved.ObjectivePolicy = ObjectivePolicy{DefinitionID: "score_attack_target_v1", TargetScore: config.TargetScore}
		resolved.RankingMetric = RankingCompletionTime
		resolved.MatchEndPrecedence = []MatchEndCondition{EndTargetScoreReached, EndNoActivePlayers}
		resolved.ResultPolicy = ResultScoreAttack
	case PresetDeathmatch, PresetTeamDeathmatch:
		resolved.ModeID = ModeDeathmatch
		resolved.LivesPolicy = LivesPolicy{InfiniteLives: true}
		resolved.ObjectivePolicy = ObjectivePolicy{DefinitionID: "deathmatch_kill_target_v1", TargetKills: config.TargetKills}
		resolved.RankingMetric = RankingKills
		resolved.MatchEndPrecedence = []MatchEndCondition{EndTargetKillsReached, EndNoActivePlayers}
		resolved.ResultPolicy = ResultDeathmatch
		resolved.PlayerDamageEnabled = true
		resolved.EncounterSpawnProfileIDs = nil
		if config.PresetID == PresetTeamDeathmatch {
			resolved.ObjectivePolicy.DefinitionID = "team_deathmatch_kill_target_v1"
			resolved.RankingMetric = RankingTeamKills
			resolved.ResultPolicy = ResultTeamDeathmatch
			resolved.TeamScoreEnabled = true
		}
	}
	return CloneResolvedMatchRules(resolved), nil
}
