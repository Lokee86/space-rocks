package teams

import (
	"reflect"
	"testing"
)

func TestCanonicalRoster(t *testing.T) {
	input := []string{"player-2", "player-1"}
	got, err := CanonicalRoster(input)
	if err != nil {
		t.Fatalf("CanonicalRoster() error = %v", err)
	}
	if want := []string{"player-1", "player-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("CanonicalRoster() = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(input, []string{"player-2", "player-1"}) {
		t.Fatal("CanonicalRoster mutated its input")
	}
}

func TestCanonicalRosterRejectsInvalidInput(t *testing.T) {
	for _, input := range [][]string{{"player-1", ""}, {"player-1", "player-1"}} {
		if _, err := CanonicalRoster(input); err == nil {
			t.Fatalf("CanonicalRoster(%v) returned nil error", input)
		}
	}
}
