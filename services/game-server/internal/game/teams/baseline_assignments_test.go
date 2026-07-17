package teams

import (
	"reflect"
	"testing"
)

func TestResolveFFA(t *testing.T) {
	input := []string{"player-2", "player-1"}
	got, err := ResolveFFA(input)
	if err != nil {
		t.Fatalf("ResolveFFA() error = %v", err)
	}
	want := Assignments{"player-1": NoTeam, "player-2": NoTeam}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveFFA() = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(input, []string{"player-2", "player-1"}) {
		t.Fatal("ResolveFFA mutated its input")
	}
}

func TestResolveCoop(t *testing.T) {
	got, err := ResolveCoop([]string{"player-2", "player-1"})
	if err != nil {
		t.Fatalf("ResolveCoop() error = %v", err)
	}
	want := Assignments{"player-1": Team1, "player-2": Team1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveCoop() = %v, want %v", got, want)
	}
}

func TestResolveBaselineEmptyRosterReturnsFreshMap(t *testing.T) {
	first, err := ResolveFFA(nil)
	if err != nil {
		t.Fatalf("ResolveFFA() error = %v", err)
	}
	second, err := ResolveFFA(nil)
	if err != nil {
		t.Fatalf("ResolveFFA() error = %v", err)
	}
	if first == nil || second == nil {
		t.Fatal("empty roster returned nil assignments")
	}
	first["player-1"] = Team1
	if len(second) != 0 {
		t.Fatal("baseline resolution reused assignment state")
	}
}

func TestResolveBaselineRejectsInvalidRoster(t *testing.T) {
	if _, err := ResolveFFA([]string{"player-1", ""}); err == nil {
		t.Fatal("ResolveFFA accepted a blank participant ID")
	}
	if _, err := ResolveCoop([]string{"player-1", "player-1"}); err == nil {
		t.Fatal("ResolveCoop accepted a duplicate participant ID")
	}
}
