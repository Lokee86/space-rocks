package pickups

import "testing"

func TestResolveCollection_OneUp(t *testing.T) {
	req := CollectionRequest{
		PlayerID:   "Player-1",
		PickupID:   "pickup-1",
		PickupType: "1_up",
		X:          12.5,
		Y:          34.75,
	}

	result := ResolveCollection(req)

	if !result.Collected {
		t.Fatalf("expected collected result")
	}

	if result.PlayerID != req.PlayerID {
		t.Fatalf("expected player id %q, got %q", req.PlayerID, result.PlayerID)
	}

	if result.PickupID != req.PickupID {
		t.Fatalf("expected pickup id %q, got %q", req.PickupID, result.PickupID)
	}

	if result.PickupType != req.PickupType {
		t.Fatalf("expected pickup type %q, got %q", req.PickupType, result.PickupType)
	}

	if result.X != req.X {
		t.Fatalf("expected x %v, got %v", req.X, result.X)
	}

	if result.Y != req.Y {
		t.Fatalf("expected y %v, got %v", req.Y, result.Y)
	}

	if result.EffectIntent.EffectType != EffectTypeAddLives {
		t.Fatalf("expected effect type %q, got %q", EffectTypeAddLives, result.EffectIntent.EffectType)
	}

	if result.EffectIntent.Amount != 1 {
		t.Fatalf("expected amount 1, got %d", result.EffectIntent.Amount)
	}
}

func TestResolveCollection_UnknownType(t *testing.T) {
	result := ResolveCollection(CollectionRequest{
		PlayerID:   "Player-1",
		PickupID:   "pickup-1",
		PickupType: "unknown",
		X:          12.5,
		Y:          34.75,
	})

	if result.Collected {
		t.Fatalf("expected uncollected result")
	}
}
