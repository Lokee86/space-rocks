package playerdata

import "testing"

func TestResolveWinnersSinglePlayerNeverSetsWonTrue(t *testing.T) {
	input := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100, Won: true},
		{GamePlayerID: "Player-2", Score: 50, Won: true},
	}

	got := ResolveWinners(MatchModeSinglePlayer, input)

	if len(got) != len(input) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(input))
	}
	for i := range got {
		if got[i].Won {
			t.Fatalf("got[%d].Won = true, want false", i)
		}
	}
}

func TestResolveWinnersMultiplayerHighestScoreGetsWonTrue(t *testing.T) {
	input := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100},
		{GamePlayerID: "Player-2", Score: 250},
		{GamePlayerID: "Player-3", Score: 175},
	}

	got := ResolveWinners(MatchModeMultiplayer, input)

	if len(got) != len(input) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(input))
	}
	if !got[1].Won {
		t.Fatalf("got[1].Won = false, want true")
	}
}

func TestResolveWinnersMultiplayerLowerScoresGetWonFalse(t *testing.T) {
	input := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100},
		{GamePlayerID: "Player-2", Score: 250},
		{GamePlayerID: "Player-3", Score: 175},
	}

	got := ResolveWinners(MatchModeMultiplayer, input)

	if got[0].Won {
		t.Fatalf("got[0].Won = true, want false")
	}
	if got[2].Won {
		t.Fatalf("got[2].Won = true, want false")
	}
}

func TestResolveWinnersMultiplayerTiedHighestScoreGivesNoWins(t *testing.T) {
	input := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 250, Won: true},
		{GamePlayerID: "Player-2", Score: 250, Won: true},
		{GamePlayerID: "Player-3", Score: 175, Won: true},
	}

	got := ResolveWinners(MatchModeMultiplayer, input)

	for i := range got {
		if got[i].Won {
			t.Fatalf("got[%d].Won = true, want false for tie", i)
		}
	}
}

func TestResolveWinnersMultiplayerEmptyPlayersReturnsEmptySlice(t *testing.T) {
	got := ResolveWinners(MatchModeMultiplayer, []PlayerMatchSummary{})

	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0", len(got))
	}
	if got == nil {
		t.Fatalf("got = nil, want empty slice")
	}
}

func TestResolveWinnersDoesNotMutateInputSlice(t *testing.T) {
	input := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100, Won: false},
		{GamePlayerID: "Player-2", Score: 250, Won: false},
	}

	got := ResolveWinners(MatchModeMultiplayer, input)
	got[1].Score = 999

	if input[0].Won || input[1].Won {
		t.Fatalf("input slice was mutated")
	}
	if input[1].Score != 250 {
		t.Fatalf("input[1].Score = %d, want 250", input[1].Score)
	}
	if got[1].Score != 999 {
		t.Fatalf("got[1].Score = %d, want 999", got[1].Score)
	}
}

func TestResolveWinnersIgnoresAccountAndLocalProfileIDs(t *testing.T) {
	input := []PlayerMatchSummary{
		{
			GamePlayerID:   "Player-1",
			AccountID:      "account-1",
			LocalProfileID: "local-1",
			Score:          100,
		},
		{
			GamePlayerID:   "Player-2",
			AccountID:      "account-2",
			LocalProfileID: "local-2",
			Score:          250,
		},
		{
			GamePlayerID:   "Player-3",
			AccountID:      "account-1",
			LocalProfileID: "local-9",
			Score:          175,
		},
	}

	got := ResolveWinners(MatchModeMultiplayer, input)

	if !got[1].Won {
		t.Fatalf("got[1].Won = false, want true")
	}
	if got[0].Won || got[2].Won {
		t.Fatalf("winner selection should ignore account and local profile IDs")
	}
}
