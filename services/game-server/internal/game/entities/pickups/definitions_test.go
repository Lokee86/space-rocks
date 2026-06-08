package pickups

import "testing"

func TestDefinitionForOneUp(t *testing.T) {
	definition, ok := DefinitionFor(TypeOneUp)
	if !ok {
		t.Fatalf("expected definition for %q", TypeOneUp)
	}

	if definition.Type != TypeOneUp {
		t.Fatalf("expected type %q, got %q", TypeOneUp, definition.Type)
	}

	if definition.Health <= 0 {
		t.Fatalf("expected positive health, got %d", definition.Health)
	}

	if definition.LifespanSeconds <= 0 {
		t.Fatalf("expected positive lifespan, got %f", definition.LifespanSeconds)
	}

	if definition.Class != ClassPowerup {
		t.Fatalf("expected class %q, got %q", ClassPowerup, definition.Class)
	}
}

func TestDefinitionForTorpedo(t *testing.T) {
	definition, ok := DefinitionFor(TypeTorpedo)
	if !ok {
		t.Fatalf("expected definition for %q", TypeTorpedo)
	}

	if definition.Type != TypeTorpedo {
		t.Fatalf("expected type %q, got %q", TypeTorpedo, definition.Type)
	}

	if definition.Class != ClassWeapon {
		t.Fatalf("expected class %q, got %q", ClassWeapon, definition.Class)
	}

	if definition.Health <= 0 {
		t.Fatalf("expected positive health, got %d", definition.Health)
	}

	if definition.LifespanSeconds <= 0 {
		t.Fatalf("expected positive lifespan, got %f", definition.LifespanSeconds)
	}
}

func TestDefinitionForUnknownType(t *testing.T) {
	definition, ok := DefinitionFor(PickupType("unknown"))
	if ok {
		t.Fatalf("expected no definition for unknown pickup type, got %+v", definition)
	}
}
