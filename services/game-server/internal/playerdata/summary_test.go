package playerdata

import "testing"

func TestBuildMatchResultSummaryCopiesMatchID(t *testing.T) {
	got := BuildMatchResultSummary("match-123", MatchModeMultiplayer, nil)

	if got.MatchID != "match-123" {
		t.Fatalf("MatchID = %q, want %q", got.MatchID, "match-123")
	}
}

func TestBuildMatchResultSummaryCopiesMode(t *testing.T) {
	got := BuildMatchResultSummary("match-123", MatchModeSinglePlayer, nil)

	if got.Mode != MatchModeSinglePlayer {
		t.Fatalf("Mode = %q, want %q", got.Mode, MatchModeSinglePlayer)
	}
}

func TestBuildMatchResultSummaryIncludesPlayers(t *testing.T) {
	players := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100},
		{GamePlayerID: "Player-2", Score: 250},
	}

	got := BuildMatchResultSummary("match-123", MatchModeMultiplayer, players)

	if len(got.Players) != len(players) {
		t.Fatalf("len(Players) = %d, want %d", len(got.Players), len(players))
	}
}

func TestBuildMatchResultSummaryAppliesMultiplayerWinnerResolution(t *testing.T) {
	players := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100},
		{GamePlayerID: "Player-2", Score: 250},
		{GamePlayerID: "Player-3", Score: 175},
	}

	got := BuildMatchResultSummary("match-123", MatchModeMultiplayer, players)

	if got.Players[1].Won != true {
		t.Fatalf("got.Players[1].Won = %t, want true", got.Players[1].Won)
	}
	if got.Players[0].Won != false {
		t.Fatalf("got.Players[0].Won = %t, want false", got.Players[0].Won)
	}
	if got.Players[2].Won != false {
		t.Fatalf("got.Players[2].Won = %t, want false", got.Players[2].Won)
	}
}

func TestBuildMatchResultSummarySinglePlayerWinnerIsFalse(t *testing.T) {
	players := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100, Won: true},
		{GamePlayerID: "Player-2", Score: 250, Won: true},
	}

	got := BuildMatchResultSummary("match-123", MatchModeSinglePlayer, players)

	for i := range got.Players {
		if got.Players[i].Won {
			t.Fatalf("got.Players[%d].Won = true, want false", i)
		}
	}
}

func TestBuildMatchResultSummaryPreservesAccountID(t *testing.T) {
	players := []PlayerMatchSummary{
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
	}

	got := BuildMatchResultSummary("match-123", MatchModeMultiplayer, players)

	if got.Players[0].AccountID != "account-1" || got.Players[1].AccountID != "account-2" {
		t.Fatalf("AccountID values changed: got %#v", got.Players)
	}
}

func TestBuildMatchResultSummaryPreservesLocalProfileID(t *testing.T) {
	players := []PlayerMatchSummary{
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
	}

	got := BuildMatchResultSummary("match-123", MatchModeMultiplayer, players)

	if got.Players[0].LocalProfileID != "local-1" || got.Players[1].LocalProfileID != "local-2" {
		t.Fatalf("LocalProfileID values changed: got %#v", got.Players)
	}
}

func TestBuildMatchResultSummaryDoesNotMutateInputPlayers(t *testing.T) {
	players := []PlayerMatchSummary{
		{GamePlayerID: "Player-1", Score: 100, Won: true},
		{GamePlayerID: "Player-2", Score: 250, Won: false},
	}

	got := BuildMatchResultSummary("match-456", MatchModeSinglePlayer, players)
	got.Players[0].Score = 999

	if players[0].Won != true || players[1].Won != false {
		t.Fatalf("input players were mutated")
	}
	if players[0].Score != 100 || players[1].Score != 250 {
		t.Fatalf("input player scores were mutated")
	}
	if got.Players[0].Score != 999 {
		t.Fatalf("got.Players[0].Score = %d, want 999", got.Players[0].Score)
	}
}
