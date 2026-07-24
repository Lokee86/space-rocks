package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

type debugSpawnPlayerTargetFake struct {
	Target
	reservedPlayerID string
	ensuredPlayerID  string
	spawnedPlayerID  string
	botPlayerID      string
}

func (target *debugSpawnPlayerTargetFake) ReservePlayerID(playerID string) bool {
	target.reservedPlayerID = playerID
	return true
}

func (target *debugSpawnPlayerTargetFake) EnsurePlayerSession(playerID string, _ physics.Vector2) bool {
	target.ensuredPlayerID = playerID
	return true
}

func (target *debugSpawnPlayerTargetFake) SpawnPlayerShip(playerID string, _ physics.Vector2, _ runtimepkg.ClientConfig) bool {
	target.spawnedPlayerID = playerID
	return true
}

func (target *debugSpawnPlayerTargetFake) EnableBotPlayer(playerID string) bool {
	target.botPlayerID = playerID
	return true
}

func TestApplyDebugSpawnPlayerEnablesBotLogic(t *testing.T) {
	target := &debugSpawnPlayerTargetFake{}

	playerID, _, ok := applyDebugSpawnPlayer(target, SpawnEntityRequest{
		TargetPlayerID: "player-2",
		X:              100,
		Y:              200,
	})
	if !ok {
		t.Fatal("expected debug player spawn to succeed")
	}
	if playerID != "player-2" {
		t.Fatalf("playerID = %q, want %q", playerID, "player-2")
	}
	if target.reservedPlayerID != playerID {
		t.Fatalf("reserved player = %q, want %q", target.reservedPlayerID, playerID)
	}
	if target.ensuredPlayerID != playerID {
		t.Fatalf("ensured player = %q, want %q", target.ensuredPlayerID, playerID)
	}
	if target.spawnedPlayerID != playerID {
		t.Fatalf("spawned player = %q, want %q", target.spawnedPlayerID, playerID)
	}
	if target.botPlayerID != playerID {
		t.Fatalf("bot player = %q, want %q", target.botPlayerID, playerID)
	}
}
