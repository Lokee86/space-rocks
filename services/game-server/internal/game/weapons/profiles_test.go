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
	if profile.CooldownSeconds != constants.BasicCannonCooldown {
		t.Fatalf("CooldownSeconds = %v, want %v", profile.CooldownSeconds, constants.BasicCannonCooldown)
	}
	if profile.Projectile.Type != "bullet" {
		t.Fatalf("Projectile.Type = %q, want %q", profile.Projectile.Type, "bullet")
	}
	if profile.Projectile.Speed != constants.BasicCannonProjectileSpeed {
		t.Fatalf("Projectile.Speed = %v, want %v", profile.Projectile.Speed, constants.BasicCannonProjectileSpeed)
	}
	if profile.Projectile.Lifetime != constants.BasicCannonProjectileLifetime {
		t.Fatalf("Projectile.Lifetime = %v, want %v", profile.Projectile.Lifetime, constants.BasicCannonProjectileLifetime)
	}
	if profile.Projectile.SpawnOffset != constants.BasicCannonProjectileSpawnOffset {
		t.Fatalf("Projectile.SpawnOffset = %v, want %v", profile.Projectile.SpawnOffset, constants.BasicCannonProjectileSpawnOffset)
	}
	if profile.Damage.Amount != constants.BasicCannonDamage {
		t.Fatalf("Damage.Amount = %d, want %d", profile.Damage.Amount, constants.BasicCannonDamage)
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
	if profile.CooldownSeconds != float64(constants.TorpedoCooldown) {
		t.Fatalf("CooldownSeconds = %v, want %v", profile.CooldownSeconds, constants.TorpedoCooldown)
	}
	if profile.Projectile.Speed != float64(constants.TorpedoProjectileSpeed) {
		t.Fatalf("Projectile.Speed = %v, want %v", profile.Projectile.Speed, constants.TorpedoProjectileSpeed)
	}
	if profile.Projectile.Lifetime != float64(constants.TorpedoProjectileLifetime) {
		t.Fatalf("Projectile.Lifetime = %v, want %v", profile.Projectile.Lifetime, constants.TorpedoProjectileLifetime)
	}
	if profile.Projectile.SpawnOffset != float64(constants.TorpedoProjectileSpawnOffset) {
		t.Fatalf("Projectile.SpawnOffset = %v, want %v", profile.Projectile.SpawnOffset, constants.TorpedoProjectileSpawnOffset)
	}
	if profile.Damage.Amount != int(constants.TorpedoImpactDamage) {
		t.Fatalf("Damage.Amount = %d, want %d", profile.Damage.Amount, constants.TorpedoImpactDamage)
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
	if got, want := profile.ImpactEffect.Radial.ZoneCount, int(constants.TorpedoRadialZoneCount); got != want {
		t.Fatalf("ImpactEffect.Radial.ZoneCount = %d, want %d", got, want)
	}
	if got, want := profile.ImpactEffect.Radial.ZoneWidth, float64(constants.TorpedoRadialZoneWidth); got != want {
		t.Fatalf("ImpactEffect.Radial.ZoneWidth = %v, want %v", got, want)
	}
	if got, want := profile.ImpactEffect.Radial.ZoneSpawnSeconds, float64(constants.TorpedoRadialZoneSpawnSeconds); got != want {
		t.Fatalf("ImpactEffect.Radial.ZoneSpawnSeconds = %v, want %v", got, want)
	}
	if got, want := profile.ImpactEffect.Radial.TickSeconds, float64(constants.TorpedoRadialTickSeconds); got != want {
		t.Fatalf("ImpactEffect.Radial.TickSeconds = %v, want %v", got, want)
	}
	if got, want := profile.ImpactEffect.Radial.TotalSeconds, float64(constants.TorpedoRadialTotalSeconds); got != want {
		t.Fatalf("ImpactEffect.Radial.TotalSeconds = %v, want %v", got, want)
	}
	if got, want := profile.ImpactEffect.Radial.ZoneLifetimeSeconds, float64(constants.TorpedoRadialZoneLifetimeSeconds); got != want {
		t.Fatalf("ImpactEffect.Radial.ZoneLifetimeSeconds = %v, want %v", got, want)
	}
	if profile.ImpactEffect.Radial.Damage.Amount != int(constants.TorpedoRadialDamage) {
		t.Fatalf("ImpactEffect.Radial.Damage.Amount = %d, want %d", profile.ImpactEffect.Radial.Damage.Amount, constants.TorpedoRadialDamage)
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
