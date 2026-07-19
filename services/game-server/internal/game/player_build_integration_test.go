package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func TestPlayerSessionUsesResolvedBuildForSpawnAndReset(t *testing.T) {
	session := newPlayerSession("player-1", physics.Vector2{})
	build := playerbuild.DefaultResolvedBuild("player-1")
	build.ShipStats.MaxShields = 40
	build.ShieldPolicy = playerbuild.ShieldPolicy{MaxShields: 40, StartsFull: true}
	build.StartingEquipmentState.PrimaryAmmo = 7
	if err := session.ApplyResolvedBuild(build); err != nil {
		t.Fatalf("apply resolved build: %v", err)
	}

	ship := session.NewShip(physics.Vector2{X: 10, Y: 20})
	if ship.Shields != 40 {
		t.Fatalf("expected full resolved shields, got %d", ship.Shields)
	}
	if ship.WeaponState.Primary.AmmoRemaining != 7 {
		t.Fatalf("expected resolved starting ammo, got %d", ship.WeaponState.Primary.AmmoRemaining)
	}

	session.PlayerArmory.Secondary = weapons.Equipped{ID: weapons.Torpedo, AmmoPolicy: weapons.LimitedAmmo}
	respawn := session.NewRespawnShip(physics.Vector2{}, lives.RestorationPolicy{
		Health:  lives.RestorationFull,
		Shields: lives.RestorationFull,
		Loadout: lives.LoadoutReset,
	})
	if respawn.ShipWeapons.Secondary.ID != "" {
		t.Fatalf("reset should restore resolved build armory, got %q", respawn.ShipWeapons.Secondary.ID)
	}
	if respawn.Shields != 40 {
		t.Fatalf("respawn should restore resolved shields, got %d", respawn.Shields)
	}
}

func TestPlayerSessionRejectsAnotherPlayersResolvedBuild(t *testing.T) {
	session := newPlayerSession("player-1", physics.Vector2{})
	if err := session.ApplyResolvedBuild(playerbuild.DefaultResolvedBuild("player-2")); err == nil {
		t.Fatal("expected mismatched player build rejection")
	}
}
