package devtools

import "testing"

func TestPlayerShapeIDReturnsWingShapeForExplicitWingType(t *testing.T) {
	got := PlayerShapeID("v_wing")
	want := "player:v_wing"
	if got != want {
		t.Fatalf("PlayerShapeID() = %q, want %q", got, want)
	}
}

func TestPlayerShapeIDDefaultsEmptyShipTypeToWing(t *testing.T) {
	got := PlayerShapeID("")
	want := "player:v_wing"
	if got != want {
		t.Fatalf("PlayerShapeID() = %q, want %q", got, want)
	}
}

func TestAsteroidShapeIDFormatsVariant(t *testing.T) {
	got := AsteroidShapeID(2)
	want := "asteroid:2"
	if got != want {
		t.Fatalf("AsteroidShapeID() = %q, want %q", got, want)
	}
}

func TestBulletShapeIDReturnsBullet(t *testing.T) {
	got := BulletShapeID()
	want := "bullet"
	if got != want {
		t.Fatalf("BulletShapeID() = %q, want %q", got, want)
	}
}

func TestPickupShapeIDFormatsPickupType(t *testing.T) {
	got := PickupShapeID("1_up")
	want := "pickup:1_up"
	if got != want {
		t.Fatalf("PickupShapeID() = %q, want %q", got, want)
	}
}
