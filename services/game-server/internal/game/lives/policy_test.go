package lives

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
)

func TestBaselinePolicyAppliesFiniteDeath(t *testing.T) {
	policy := NewBaselinePolicy()
	if policy.StartingLives != constants.PlayerStartingLives {
		t.Fatalf("expected starting lives %d, got %d", constants.PlayerStartingLives, policy.StartingLives)
	}
	if policy.RespawnDelay != constants.PlayerRespawnDelay {
		t.Fatalf("expected respawn delay %v, got %v", constants.PlayerRespawnDelay, policy.RespawnDelay)
	}

	result := policy.ApplyDeath(policy.StartingLives, false)
	if result.RemainingLives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected %d remaining lives, got %d", constants.PlayerStartingLives-1, result.RemainingLives)
	}
	if result.RespawnCooldown != constants.PlayerRespawnDelay {
		t.Fatalf("expected respawn cooldown %v, got %v", constants.PlayerRespawnDelay, result.RespawnCooldown)
	}
}

func TestBaselinePolicyKeepsInfiniteLivesAndBlocksFinalRespawn(t *testing.T) {
	policy := NewBaselinePolicy()

	result := policy.ApplyDeath(policy.StartingLives, true)
	if result.RemainingLives != policy.StartingLives {
		t.Fatalf("expected infinite lives to remain %d, got %d", policy.StartingLives, result.RemainingLives)
	}
	if result.RespawnCooldown != policy.RespawnDelay {
		t.Fatalf("expected infinite-lives respawn cooldown %v, got %v", policy.RespawnDelay, result.RespawnCooldown)
	}

	finalDeath := policy.ApplyDeath(1, false)
	if finalDeath.RemainingLives != 0 {
		t.Fatalf("expected final death to leave 0 lives, got %d", finalDeath.RemainingLives)
	}
	if finalDeath.RespawnCooldown != 0 {
		t.Fatalf("expected final death cooldown 0, got %v", finalDeath.RespawnCooldown)
	}
	if policy.EvaluateRespawn("player-1", playerstate.StatusPendingRespawn, finalDeath.RemainingLives, finalDeath.RespawnCooldown).Accepted {
		t.Fatal("expected respawn to be blocked after final death")
	}
}

func TestBaselinePolicyRetainsRespawnCooldownEligibility(t *testing.T) {
	policy := NewBaselinePolicy()

	if !policy.EvaluateRespawn("player-1", playerstate.StatusPendingRespawn, 1, 0).Accepted {
		t.Fatal("expected player with lives and no cooldown to respawn")
	}
	if policy.EvaluateRespawn("player-1", playerstate.StatusPendingRespawn, 1, policy.RespawnDelay).Accepted {
		t.Fatal("expected active respawn cooldown to block respawn")
	}
	if policy.EvaluateRespawn("player-1", playerstate.StatusPendingRespawn, 0, 0).Accepted {
		t.Fatal("expected no lives to block respawn")
	}
}
