package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestCreateLobbyRoomWithConfigStoresTeamConfigurationAndCapacity(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoomWithConfig(RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 3},
		MaxPlayers: 4,
	})
	if err != nil {
		t.Fatalf("create configured room: %v", err)
	}
	if room.TeamConfig() != (teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 3}) || room.MaxPlayers != 4 {
		t.Fatalf("unexpected room configuration: %+v, capacity %d", room.TeamConfig(), room.MaxPlayers)
	}
}

func TestCreateLobbyRoomWithConfigRejectsInvalidConfigurationAndCapacity(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	if _, err := manager.CreateLobbyRoomWithConfig(RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 1},
		MaxPlayers: 4,
	}); err == nil {
		t.Fatal("expected invalid team configuration rejection")
	}
	if _, err := manager.CreateLobbyRoomWithConfig(RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureFFA},
		MaxPlayers: MaxPlayersPerRoom + 1,
	}); err == nil {
		t.Fatal("expected invalid capacity rejection")
	}
}

func TestDefaultRoomCreationPreservesFFAAndMaximumCapacity(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create default room: %v", err)
	}
	if room.TeamConfig().Structure != teams.StructureFFA || room.MaxPlayers != MaxPlayersPerRoom {
		t.Fatalf("unexpected defaults: %+v, capacity %d", room.TeamConfig(), room.MaxPlayers)
	}
}
