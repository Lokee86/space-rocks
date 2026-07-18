package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestGameUsesStandardDamagePolicyByDefault(t *testing.T) {
	game := New()
	if game.damagePolicy.ID != damage.PolicyStandardV1 {
		t.Fatalf("damage policy = %q, want %q", game.damagePolicy.ID, damage.PolicyStandardV1)
	}
}

func TestGameDamageEligibilityUsesPvPAndTeamRelationships(t *testing.T) {
	game := New()
	sourceID := game.AddPlayer()
	targetID := game.AddPlayer()
	target := game.entities.Players[targetID]
	target.Health = 5
	target.Stats.MaxHealth = 5

	request := damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:            sourceID,
			EntityType:          damage.EntityTypePlayer,
			Cause:               damage.DamageCauseArea,
			ResponsiblePlayerID: sourceID,
		},
		Target: damage.DamageTarget{
			EntityID:   targetID,
			EntityType: damage.EntityTypePlayer,
			Health:     target.Health,
			MaxHealth:  target.Stats.MaxHealth,
		},
		Spec: damage.DamageSpec{Amount: 1, Type: damage.DamageTypeExplosive, Cause: damage.DamageCauseArea},
	}

	blocked := game.resolveDamageRequest(request)
	if blocked.Kind != damage.DamageResultKindBlocked || blocked.Reason != string(damage.EligibilityBlockPvPDisabled) {
		t.Fatalf("PvP-disabled result = %+v", blocked)
	}

	game.damagePolicy.PvPEnabled = true
	allowed := game.resolveDamageRequest(request)
	if allowed.Kind != damage.DamageResultKindDamage || allowed.AppliedToHealth != 1 {
		t.Fatalf("PvP-enabled result = %+v", allowed)
	}

	game.teamStructure = teams.StructureCustom
	game.playerSessions[sourceID].TeamID = teams.Team1
	game.playerSessions[targetID].TeamID = teams.Team1
	friendlyFire := game.resolveDamageRequest(request)
	if friendlyFire.Kind != damage.DamageResultKindBlocked || friendlyFire.Reason != string(damage.EligibilityBlockPlayerFriendlyFire) {
		t.Fatalf("friendly-fire result = %+v", friendlyFire)
	}
}

func TestGameDamageEligibilityBlocksInvulnerabilityExceptAuthorizedDevAdmin(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	player.Health = 5
	player.Stats.MaxHealth = 5
	player.DamageOptions.Invincible = true

	request := damage.DamageResolutionRequest{
		Source: damage.DamageSource{EntityID: "hazard-1", EntityType: damage.EntityTypeEnvironment, Cause: damage.DamageCauseHazard},
		Target: damage.DamageTarget{
			EntityID:     playerID,
			EntityType:   damage.EntityTypePlayer,
			Health:       player.Health,
			MaxHealth:    player.Stats.MaxHealth,
			Invulnerable: true,
		},
		Spec: damage.DamageSpec{Amount: 1, Type: damage.DamageTypeThermal, Cause: damage.DamageCauseHazard},
	}
	blocked := game.resolveDamageRequest(request)
	if blocked.Reason != string(damage.EligibilityBlockInvulnerable) {
		t.Fatalf("blocked result = %+v", blocked)
	}

	request.Source.Cause = damage.DamageCauseDebug
	request.Source.BypassInvulnerability = true
	request.Source.AuthorizedDevAdminSource = true
	allowed := game.resolveDamageRequest(request)
	if allowed.Kind != damage.DamageResultKindDamage || allowed.AppliedToHealth != 1 {
		t.Fatalf("authorized result = %+v", allowed)
	}
}

func TestGameDamageOverTimeTicksAndIsRemovedOnDeath(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	player.Health = 2
	player.Stats.MaxHealth = 2

	effect := damage.ActiveDamageOverTime{
		Source: damage.DamageSource{
			EntityID:             "projectile-1",
			EntityType:           damage.EntityTypeProjectile,
			Cause:                damage.DamageCauseDot,
			ResponsiblePlayerID:  "disconnected-source",
			OriginalInstigatorID: "disconnected-source",
		},
		Target:          damage.DamageTargetRef{EntityID: playerID, EntityType: damage.EntityTypePlayer},
		AmountPerTick:   1,
		TickSeconds:     1,
		DurationSeconds: 3,
		Type:            damage.DamageTypeThermal,
	}
	game.damageOverTime().Add(effect)
	game.stepDamageOverTime(1)
	if player.Health != 1 || player.PendingDespawn {
		t.Fatalf("after first tick health=%d pending=%v", player.Health, player.PendingDespawn)
	}
	game.stepDamageOverTime(1)
	if !player.PendingDespawn {
		t.Fatal("expected lethal DoT tick to enter player death flow")
	}
	if got := game.damageOverTime().CountTarget(playerID); got != 0 {
		t.Fatalf("remaining DoT effects = %d, want 0", got)
	}
}

func TestApplyDamageResultToPlayerSupportsHealingAndRepair(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]
	player.Health = 1
	player.Shields = 1
	applyDamageResultToPlayer(player, damage.DamageResult{
		Kind:            damage.DamageResultKindHealing,
		RemainingHealth: 4,
		RemainingShield: 3,
	})
	if player.Health != 4 || player.Shields != 3 {
		t.Fatalf("player health=%d shields=%d", player.Health, player.Shields)
	}
}
