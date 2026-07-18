package awards

import "testing"

func TestFixedAndCustomCounters(t *testing.T) {
	runtime := NewRuntime()
	if !runtime.CounterRegistered(CounterScore) {
		t.Fatal("SCORE should be registered")
	}
	if err := runtime.RegisterCustomCounter("BONUS_POINTS"); err != nil {
		t.Fatalf("register custom counter: %v", err)
	}
	if !runtime.CounterRegistered("BONUS_POINTS") {
		t.Fatal("custom counter should be registered")
	}
	if err := runtime.RegisterCustomCounter("bad-counter"); err == nil {
		t.Fatal("expected invalid custom counter rejection")
	}
}

func TestApplyEventMutationsAndDuplicateSuppression(t *testing.T) {
	runtime := NewRuntime()
	owner := Owner{Scope: ScopePlayer, ID: "player-1"}
	mutations := []Mutation{
		{Owner: owner, CounterID: CounterScore, Operation: MutationIncrement, Value: 10},
		{Owner: owner, CounterID: CounterKills, Operation: MutationSet, Value: 2},
		{Owner: owner, CounterID: CounterKills, Operation: MutationMinimum, Value: 4},
		{Owner: owner, CounterID: CounterScore, Operation: MutationMaximum, Value: 8},
		{Owner: owner, CounterID: CounterCompletionTime, Operation: MutationTimedAccumulate, Value: 1.5},
	}
	result, err := runtime.ApplyEvent("event-1", mutations)
	if err != nil {
		t.Fatalf("apply event: %v", err)
	}
	if !result.Applied || result.Duplicate {
		t.Fatalf("unexpected result: %+v", result)
	}
	if score, _ := runtime.Counter(owner, CounterScore); score != 8 {
		t.Fatalf("score = %v, want 8", score)
	}
	if kills, _ := runtime.Counter(owner, CounterKills); kills != 4 {
		t.Fatalf("kills = %v, want 4", kills)
	}
	duplicate, err := runtime.ApplyEvent("event-1", mutations)
	if err != nil || !duplicate.Duplicate || duplicate.Applied {
		t.Fatalf("duplicate result = %+v, err = %v", duplicate, err)
	}
	if score, _ := runtime.Counter(owner, CounterScore); score != 8 {
		t.Fatalf("duplicate changed score to %v", score)
	}
}

func TestAllOwnershipScopesAndVisibilityUseOneMutationPath(t *testing.T) {
	runtime := NewRuntime()
	if err := runtime.RegisterCustomCounterWithVisibility("CUSTOM_VALUE", VisibilityPlayerPrivate); err != nil {
		t.Fatalf("register custom counter: %v", err)
	}
	mutations := []Mutation{
		{Owner: Owner{Scope: ScopePlayer, ID: "player-1"}, CounterID: "CUSTOM_VALUE", Operation: MutationSet, Value: 1},
		{Owner: Owner{Scope: ScopeTeam, ID: "team-1"}, CounterID: CounterScore, Operation: MutationIncrement, Value: 2},
		{Owner: Owner{Scope: ScopeMatch, ID: "match-1"}, CounterID: CounterCompletionTime, Operation: MutationTimedAccumulate, Value: 3.5},
		{Owner: Owner{Scope: ScopeObjective, ID: "objective-1"}, CounterID: CounterObjectiveProgress, Operation: MutationIncrement, Value: 4},
	}
	if _, err := runtime.ApplyEvent("scopes", mutations); err != nil {
		t.Fatalf("apply scoped event: %v", err)
	}
	snapshot := runtime.Snapshot()
	if len(snapshot) != 4 {
		t.Fatalf("snapshot length = %d, want 4", len(snapshot))
	}
	for _, counter := range snapshot {
		if counter.CounterID == "CUSTOM_VALUE" && counter.Visibility != VisibilityPlayerPrivate {
			t.Fatalf("custom visibility = %q", counter.Visibility)
		}
	}
}

func TestSnapshotAndDerivedTeamTotalsAreDeterministic(t *testing.T) {
	runtime := NewRuntime()
	_, _ = runtime.ApplyEvent("p2", []Mutation{{Owner: Owner{Scope: ScopePlayer, ID: "player-2"}, CounterID: CounterScore, Operation: MutationIncrement, Value: 20}})
	_, _ = runtime.ApplyEvent("p1", []Mutation{{Owner: Owner{Scope: ScopePlayer, ID: "player-1"}, CounterID: CounterScore, Operation: MutationIncrement, Value: 10}})

	snapshot := runtime.Snapshot()
	if len(snapshot) != 2 || snapshot[0].Owner.ID != "player-1" || snapshot[1].Owner.ID != "player-2" {
		t.Fatalf("unexpected snapshot order: %+v", snapshot)
	}
	teams := runtime.DerivedTeamTotals(map[string]string{"player-2": "team-1", "player-1": "team-1"}, []CounterID{CounterScore})
	if len(teams) != 1 || teams[0].Value != 30 {
		t.Fatalf("unexpected team totals: %+v", teams)
	}
}
