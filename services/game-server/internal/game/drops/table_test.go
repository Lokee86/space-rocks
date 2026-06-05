package drops

import "testing"

func TestEvaluateReturnsNoResultForUnknownTableID(t *testing.T) {
	tables := Tables{ByID: map[string]Table{}}

	results, ok := tables.Evaluate(
		"missing",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0}},
	)

	if ok {
		t.Fatalf("expected no result, got %#v", results)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %#v", results)
	}
}

func TestEvaluateReturnsNoResultForSourceTypeMismatch(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeSingle,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "1_up", Chance: 0.05, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceType("ship"), ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0}},
	)

	if ok {
		t.Fatalf("expected no result, got %#v", results)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %#v", results)
	}
}

func TestEvaluateReturnsNoResultForSourceSizeOutOfBounds(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeSingle,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "1_up", Chance: 0.05, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 5, X: 10, Y: 20},
		Roll{Values: []float64{0}},
	)

	if ok {
		t.Fatalf("expected no result, got %#v", results)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %#v", results)
	}
}

func TestEvaluateReturnsNoResultWhenRollMeetsOrExceedsChance(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeSingle,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "1_up", Chance: 0.05, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0.05}},
	)

	if ok {
		t.Fatalf("expected no result, got %#v", results)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %#v", results)
	}
}

func TestEvaluateReturnsResultWhenRollIsBelowChance(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeSingle,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "1_up", Chance: 0.05, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0.049}},
	)

	if !ok {
		t.Fatalf("expected result, got none")
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	result := results[0]
	if result.TableID != "basicasteroids" {
		t.Fatalf("unexpected table id: %#v", result)
	}
	if result.PickupType != "1_up" {
		t.Fatalf("unexpected pickup type: %#v", result)
	}
	if result.X != 10 || result.Y != 20 {
		t.Fatalf("unexpected result position: %#v", result)
	}
}

func TestEvaluateSingleModeReturnsFirstMatchingEntry(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeSingle,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "first", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
					{PickupType: "second", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0}},
	)

	if !ok {
		t.Fatal("expected result, got none")
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	if results[0].PickupType != "first" {
		t.Fatalf("expected first matching entry, got %#v", results[0])
	}
}

func TestEvaluateMultiModeCanReturnTwoResults(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeMulti,
				MaxDropsPerSource: 2,
				Entries: []Entry{
					{PickupType: "first", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
					{PickupType: "second", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0, 0}},
	)

	if !ok {
		t.Fatal("expected result, got none")
	}
	if len(results) != 2 {
		t.Fatalf("expected two results, got %#v", results)
	}
	if results[0].PickupType != "first" || results[1].PickupType != "second" {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestEvaluateMultiModeRespectsMaxDropsPerSource(t *testing.T) {
	tables := Tables{
		ByID: map[string]Table{
			"basicasteroids": {
				ID:               "basicasteroids",
				SourceType:       SourceTypeAsteroid,
				DropMode:         DropModeMulti,
				MaxDropsPerSource: 1,
				Entries: []Entry{
					{PickupType: "first", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
					{PickupType: "second", Chance: 1.0, MinSourceSize: 1, MaxSourceSize: 4},
				},
			},
		},
	}

	results, ok := tables.Evaluate(
		"basicasteroids",
		Source{Type: SourceTypeAsteroid, ID: "asteroid-1", Size: 2, X: 10, Y: 20},
		Roll{Values: []float64{0, 0}},
	)

	if !ok {
		t.Fatal("expected result, got none")
	}
	if len(results) != 1 {
		t.Fatalf("expected one result due to max cap, got %#v", results)
	}
	if results[0].PickupType != "first" {
		t.Fatalf("unexpected result: %#v", results[0])
	}
}
