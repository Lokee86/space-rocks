package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/participation"
)

func TestGameUsesConfiguredAFKTimeoutAndExpiresPendingRespawn(t *testing.T) {
	game, err := NewWithPolicies(lives.NewBaselinePolicy(), participation.AFKPolicy{Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	playerID := game.AddPlayer()
	game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
	game.stepPlayerSessions(1)
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != "removed" {
		t.Fatalf("AFK expiry status = %q, ok=%t", status, ok)
	}
	if _, ok := game.playerSessions[playerID]; ok {
		t.Fatal("AFK expiry retained active session")
	}
	if _, ok := game.participantRecords[playerID]; !ok {
		t.Fatal("AFK expiry removed historical participant record")
	}
}

func TestGameAcceptedInputResetsAFKActivity(t *testing.T) {
	game, err := NewWithPolicies(lives.NewBaselinePolicy(), participation.AFKPolicy{Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	playerID := game.AddPlayer()
	game.stepPlayerSessions(0.9)
	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeInput})
	game.stepPlayerSessions(0.9)
	if _, ok := game.playerSessions[playerID]; !ok {
		t.Fatal("accepted input did not reset AFK timer")
	}
	game.stepPlayerSessions(0.1)
	if _, ok := game.playerSessions[playerID]; ok {
		t.Fatal("AFK timer did not expire after reset allowance")
	}
}
