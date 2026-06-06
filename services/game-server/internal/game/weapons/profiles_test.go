package weapons

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
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

func TestLookupTorpedoReturnsExpectedProfile(t *testing.T) {
	profile, ok := Lookup(Torpedo)
	if !ok {
		t.Fatal("expected torpedo profile to be found")
	}

	if profile.ID != Torpedo {
		t.Fatalf("ID = %q, want %q", profile.ID, Torpedo)
	}
	if profile.Slot != Secondary {
		t.Fatalf("Slot = %q, want %q", profile.Slot, Secondary)
	}
	if profile.Projectile.Type != "torpedo" {
		t.Fatalf("Projectile.Type = %q, want %q", profile.Projectile.Type, "torpedo")
	}
	if profile.ImpactEffect.Kind != ImpactEffectRadial {
		t.Fatalf("ImpactEffect.Kind = %q, want %q", profile.ImpactEffect.Kind, ImpactEffectRadial)
	}
	if profile.ImpactEffect.Radial.CoverageMode != radial.CoverageAnnularWave {
		t.Fatalf("ImpactEffect.Radial.CoverageMode = %q, want %q", profile.ImpactEffect.Radial.CoverageMode, radial.CoverageAnnularWave)
	}
	if profile.ImpactEffect.Radial.ExpirationMode != radial.ExpirationSimultaneous {
		t.Fatalf("ImpactEffect.Radial.ExpirationMode = %q, want %q", profile.ImpactEffect.Radial.ExpirationMode, radial.ExpirationSimultaneous)
	}
	if got, want := profile.ImpactEffect.Radial.ZoneCount, 4; got != want {
		t.Fatalf("ImpactEffect.Radial.ZoneCount = %d, want %d", got, want)
	}
	if !profile.ImpactEffect.Radial.TargetFilter.Allows(radial.TargetAsteroid) {
		t.Fatal("expected asteroids to be allowed")
	}
	if !profile.ImpactEffect.Radial.TargetFilter.Allows(radial.TargetEnemy) {
		t.Fatal("expected enemies to be allowed")
	}
	if profile.ImpactEffect.Radial.TargetFilter.Allows(radial.TargetPlayer) {
		t.Fatal("expected players to be excluded")
	}
	if profile.ImpactEffect.Radial.TargetFilter.Allows(radial.TargetProjectile) {
		t.Fatal("expected projectiles to be excluded")
	}
	if profile.ImpactEffect.Radial.TargetFilter.Allows(radial.TargetPickup) {
		t.Fatal("expected pickups to be excluded")
	}
}
