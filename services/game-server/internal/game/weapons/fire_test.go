package weapons

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestFireBasicCannonInfiniteAmmoFiresAtZeroAmmo(t *testing.T) {
	result := Fire(FireRequest{
		Equipped: Equipped{
			ID:         BasicCannon,
			AmmoPolicy: InfiniteAmmo,
		},
		State: SlotState{
			AmmoRemaining: 0,
		},
		Position: physics.Vector2{X: 10, Y: 20},
		Forward:  physics.Vector2{X: 1, Y: 0},
		Rotation: 0.75,
	})

	if !result.Fired {
		t.Fatal("expected basic cannon to fire with infinite ammo")
	}
	if result.NewState.AmmoRemaining != 0 {
		t.Fatalf("AmmoRemaining = %d, want 0", result.NewState.AmmoRemaining)
	}
	if result.Projectile.WeaponID != BasicCannon {
		t.Fatalf("WeaponID = %q, want %q", result.Projectile.WeaponID, BasicCannon)
	}
}

func TestFireLimitedAmmoBlocksAtZeroAmmo(t *testing.T) {
	result := Fire(FireRequest{
		Equipped: Equipped{
			ID:         BasicCannon,
			AmmoPolicy: LimitedAmmo,
		},
		State: SlotState{
			AmmoRemaining: 0,
		},
	})

	if result.Fired {
		t.Fatal("expected limited ammo weapon to be blocked at zero ammo")
	}
}

func TestFireLimitedAmmoDecrementsWhenFiring(t *testing.T) {
	result := Fire(FireRequest{
		Equipped: Equipped{
			ID:         BasicCannon,
			AmmoPolicy: LimitedAmmo,
		},
		State: SlotState{
			AmmoRemaining: 3,
		},
		Forward: physics.Vector2{X: 0, Y: 1},
	})

	if !result.Fired {
		t.Fatal("expected limited ammo weapon to fire with ammo available")
	}
	if result.NewState.AmmoRemaining != 2 {
		t.Fatalf("AmmoRemaining = %d, want 2", result.NewState.AmmoRemaining)
	}
}

func TestFireCooldownBlocksFiring(t *testing.T) {
	result := Fire(FireRequest{
		Equipped: Equipped{
			ID:         BasicCannon,
			AmmoPolicy: InfiniteAmmo,
		},
		State: SlotState{
			CooldownRemaining: 0.1,
		},
	})

	if result.Fired {
		t.Fatal("expected cooldown to block firing")
	}
}
