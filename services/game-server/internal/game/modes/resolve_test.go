package modes

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestResolveArcadeSurvivalBaseline(t *testing.T) {
	resolved, err := Resolve(DefaultRoomModeConfig(), teams.Config{Structure: teams.StructureFFA})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ModeID != ModeArcadeSurvival || resolved.RankingMetric != RankingNone {
		t.Fatalf("resolved = %+v", resolved)
	}
	if resolved.PlayerSpawnProfileID != DefaultSpawnProfileID || len(resolved.EncounterSpawnProfileIDs) != 1 {
		t.Fatalf("spawn profiles = %+v", resolved)
	}
}

func TestResolveScoreAttackOverrides(t *testing.T) {
	resolved, err := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 4, TargetScore: 2500}, teams.Config{Structure: teams.StructureFFA})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ModeID != ModeScoreAttack || resolved.ObjectivePolicy.TargetScore != 2500 || resolved.RankingMetric != RankingCompletionTime {
		t.Fatalf("resolved = %+v", resolved)
	}
	if got := resolved.MatchEndPrecedence; len(got) != 2 || got[0] != EndTargetScoreReached {
		t.Fatalf("precedence = %+v", got)
	}
}

func TestResolveInfiniteLives(t *testing.T) {
	resolved, err := Resolve(RoomModeConfig{PresetID: PresetArcadeSurvival, InfiniteLives: true}, teams.Config{Structure: teams.StructureCoOp})
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.LivesPolicy.InfiniteLives || resolved.LivesPolicy.StartingLives != 0 {
		t.Fatalf("lives policy = %+v", resolved.LivesPolicy)
	}
}

func TestValidateRejectsTargetScoreForArcadeSurvival(t *testing.T) {
	err := ValidateRoomModeConfig(RoomModeConfig{PresetID: PresetArcadeSurvival, StartingLives: 3, TargetScore: 10})
	if err == nil {
		t.Fatal("expected invalid target score")
	}
}
