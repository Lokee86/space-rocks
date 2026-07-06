package realtime

import "testing"

func TestDefaultHotLaneOffloadPolicy(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()

	if policy.AsteroidHotLaneEntityBudget != DefaultAsteroidHotLaneEntityBudget {
		t.Fatalf("AsteroidHotLaneEntityBudget = %d, want %d", policy.AsteroidHotLaneEntityBudget, DefaultAsteroidHotLaneEntityBudget)
	}
	if policy.BulletHotLaneEntityBudget != DefaultBulletHotLaneEntityBudget {
		t.Fatalf("BulletHotLaneEntityBudget = %d, want %d", policy.BulletHotLaneEntityBudget, DefaultBulletHotLaneEntityBudget)
	}
	if policy.TargetEncodedPacketBytes != 800 {
		t.Fatalf("TargetEncodedPacketBytes = %d, want 800", policy.TargetEncodedPacketBytes)
	}
	if policy.HardEncodedPacketBytes != 1200 {
		t.Fatalf("HardEncodedPacketBytes = %d, want 1200", policy.HardEncodedPacketBytes)
	}
	if policy.MTUSafePacketBytes != 1500 {
		t.Fatalf("MTUSafePacketBytes = %d, want 1500", policy.MTUSafePacketBytes)
	}
}

func TestDefaultHotLaneOffloadPolicyDerivedThresholds(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	wantWorldBudget := DefaultAsteroidHotLaneEntityBudget
	if DefaultBulletHotLaneEntityBudget < wantWorldBudget {
		wantWorldBudget = DefaultBulletHotLaneEntityBudget
	}

	if got := policy.WorldHotEntityBudget(); got != wantWorldBudget {
		t.Fatalf("WorldHotEntityBudget() = %d, want %d", got, wantWorldBudget)
	}
	if got := policy.AsteroidFullOffloadThreshold(); got != DefaultAsteroidHotLaneEntityBudget*2 {
		t.Fatalf("AsteroidFullOffloadThreshold() = %d, want %d", got, DefaultAsteroidHotLaneEntityBudget*2)
	}
	if got := policy.BulletFullOffloadThreshold(); got != DefaultBulletHotLaneEntityBudget*2 {
		t.Fatalf("BulletFullOffloadThreshold() = %d, want %d", got, DefaultBulletHotLaneEntityBudget*2)
	}
	if got := policy.AsteroidNeedsChunkingThreshold(); got != DefaultAsteroidHotLaneEntityBudget*3 {
		t.Fatalf("AsteroidNeedsChunkingThreshold() = %d, want %d", got, DefaultAsteroidHotLaneEntityBudget*3)
	}
	if got := policy.BulletNeedsChunkingThreshold(); got != DefaultBulletHotLaneEntityBudget*3 {
		t.Fatalf("BulletNeedsChunkingThreshold() = %d, want %d", got, DefaultBulletHotLaneEntityBudget*3)
	}
}

func TestHotLaneOffloadPolicyWorldBudgetUsesLowerValue(t *testing.T) {
	policy := HotLaneOffloadPolicy{
		AsteroidHotLaneEntityBudget: 80,
		BulletHotLaneEntityBudget:   48,
	}

	if got := policy.WorldHotEntityBudget(); got != 48 {
		t.Fatalf("WorldHotEntityBudget() = %d, want 48", got)
	}
}
