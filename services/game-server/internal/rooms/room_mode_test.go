package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestRoomStoresAndLocksResolvedScoreAttackRulesAtStart(t *testing.T) {
	room, err := NewRoomWithConfig("room", RoomStateLobby, nil, RoomCreationConfig{
		ModeConfig: modes.RoomModeConfig{PresetID: modes.PresetScoreAttack, StartingLives: 5, TargetScore: 750},
		TeamConfig: teams.Config{Structure: teams.StructureFFA},
		MaxPlayers: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	room.SetJoinable(false)
	room.AddMember(NewRoomMember("session-owner"))
	before := room.SnapshotForSession("session-owner")
	if before.ModeLocked || before.ModeConfig.PresetID != modes.PresetScoreAttack {
		t.Fatalf("before = %+v", before)
	}

	if roomErr := room.StartSinglePlayerGame(game.New); roomErr != nil {
		t.Fatal(roomErr)
	}
	after := room.SnapshotForSession("session-owner")
	if !after.ModeLocked || after.ResolvedModeID != string(modes.ModeScoreAttack) {
		t.Fatalf("after = %+v", after)
	}
	resolved, ok := room.ResolvedMatchRules()
	if !ok || resolved.ObjectivePolicy.TargetScore != 750 || resolved.LivesPolicy.StartingLives != 5 {
		t.Fatalf("resolved = %+v, ok = %v", resolved, ok)
	}
	if gameRules := room.GameInstance().ResolvedMatchRules(); gameRules.ModeID != modes.ModeScoreAttack || gameRules.ObjectivePolicy.TargetScore != 750 {
		t.Fatalf("game rules = %+v", gameRules)
	}
	room.GameInstance().Stop()
}

func TestRoomCreationRejectsInvalidModeOptions(t *testing.T) {
	_, err := NewRoomWithConfig("room", RoomStateLobby, nil, RoomCreationConfig{
		ModeConfig: modes.RoomModeConfig{PresetID: modes.PresetArcadeSurvival, StartingLives: 3, TargetScore: 10},
		TeamConfig: teams.Config{Structure: teams.StructureFFA},
	})
	if err == nil {
		t.Fatal("expected invalid mode options")
	}
}
