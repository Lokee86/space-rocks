package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestAddPlayerDefaultsToNoTeam(t *testing.T) {
	gameInstance := NewWithSeed(1)
	playerID := gameInstance.AddPlayer()

	if got := gameInstance.PlayerTeam(playerID); got != teams.NoTeam {
		t.Fatalf("direct AddPlayer team = %q, want NoTeam", got)
	}
}

func TestExplicitTeamAssignmentReachesGameOwnership(t *testing.T) {
	gameInstance := NewWithSeed(2)
	gameInstance.SetTeamStructure(teams.StructureCustom)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team3)

	if got := gameInstance.PlayerTeam(playerID); got != teams.Team3 {
		t.Fatalf("custom team = %q, want %q", got, teams.Team3)
	}
	if facts := gameInstance.PlayerMatchFacts(); len(facts) != 1 || facts[0].TeamID != teams.Team3 {
		t.Fatalf("custom team fact = %#v, want Team3", facts)
	}
}

func TestCoOpPlayerReceivesSharedTeam(t *testing.T) {
	gameInstance := NewWithSeed(3)
	gameInstance.SetTeamStructure(teams.StructureCoOp)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team1)

	if got := gameInstance.PlayerTeam(playerID); got != teams.Team1 {
		t.Fatalf("co-op team = %q, want %q", got, teams.Team1)
	}
}

func TestPlayerRelationshipUsesGameOwnedMembership(t *testing.T) {
	gameInstance := NewWithSeed(4)
	gameInstance.SetTeamStructure(teams.StructureCustom)
	first := gameInstance.AddPlayerWithTeam(teams.Team1)
	second := gameInstance.AddPlayerWithTeam(teams.Team1)
	third := gameInstance.AddPlayerWithTeam(teams.Team2)

	if got := gameInstance.PlayerRelationship(first, second); got != teams.RelationshipSameTeam {
		t.Fatalf("same-team relationship = %q, want %q", got, teams.RelationshipSameTeam)
	}
	if got := gameInstance.PlayerRelationship(first, third); got != teams.RelationshipOpposing {
		t.Fatalf("different-team relationship = %q, want %q", got, teams.RelationshipOpposing)
	}
}

func TestRespawnPreservesTeamMembership(t *testing.T) {
	gameInstance := NewWithSeed(5)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team2)
	delete(gameInstance.entities.Players, playerID)

	gameInstance.respawnPlayer(playerID)

	if got := gameInstance.PlayerTeam(playerID); got != teams.Team2 {
		t.Fatalf("respawn team = %q, want %q", got, teams.Team2)
	}
}

func TestRemovePlayerPreservesHistoricalTeamFact(t *testing.T) {
	gameInstance := NewWithSeed(6)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team4)

	gameInstance.RemovePlayer(playerID)

	if got := gameInstance.PlayerTeam(playerID); got != teams.Team4 {
		t.Fatalf("removed player team = %q, want %q", got, teams.Team4)
	}
	facts := gameInstance.PlayerMatchFacts()
	if len(facts) != 1 || facts[0].TeamID != teams.Team4 {
		t.Fatalf("removed player facts = %#v, want Team4", facts)
	}
}

func TestRollbackRemovesTeamMembershipAndParticipantState(t *testing.T) {
	gameInstance := NewWithSeed(7)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team5)

	gameInstance.RollbackPlayerAdd(playerID)

	if got := gameInstance.PlayerTeam(playerID); got != teams.NoTeam {
		t.Fatalf("rolled-back player team = %q, want NoTeam", got)
	}
	if facts := gameInstance.PlayerMatchFacts(); len(facts) != 0 {
		t.Fatalf("rolled-back player facts = %#v, want empty", facts)
	}
}

func TestNewGameDoesNotInheritPreviousMatchMembership(t *testing.T) {
	previous := NewWithSeed(8)
	previous.AddPlayerWithTeam(teams.Team6)

	fresh := NewWithSeed(9)
	if got := fresh.PlayerTeam("player-1"); got != teams.NoTeam {
		t.Fatalf("fresh game team = %q, want NoTeam", got)
	}
	if facts := fresh.PlayerMatchFacts(); len(facts) != 0 {
		t.Fatalf("fresh game facts = %#v, want empty", facts)
	}
}
