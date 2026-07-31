package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildRoomSnapshotProjectsSelectedAndResolvedMode(t *testing.T) {
	room, err := rooms.NewRoomWithConfig("room", rooms.RoomStateLobby, nil, rooms.RoomCreationConfig{
		ModeConfig: modes.RoomModeConfig{
			PresetID: modes.PresetScoreAttack, StartingLives: 5, TargetScore: 750,
		},
		TeamConfig: teams.Config{Structure: teams.StructureFFA},
		MaxPlayers: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	room.SetJoinable(false)
	room.AddMember(rooms.NewRoomMember("session-owner"))

	before := BuildRoomSnapshot(room, "session-owner")
	if before.PresetID != string(modes.PresetScoreAttack) || before.ModeID != "" || before.ModeLocked {
		t.Fatalf("before = %+v", before)
	}
	if before.StartingLives != 5 || before.InfiniteLives || before.TargetScore != 750 {
		t.Fatalf("before options = %+v", before)
	}

	if roomErr := room.StartSinglePlayerGame(game.New); roomErr != nil {
		t.Fatal(roomErr)
	}
	after := BuildRoomSnapshot(room, "session-owner")
	if after.ModeID != string(modes.ModeScoreAttack) || !after.ModeLocked {
		t.Fatalf("after = %+v", after)
	}
	room.GameInstance().Stop()
}
