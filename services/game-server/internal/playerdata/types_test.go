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
