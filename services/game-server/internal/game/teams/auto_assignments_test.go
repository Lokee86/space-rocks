package teams

import (
	"reflect"
	"testing"
)

func TestResolveAutoBalancedIsDeterministicAndSizeOnly(t *testing.T) {
	first, err := ResolveAutoBalanced([]string{"player-4", "player-1", "player-3", "player-2", "player-5"}, 2)
	if err != nil {
		t.Fatalf("ResolveAutoBalanced() error = %v", err)
	}
	second, err := ResolveAutoBalanced([]string{"player-5", "player-2", "player-4", "player-3", "player-1"}, 2)
	if err != nil {
		t.Fatalf("ResolveAutoBalanced() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("assignment changed with input order: %v vs %v", first, second)
	}
	want := Assignments{"player-1": Team1, "player-2": Team2, "player-3": Team1, "player-4": Team2, "player-5": Team1}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("ResolveAutoBalanced() = %v, want %v", first, want)
	}
}

func TestResolveAutoBalancedKeepsTeamSizesWithinOne(t *testing.T) {
	assignments, err := ResolveAutoBalanced(participantIDs(17), 5)
	if err != nil {
		t.Fatalf("ResolveAutoBalanced() error = %v", err)
	}
	sizes := make(map[ID]int)
	for _, teamID := range assignments {
		sizes[teamID]++
	}
	minimum, maximum := len(assignments), 0
	for _, teamID := range OrderedIDs()[:5] {
		if sizes[teamID] < minimum {
			minimum = sizes[teamID]
		}
		if sizes[teamID] > maximum {
			maximum = sizes[teamID]
		}
	}
	if maximum-minimum > 1 {
		t.Fatalf("team sizes differ by more than one: %v", sizes)
	}
}

func TestResolveAutoBalancedRejectsInvalidCountsAndRoster(t *testing.T) {
	for _, count := range []int{0, 1, 9} {
		if _, err := ResolveAutoBalanced(nil, count); err == nil {
			t.Errorf("ResolveAutoBalanced(teamCount=%d) returned nil error", count)
		}
	}
	if _, err := ResolveAutoBalanced([]string{"player-1", ""}, 2); err == nil {
		t.Fatal("ResolveAutoBalanced accepted a blank participant ID")
	}
}

func TestResolveAutoBalancedEmptyRosterReturnsFreshMap(t *testing.T) {
	first, err := ResolveAutoBalanced(nil, 2)
	if err != nil {
		t.Fatalf("ResolveAutoBalanced() error = %v", err)
	}
	second, err := ResolveAutoBalanced(nil, 2)
	if err != nil {
		t.Fatalf("ResolveAutoBalanced() error = %v", err)
	}
	if first == nil || second == nil {
		t.Fatal("empty roster returned nil assignments")
	}
	first["player-1"] = Team1
	if len(second) != 0 {
		t.Fatal("auto-balanced resolution reused assignment state")
	}
}
