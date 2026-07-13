package spatial

import "testing"

func TestKindMaskAllows(t *testing.T) {
	mask := KindMask(KindPlayer | KindAsteroid)

	if !mask.Allows(KindPlayer) {
		t.Fatal("player should be allowed")
	}
	if mask.Allows(KindProjectile) {
		t.Fatal("projectile should not be allowed")
	}
	if KindMask(0).Allows(KindPlayer) {
		t.Fatal("zero mask should allow nothing")
	}
	if !AllKinds.Allows(KindPickup) {
		t.Fatal("all-kinds mask should allow pickup")
	}
}


