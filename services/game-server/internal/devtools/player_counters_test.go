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

func TestHandleDebugSetScoreTargetsAllPlayers(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugSetScore,
		TargetScope: targetScopeAllPlayers,
		Score:       44,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, callerID, 44)
	assertPlayerPacketScore(t, target, otherPlayerID, 44)
}

func TestHandleDebugAddScoreTargetsAllPlayers(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugSetScore,
		TargetScope: targetScopeAllPlayers,
		Score:       10,
	})

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugAddScore,
		TargetScope: targetScopeAllPlayers,
		Amount:      6,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketScore(t, target, callerID, 16)
	assertPlayerPacketScore(t, target, otherPlayerID, 16)
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

func TestHandleDebugSetLivesTargetsAllPlayers(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugSetLives,
		TargetScope: targetScopeAllPlayers,
		Lives:       7,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, callerID, 7)
	assertPlayerPacketLives(t, target, otherPlayerID, 7)
}

func TestHandleDebugAddLivesTargetsAllPlayers(t *testing.T) {
	target := game.New()
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugSetLives,
		TargetScope: targetScopeAllPlayers,
		Lives:       3,
	})

	ok := HandleCommand(target, callerID, DebugCommand{
		Type:        PacketTypeDebugAddLives,
		TargetScope: targetScopeAllPlayers,
		Amount:      2,
	})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	assertPlayerPacketLives(t, target, callerID, 5)
	assertPlayerPacketLives(t, target, otherPlayerID, 5)
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
	session, ok := packet.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected player session %q in state packet", playerID)
	}
	if session.Score != expected {
		t.Fatalf("expected player score %d, got %d", expected, session.Score)
	}
}

func assertPlayerPacketLives(t *testing.T, target *game.Game, playerID string, expected int) {
	t.Helper()

	packet := target.StatePacket(playerID)
	session, ok := packet.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected player session %q in state packet", playerID)
	}
	if packet.Lives != expected {
		t.Fatalf("expected packet lives %d, got %d", expected, packet.Lives)
	}
	if session.Lives != expected {
		t.Fatalf("expected player lives %d, got %d", expected, session.Lives)
	}
}
