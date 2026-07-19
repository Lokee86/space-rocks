package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
)

func TestApplyPlayerBuildReplacesProvisionalBuildBeforeMatchStart(t *testing.T) {
	game := NewWithSeed(41)
	playerID := game.AddPlayer()
	build := playerbuild.DefaultResolvedBuild(playerID)
	build.ShipStats.MaxShields = 30
	build.ShieldPolicy = playerbuild.ShieldPolicy{MaxShields: 30, StartsFull: true}

	if err := game.ApplyPlayerBuild(playerID, build); err != nil {
		t.Fatalf("apply player build: %v", err)
	}
	stored, ok := game.PlayerResolvedBuild(playerID)
	if !ok || stored.ShipStats.MaxShields != 30 {
		t.Fatalf("unexpected stored build: %#v, ok=%v", stored, ok)
	}
	ship := game.entities.Players[playerID]
	if ship == nil || ship.Shields != 30 || ship.Stats.MaxShields != 30 {
		t.Fatalf("active provisional ship did not receive resolved build: %#v", ship)
	}
}

func TestApplyPlayerBuildRejectsMidMatchChanges(t *testing.T) {
	game := NewWithSeed(42)
	playerID := game.AddPlayer()
	game.Step(1.0 / 60.0)

	if err := game.ApplyPlayerBuild(playerID, playerbuild.DefaultResolvedBuild(playerID)); err == nil {
		t.Fatal("expected mid-match build change rejection")
	}
}

func TestPlayerResolvedBuildReturnsClone(t *testing.T) {
	game := NewWithSeed(43)
	playerID := game.AddPlayer()
	first, ok := game.PlayerResolvedBuild(playerID)
	if !ok {
		t.Fatal("expected player build")
	}
	first.WeaponPointLayout[playerbuild.Primary1] = playerbuild.PointNone
	second, _ := game.PlayerResolvedBuild(playerID)
	if second.WeaponPointLayout[playerbuild.Primary1] != playerbuild.PointHardpoint {
		t.Fatal("player build accessor leaked mutable state")
	}
}
