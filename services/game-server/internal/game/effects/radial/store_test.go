package radial

import "testing"

func TestStoreEmpty(t *testing.T) {
	store := NewStore()

	if got, want := store.Len(), 0; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if got, want := len(store.All()), 0; got != want {
		t.Fatalf("len(All()) = %d, want %d", got, want)
	}
}

func TestStoreAddEffect(t *testing.T) {
	store := NewStore()
	effect := Effect{ID: "effect-1"}

	store.Add(effect)

	if got, want := store.Len(), 1; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if _, ok := store.All()["effect-1"]; !ok {
		t.Fatal("expected effect-1 to be present")
	}
}

func TestStoreRemoveEffect(t *testing.T) {
	store := NewStore()
	store.Add(Effect{ID: "effect-1"})

	store.Remove("effect-1")

	if got, want := store.Len(), 0; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	if _, ok := store.All()["effect-1"]; ok {
		t.Fatal("expected effect-1 to be removed")
	}
}
