package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestDeathmatchProjectileKillAwardsOneKillAndEndsAtTarget(t *testing.T) {
	game := New()
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetDeathmatch, TargetKills: 1},
		teams.Config{Structure: teams.StructureFFA},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}

	attackerID := game.AddPlayer()
	victimID := game.AddPlayer()
	attacker := game.entities.Players[attackerID]
	victim := game.entities.Players[victimID]
	victim.X = attacker.X
	victim.Y = attacker.Y
	victim.Health = 1
	game.entities.Projectiles["kill-shot"] = runtime.NewBullet(
		"kill-shot",
		attackerID,
		physics.Vector2{X: victim.X, Y: victim.Y},
		0,
		physics.Vector2{},
		1,
	)

	game.handleBulletPlayerCollisions()

	if score, _ := game.currentPlayerScoreLocked(attackerID); score != 1 {
		t.Fatalf("attacker score = %d, want 1 kill", score)
	}
	if deaths := game.playerSessions[victimID].ShipDeaths; deaths != 1 {
		t.Fatalf("victim deaths = %d, want 1", deaths)
	}
	if !victim.IsPendingDespawn() {
		t.Fatal("victim should be pending despawn after fatal projectile hit")
	}
	decision := game.MatchDecision()
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalCompleted || decision.EndReason != string(modes.EndTargetKillsReached) {
		t.Fatalf("decision = %+v", decision)
	}
	if len(decision.WinningPlayerIDs) != 1 || decision.WinningPlayerIDs[0] != attackerID {
		t.Fatalf("winning players = %+v, want %q", decision.WinningPlayerIDs, attackerID)
	}
}

func TestDeathmatchProjectileCannotDamageOwner(t *testing.T) {
	game := New()
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetDeathmatch, TargetKills: 1},
		teams.Config{Structure: teams.StructureFFA},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}

	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	before := player.Health
	game.entities.Projectiles["self-shot"] = runtime.NewBullet(
		"self-shot",
		playerID,
		player.Position(),
		0,
		physics.Vector2{},
		1,
	)

	game.handleBulletPlayerCollisions()

	if player.Health != before {
		t.Fatalf("owner health = %d, want unchanged %d", player.Health, before)
	}
	if game.entities.Projectiles["self-shot"].IsPendingDespawn() {
		t.Fatal("self projectile should not be consumed by its owner")
	}
}

func TestArcadeSurvivalDoesNotEnableProjectilePlayerDamage(t *testing.T) {
	game := New()
	attackerID := game.AddPlayer()
	victimID := game.AddPlayer()
	attacker := game.entities.Players[attackerID]
	victim := game.entities.Players[victimID]
	victim.X = attacker.X
	victim.Y = attacker.Y
	before := victim.Health
	game.entities.Projectiles["disabled-shot"] = runtime.NewBullet(
		"disabled-shot",
		attackerID,
		victim.Position(),
		0,
		physics.Vector2{},
		1,
	)

	game.handleBulletPlayerCollisions()

	if victim.Health != before {
		t.Fatalf("victim health = %d, want unchanged %d", victim.Health, before)
	}
}
