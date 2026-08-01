package modes

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

type PresetID string

type ModeID string

const (
	PresetArcadeSurvival PresetID = "arcade_survival"
	PresetScoreAttack    PresetID = "score_attack"
	PresetDeathmatch     PresetID = "deathmatch"

	ModeArcadeSurvival ModeID = "arcade_survival"
	ModeScoreAttack    ModeID = "score_attack"
	ModeDeathmatch     ModeID = "deathmatch"

	DefaultScoreAttackTarget     = 1000
	DefaultDeathmatchTargetKills = 10
	EncounterAsteroidsV1         = "playercentric_asteroids_v1"
	DefaultSpawnProfileID        = "baseline"
	StandardAwardPolicyID        = "standard"
)

type RoomModeConfig struct {
	PresetID      PresetID
	StartingLives int
	InfiniteLives bool
	TargetScore   int
	TargetKills   int
}

type LivesPolicy struct {
	StartingLives int
	InfiniteLives bool
}

type ObjectivePolicy struct {
	DefinitionID string
	TargetScore  int
	TargetKills  int
}

type RankingMetric string

const (
	RankingNone           RankingMetric = "none"
	RankingCompletionTime RankingMetric = "completion_time"
	RankingKills          RankingMetric = "kills"
)

type MatchEndCondition string

const (
	EndTargetScoreReached MatchEndCondition = "target_score_reached"
	EndTargetKillsReached MatchEndCondition = "target_kills_reached"
	EndNoActivePlayers    MatchEndCondition = "no_active_participants"
)

type ResultPolicy string

const (
	ResultFinalFacts  ResultPolicy = "final_facts"
	ResultScoreAttack ResultPolicy = "score_attack"
	ResultDeathmatch  ResultPolicy = "deathmatch"
)

type ResolvedMatchRules struct {
	PresetID                 PresetID
	ModeID                   ModeID
	TeamConfig               teams.Config
	LivesPolicy              LivesPolicy
	AwardPolicyID            string
	ObjectivePolicy          ObjectivePolicy
	RankingMetric            RankingMetric
	MatchEndPrecedence       []MatchEndCondition
	ResultPolicy             ResultPolicy
	PlayerDamageEnabled      bool
	PlayerSpawnProfileID     string
	EncounterSpawnProfileIDs []string
	InGameJoiningAllowed     bool
	ProgressionEligible      bool
	FreezeGameplayOnEnd      bool
}

type PlayerFact struct {
	ID             string
	TeamID         teams.ID
	Status         rules.PlayerParticipationStatus
	Active         bool
	Score          int
	CompletionTime float64
	SuccessOrder   int
}

type MatchFacts struct {
	Players         []PlayerFact
	HadParticipants bool
	Elapsed         float64
}

func CloneResolvedMatchRules(source ResolvedMatchRules) ResolvedMatchRules {
	clone := source
	clone.MatchEndPrecedence = append([]MatchEndCondition(nil), source.MatchEndPrecedence...)
	clone.EncounterSpawnProfileIDs = append([]string(nil), source.EncounterSpawnProfileIDs...)
	return clone
}
