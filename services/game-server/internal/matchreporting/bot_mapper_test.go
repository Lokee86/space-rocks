package matchreporting

import (
	"testing"

	serverplayerdata "github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func TestBuildRecordMatchResultCommandsSkipsBotResults(t *testing.T) {
	summary := serverplayerdata.MatchResultSummary{
		MatchID: "room-1-match-1",
		Mode:    serverplayerdata.MatchModeMultiplayer,
		Players: []serverplayerdata.PlayerMatchSummary{
			{GamePlayerID: "player-human", AccountID: "account-1", Score: 10},
			{GamePlayerID: "player-bot", IsBot: true, Score: 20, Won: true},
		},
	}

	commands := BuildRecordMatchResultCommands(summary)
	if len(commands) != 1 {
		t.Fatalf("expected only human result command, got %d", len(commands))
	}
	if commands[0].ResultID != "room-1-match-1:player-human" {
		t.Fatalf("unexpected result command: %+v", commands[0])
	}
}
