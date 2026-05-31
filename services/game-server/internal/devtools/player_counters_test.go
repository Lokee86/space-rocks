package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestHandleDebugSetScoreSetsExactScore(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: 42,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, playerID, 42)
}

func TestHandleDebugSetScoreClampsNegativeToZero(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: -10,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, playerID, 0)
}

func TestHandleDebugAddScoreIncreasesScore(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: 10,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddScore,
		Amount: 5,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, playerID, 15)
}

func TestHandleDebugAddScoreCanReduceScore(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: 10,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddScore,
		Amount: -3,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, playerID, 7)
}

func TestHandleDebugAddScoreClampsBelowZero(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: 2,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddScore,
		Amount: -5,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, playerID, 0)
}

func TestHandleDebugSetLivesSetsExactLives(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetLives,
		Lives: 5,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, playerID, 5)
}

func TestHandleDebugSetLivesClampsNegativeToZero(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetLives,
		Lives: -10,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, playerID, 0)
}

func TestHandleDebugAddLivesIncreasesLives(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetLives,
		Lives: 3,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddLives,
		Amount: 2,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, playerID, 5)
}

func TestHandleDebugAddLivesCanReduceLives(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetLives,
		Lives: 3,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddLives,
		Amount: -1,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, playerID, 2)
}

func TestHandleDebugAddLivesClampsBelowZero(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	HandleCommand(target, playerID, DebugCommand{
		Type:  PacketTypeDebugSetLives,
		Lives: 1,
	})

	ok := HandleCommand(target, playerID, DebugCommand{
		Type:   PacketTypeDebugAddLives,
		Amount: -5,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, playerID, 0)
}

func TestHandleDebugSetScoreTargetsAnotherPlayer(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugSetScore,
		TargetPlayerID: targetPlayerID,
		Score:          44,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, targetPlayerID, 44)
}

func TestHandleDebugAddScoreTargetsAnotherPlayer(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugSetScore,
		TargetPlayerID: targetPlayerID,
		Score:          10,
	})

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugAddScore,
		TargetPlayerID: targetPlayerID,
		Amount:         6,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, targetPlayerID, 16)
}

func TestHandleDebugSetLivesTargetsAnotherPlayer(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugSetLives,
		TargetPlayerID: targetPlayerID,
		Lives:          7,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, targetPlayerID, 7)
}

func TestHandleDebugAddLivesTargetsAnotherPlayer(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	targetPlayerID := target.AddPlayer()

	HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugSetLives,
		TargetPlayerID: targetPlayerID,
		Lives:          3,
	})

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:           PacketTypeDebugAddLives,
		TargetPlayerID: targetPlayerID,
		Amount:         2,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, targetPlayerID, 5)
}

func TestHandleDebugSetScoreFallsBackToCallingPlayerWhenTargetEmpty(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:  PacketTypeDebugSetScore,
		Score: 31,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, callerID, 31)
	assertPlayerPacketScore(t, target, otherPlayerID, 0)
}

func assertPlayerPacketScore(t *testing.T, target *game.Game, playerID string, expected int) {
	t.Helper()

	packet := target.StatePacket(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if player.Score != expected {
		t.Fatalf("expected player score %d, got %d", expected, player.Score)
	}
}

func assertPlayerPacketLives(t *testing.T, target *game.Game, playerID string, expected int) {
	t.Helper()

	packet := target.StatePacket(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if packet.Lives != expected {
		t.Fatalf("expected packet lives %d, got %d", expected, packet.Lives)
	}
	if player.Lives != expected {
		t.Fatalf("expected player lives %d, got %d", expected, player.Lives)
	}
}
