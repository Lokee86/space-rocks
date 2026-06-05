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

	const expectedScenePath = "res://scenes/pickups/1_up.tscn"
	if definition.ScenePath != expectedScenePath {
		t.Fatalf("expected scene path %q, got %q", expectedScenePath, definition.ScenePath)
	}
}

func TestDefinitionForUnknownType(t *testing.T) {
	definition, ok := DefinitionFor(PickupType("unknown"))
	if ok {
		t.Fatalf("expected no definition for unknown pickup type, got %+v", definition)
	}
}
