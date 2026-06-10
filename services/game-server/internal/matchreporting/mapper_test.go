package matchreporting

import (
	"testing"

	serverplayerdata "github.com/Lokee86/space-rocks/server/internal/playerdata"
)

func TestBuildRecordMatchResultCommandsUsesAccountIdentity(t *testing.T) {
	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				AccountID:    "acct-1",
				Score:        275,
				ShipDeaths:   3,
				Won:          true,
			},
		},
	}

	commands := BuildRecordMatchResultCommands(summary)
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	command := commands[0]
	if command.ResultID != "room-1-match-1:Player-1" {
		t.Fatalf("expected ResultID %q, got %q", "room-1-match-1:Player-1", command.ResultID)
	}
	if command.Identity.IdentityKind != "authenticated_account" {
		t.Fatalf("expected IdentityKind %q, got %q", "authenticated_account", command.Identity.IdentityKind)
	}
	if command.Identity.AccountID != "acct-1" {
		t.Fatalf("expected AccountID %q, got %q", "acct-1", command.Identity.AccountID)
	}
	if command.Score != 275 {
		t.Fatalf("expected Score 275, got %d", command.Score)
	}
	if command.ShipDeaths != 3 {
		t.Fatalf("expected ShipDeaths 3, got %d", command.ShipDeaths)
	}
	if !command.Won {
		t.Fatal("expected Won to be true")
	}
}

func TestBuildRecordMatchResultCommandsUsesLocalProfileIdentity(t *testing.T) {
	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID:   "Player-1",
				LocalProfileID: "local-1",
				Score:          180,
				ShipDeaths:     1,
				Won:            false,
			},
		},
	}

	commands := BuildRecordMatchResultCommands(summary)
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	command := commands[0]
	if command.Identity.IdentityKind != "local_profile" {
		t.Fatalf("expected IdentityKind %q, got %q", "local_profile", command.Identity.IdentityKind)
	}
	if command.Identity.LocalProfileID != "local-1" {
		t.Fatalf("expected LocalProfileID %q, got %q", "local-1", command.Identity.LocalProfileID)
	}
	if command.Identity.AccountID != "" {
		t.Fatalf("expected AccountID to be empty, got %q", command.Identity.AccountID)
	}
}

func TestBuildRecordMatchResultCommandsUsesGuestIdentity(t *testing.T) {
	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				Score:        10,
			},
		},
	}

	commands := BuildRecordMatchResultCommands(summary)
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	command := commands[0]
	if command.Identity.IdentityKind != "guest" {
		t.Fatalf("expected IdentityKind %q, got %q", "guest", command.Identity.IdentityKind)
	}
	if command.Identity.AccountID != "" {
		t.Fatalf("expected AccountID to be empty, got %q", command.Identity.AccountID)
	}
	if command.Identity.LocalProfileID != "" {
		t.Fatalf("expected LocalProfileID to be empty, got %q", command.Identity.LocalProfileID)
	}
}

func TestBuildRecordMatchResultCommandsPrefersAccountIdentityOverLocalProfileIdentity(t *testing.T) {
	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Players: []serverplayerdata.PlayerMatchSummary{
			{
				GamePlayerID:   "Player-1",
				AccountID:      "acct-1",
				LocalProfileID: "local-1",
			},
		},
	}

	commands := BuildRecordMatchResultCommands(summary)
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	command := commands[0]
	if command.Identity.IdentityKind != "authenticated_account" {
		t.Fatalf("expected IdentityKind %q, got %q", "authenticated_account", command.Identity.IdentityKind)
	}
	if command.Identity.AccountID != "acct-1" {
		t.Fatalf("expected AccountID %q, got %q", "acct-1", command.Identity.AccountID)
	}
	if command.Identity.LocalProfileID != "" {
		t.Fatalf("expected LocalProfileID to be empty, got %q", command.Identity.LocalProfileID)
	}
}
