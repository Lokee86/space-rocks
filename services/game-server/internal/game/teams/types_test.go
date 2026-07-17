package teams

import "testing"

func TestTeamIDValues(t *testing.T) {
	tests := []struct {
		name string
		got  TeamID
		want TeamID
	}{
		{name: "no team", got: NoTeam, want: ""},
		{name: "team 1", got: Team1, want: "team_1"},
		{name: "team 2", got: Team2, want: "team_2"},
		{name: "team 3", got: Team3, want: "team_3"},
		{name: "team 4", got: Team4, want: "team_4"},
		{name: "team 5", got: Team5, want: "team_5"},
		{name: "team 6", got: Team6, want: "team_6"},
		{name: "team 7", got: Team7, want: "team_7"},
		{name: "team 8", got: Team8, want: "team_8"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("got %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestIDCompatibilityAlias(t *testing.T) {
	var canonical TeamID = Team1
	var compatibility ID = canonical
	if compatibility != canonical {
		t.Fatalf("ID alias value = %q, want %q", compatibility, canonical)
	}

	compatibilityIDs := []ID{Team1ID, Team2ID, Team3ID, Team4ID, Team5ID, Team6ID, Team7ID, Team8ID}
	canonicalIDs := []TeamID{Team1, Team2, Team3, Team4, Team5, Team6, Team7, Team8}
	for index := range canonicalIDs {
		if compatibilityIDs[index] != canonicalIDs[index] {
			t.Fatalf("compatibility ID at index %d = %q, want %q", index, compatibilityIDs[index], canonicalIDs[index])
		}
	}
}

func TestOrderedIDs(t *testing.T) {
	want := []TeamID{Team1, Team2, Team3, Team4, Team5, Team6, Team7, Team8}
	got := OrderedIDs()
	if len(got) != len(want) {
		t.Fatalf("OrderedIDs() length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("OrderedIDs()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
