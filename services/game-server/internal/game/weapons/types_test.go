package weapons

import "testing"

func TestDefaultPlayerArmory(t *testing.T) {
	armory := DefaultPlayerArmory()

	if armory.Primary.ID != BasicCannon {
		t.Fatalf("Primary.ID = %q, want %q", armory.Primary.ID, BasicCannon)
	}

	if armory.Primary.AmmoPolicy != InfiniteAmmo {
		t.Fatalf("Primary.AmmoPolicy = %q, want %q", armory.Primary.AmmoPolicy, InfiniteAmmo)
	}

	if armory.Secondary != EmptyEquipped() {
		t.Fatalf("Secondary = %#v, want %#v", armory.Secondary, EmptyEquipped())
	}
}
