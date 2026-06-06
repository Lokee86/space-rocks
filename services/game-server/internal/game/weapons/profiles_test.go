package weapons

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
)

func TestLookupBasicCannonReturnsExpectedProfile(t *testing.T) {
	profile, ok := Lookup(BasicCannon)
	if !ok {
		t.Fatal("expected basic cannon profile to be found")
	}

	if profile.ID != BasicCannon {
		t.Fatalf("ID = %q, want %q", profile.ID, BasicCannon)
	}
	if profile.Slot != Primary {
		t.Fatalf("Slot = %q, want %q", profile.Slot, Primary)
	}
	if profile.CooldownSeconds != constants.BulletCooldown {
		t.Fatalf("CooldownSeconds = %v, want %v", profile.CooldownSeconds, constants.BulletCooldown)
	}
	if profile.Projectile.Type != "bullet" {
		t.Fatalf("Projectile.Type = %q, want %q", profile.Projectile.Type, "bullet")
	}
	if profile.Projectile.Speed != constants.BulletSpeed {
		t.Fatalf("Projectile.Speed = %v, want %v", profile.Projectile.Speed, constants.BulletSpeed)
	}
	if profile.Projectile.Lifetime != constants.BulletLifetime {
		t.Fatalf("Projectile.Lifetime = %v, want %v", profile.Projectile.Lifetime, constants.BulletLifetime)
	}
	if profile.Projectile.SpawnOffset != constants.BulletSpawnOffset {
		t.Fatalf("Projectile.SpawnOffset = %v, want %v", profile.Projectile.SpawnOffset, constants.BulletSpawnOffset)
	}
	if profile.Damage.Amount != constants.BulletDamage {
		t.Fatalf("Damage.Amount = %d, want %d", profile.Damage.Amount, constants.BulletDamage)
	}
	if profile.Damage.Type != damage.DamageTypeKinetic {
		t.Fatalf("Damage.Type = %q, want %q", profile.Damage.Type, damage.DamageTypeKinetic)
	}
	if profile.Damage.Cause != damage.DamageCauseProjectile {
		t.Fatalf("Damage.Cause = %q, want %q", profile.Damage.Cause, damage.DamageCauseProjectile)
	}
	if profile.Damage.BypassShield {
		t.Fatal("expected ammo policy and bypass shield to stay outside the immutable profile")
	}
}

