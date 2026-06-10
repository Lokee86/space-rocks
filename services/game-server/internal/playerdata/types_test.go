package playerdata

import "testing"

func TestStatsZeroValueDefaults(t *testing.T) {
	var stats Stats

	if stats.TotalScore != 0 {
		t.Fatalf("TotalScore = %d, want 0", stats.TotalScore)
	}
	if stats.HighScore != 0 {
		t.Fatalf("HighScore = %d, want 0", stats.HighScore)
	}
	if stats.ShipDeaths != 0 {
		t.Fatalf("ShipDeaths = %d, want 0", stats.ShipDeaths)
	}
	if stats.GamesPlayed != 0 {
		t.Fatalf("GamesPlayed = %d, want 0", stats.GamesPlayed)
	}
	if stats.Wins != 0 {
		t.Fatalf("Wins = %d, want 0", stats.Wins)
	}
}

func TestLocalPlayerSummaryCanOmitAccountID(t *testing.T) {
	summary := PlayerMatchSummary{
		GamePlayerID:   "Player-1",
		LocalProfileID: "local-profile-1",
		Score:          120,
	}

	if summary.AccountID != "" {
		t.Fatalf("AccountID = %q, want empty string", summary.AccountID)
	}
}

func TestAccountPlayerSummaryCanOmitLocalProfileID(t *testing.T) {
	summary := PlayerMatchSummary{
		GamePlayerID: "Player-2",
		AccountID:    "account-42",
		Score:        220,
	}

	if summary.LocalProfileID != "" {
		t.Fatalf("LocalProfileID = %q, want empty string", summary.LocalProfileID)
	}
}

func TestMatchModeConstants(t *testing.T) {
	if MatchModeSinglePlayer != "single_player" {
		t.Fatalf("MatchModeSinglePlayer = %q, want %q", MatchModeSinglePlayer, "single_player")
	}
	if MatchModeMultiplayer != "multiplayer" {
		t.Fatalf("MatchModeMultiplayer = %q, want %q", MatchModeMultiplayer, "multiplayer")
	}
}

func TestMatchResultSummaryUsesMatchMode(t *testing.T) {
	summary := MatchResultSummary{
		MatchID: "match-1",
		Mode:    MatchModeMultiplayer,
	}

	if summary.Mode != MatchModeMultiplayer {
		t.Fatalf("Mode = %q, want %q", summary.Mode, MatchModeMultiplayer)
	}
}
