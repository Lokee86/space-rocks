package lives

import "testing"

func TestBaselinePolicyResolvesP4Defaults(t *testing.T) {
	policy := NewBaselinePolicy()

	if err := policy.Validate(); err != nil {
		t.Fatalf("baseline policy should validate: %v", err)
	}
	if policy.Model != LifeModelFinitePerPlayer {
		t.Fatalf("model = %q, want finite per-player", policy.Model)
	}
	if policy.RespawnTrigger != RespawnTriggerManual {
		t.Fatalf("respawn trigger = %q, want manual", policy.RespawnTrigger)
	}
	if policy.Restoration != NewBaselineRestorationPolicy() {
		t.Fatalf("restoration = %+v, want baseline restoration", policy.Restoration)
	}
	if policy.SpawnProfileID != DefaultSpawnProfileID {
		t.Fatalf("spawn profile = %q, want %q", policy.SpawnProfileID, DefaultSpawnProfileID)
	}
}

func TestPolicyValidationRequiresModelSpecificFields(t *testing.T) {
	finiteWithoutLives := NewBaselinePolicy()
	finiteWithoutLives.StartingLives = 0

	sharedWithoutPool := NewBaselinePolicy()
	sharedWithoutPool.Model = LifeModelSharedTeamPool
	sharedWithoutPool.StartingLives = 0

	shared := NewBaselinePolicy()
	shared.Model = LifeModelSharedTeamPool
	shared.StartingLives = 0
	shared.TeamPool = &TeamPoolPolicy{
		PoolID:        "team-1",
		StartingLives: 4,
	}

	infinite := NewBaselinePolicy()
	infinite.Model = LifeModelInfinite
	infinite.StartingLives = 0

	tests := []struct {
		name   string
		policy Policy
		valid  bool
	}{
		{name: "finite requires lives", policy: finiteWithoutLives},
		{name: "shared requires pool", policy: sharedWithoutPool},
		{name: "valid shared pool", policy: shared, valid: true},
		{name: "valid infinite", policy: infinite, valid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.policy.Validate()
			if (err == nil) != test.valid {
				t.Fatalf("Validate() error = %v, valid = %t", err, test.valid)
			}
		})
	}
}

func TestPolicyValidationRejectsInvalidValuesAndNegativeDurations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy)
	}{
		{name: "life model", mutate: func(policy *Policy) { policy.Model = LifeModel("invalid") }},
		{name: "respawn trigger", mutate: func(policy *Policy) { policy.RespawnTrigger = RespawnTrigger("invalid") }},
		{name: "unimplemented respawn trigger", mutate: func(policy *Policy) { policy.RespawnTrigger = RespawnTriggerAutomatic }},
		{name: "negative starting lives", mutate: func(policy *Policy) { policy.StartingLives = -1 }},
		{name: "negative respawn delay", mutate: func(policy *Policy) { policy.RespawnDelay = -1 }},
		{name: "empty spawn profile", mutate: func(policy *Policy) { policy.SpawnProfileID = "" }},
		{name: "unsupported spawn profile", mutate: func(policy *Policy) { policy.SpawnProfileID = "future_profile" }},
		{name: "health restoration", mutate: func(policy *Policy) { policy.Restoration.Health = RestorationMode("invalid") }},
		{name: "negative cooldown threshold", mutate: func(policy *Policy) { policy.Restoration.ShortCooldownThreshold = -1 }},
		{name: "temporary effects", mutate: func(policy *Policy) { policy.Restoration.TemporaryEffects = TemporaryEffectsPolicy("invalid") }},
		{name: "loadout persistence", mutate: func(policy *Policy) { policy.Restoration.Loadout = LoadoutPersistence("invalid") }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := NewBaselinePolicy()
			test.mutate(&policy)
			if err := policy.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
		})
	}
}

func TestTeamPoolValidation(t *testing.T) {
	invalidPool := TeamPoolPolicy{
		PoolID:        "team-1",
		StartingLives: -1,
	}
	if err := invalidPool.Validate(); err == nil {
		t.Fatal("negative team-pool lives unexpectedly validated")
	}
}
