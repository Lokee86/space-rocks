package objectives

import "testing"

func TestCompositionHierarchy(t *testing.T) {
	valid := []Composition{
		{ID: "single", Kind: CompositionSingle},
		{ID: "multiple", Kind: CompositionMultiple, Members: []ComponentRef{{ID: "single", Kind: CompositionSingle}}},
		{ID: "meta", Kind: CompositionMeta, Members: []ComponentRef{{ID: "multiple", Kind: CompositionMultiple}}},
	}
	for _, composition := range valid {
		if err := composition.Validate(); err != nil {
			t.Fatalf("%s: %v", composition.ID, err)
		}
	}
	invalid := Composition{ID: "bad", Kind: CompositionMeta, Members: []ComponentRef{{ID: "single", Kind: CompositionSingle}}}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid hierarchy to fail")
	}
}
