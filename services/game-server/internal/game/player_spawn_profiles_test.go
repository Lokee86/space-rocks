package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestDeathmatchSpawnProfileSeparatesOpponents(t *testing.T) {
	game := NewWithSeed(17)
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetDeathmatch, TargetKills: 10},
		teams.Config{Structure: teams.StructureFFA},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}

	firstID := game.AddPlayer()
	secondID := game.AddPlayer()
	first := game.entities.Players[firstID]
	second := game.entities.Players[secondID]
	distance := space.Distance(
		physics.Vector2{X: first.X, Y: first.Y},
		physics.Vector2{X: second.X, Y: second.Y},
	)
	if distance < space.DefaultBounds().Height*0.35 {
		t.Fatalf("opponent spawn distance = %.2f, want at least %.2f", distance, space.DefaultBounds().Height*0.35)
	}
}

func TestDeathmatchRespawnMovesAwayFromPreviousSpawnAndOpponent(t *testing.T) {
	game := NewWithSeed(29)
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetDeathmatch, TargetKills: 10},
		teams.Config{Structure: teams.StructureFFA},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}

	respawningID := game.AddPlayer()
	opponentID := game.AddPlayer()
	previous := game.playerSessions[respawningID].SpawnPosition
	delete(game.entities.Players, respawningID)

	plan := game.planPlayerRespawn(game.playerSessions[respawningID])
	opponent := game.entities.Players[opponentID]
	bounds := space.DefaultBounds()
	if distance := space.WrappedDistance(plan.Position, previous, bounds); distance < bounds.Height*0.1 {
		t.Fatalf("respawn moved %.2f from previous spawn, want at least %.2f", distance, bounds.Height*0.1)
	}
	if distance := space.WrappedDistance(plan.Position, physics.Vector2{X: opponent.X, Y: opponent.Y}, bounds); distance < bounds.Height*0.35 {
		t.Fatalf("respawn distance from opponent = %.2f, want at least %.2f", distance, bounds.Height*0.35)
	}
}

func TestTeamDeathmatchSpawnGroupsTeammatesAwayFromEnemy(t *testing.T) {
	game := NewWithSeed(41)
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetTeamDeathmatch, TargetKills: 10},
		teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 2},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}

	teammateAID := game.AddPlayerWithTeam(teams.Team1)
	teammateBID := game.AddPlayerWithTeam(teams.Team1)
	enemyID := game.AddPlayerWithTeam(teams.Team2)
	teammateA := game.entities.Players[teammateAID]
	teammateB := game.entities.Players[teammateBID]
	enemy := game.entities.Players[enemyID]
	teammateDistance := space.Distance(
		physics.Vector2{X: teammateA.X, Y: teammateA.Y},
		physics.Vector2{X: teammateB.X, Y: teammateB.Y},
	)
	enemyDistance := min(
		space.Distance(physics.Vector2{X: teammateA.X, Y: teammateA.Y}, physics.Vector2{X: enemy.X, Y: enemy.Y}),
		space.Distance(physics.Vector2{X: teammateB.X, Y: teammateB.Y}, physics.Vector2{X: enemy.X, Y: enemy.Y}),
	)
	if teammateDistance >= enemyDistance {
		t.Fatalf("teammate distance %.2f must be less than nearest enemy distance %.2f", teammateDistance, enemyDistance)
	}
}
