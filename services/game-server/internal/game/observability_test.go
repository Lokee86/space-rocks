package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	gameObservabilityTraceID  = "550e8400-e29b-41d4-a716-446655440101"
	gameObservabilityMatchID  = "room-1-match-1"
	gameObservabilityPlayerID = "Player-1"
)

func TestPlayerDiedObservabilityRecord(t *testing.T) {
	path := configureGameObservability(t)
	game := NewWithSeed(7)
	game.SetMatchContext(gameObservabilityMatchID, gameObservabilityTraceID)

	session := newPlayerSession(gameObservabilityPlayerID, physics.Vector2{X: 123.5, Y: 456.75})
	session.Lives = 2
	game.playerSessions[session.ID] = session
	game.entities.Players[session.ID] = session.NewShip(session.SpawnPosition)

	game.applyFatalPlayerDamage(session.ID, game.entities.Players[session.ID], "collision")
	assertAcceptedObservabilityRecords(t, 1)
	records := readGameObservabilityRecords(t, path)
	if len(records) != 1 {
		t.Fatalf("JSONL records = %d, want 1", len(records))
	}

	record := records[0]
	assertGameObservabilityContext(t, record, observability.EventNamePlayerDied, gameObservabilityPlayerID)
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields = %#v, want object", record["fields"])
	}
	if fields["reason_code"] != "collision" {
		t.Fatalf("reason_code = %#v, want collision", fields["reason_code"])
	}
	assertNumericGameField(t, fields, "lives", 1)
	assertNumericGameField(t, fields, "respawn_delay", constants.PlayerRespawnDelay)
	assertNumericGameField(t, fields, "x", 123.5)
	assertNumericGameField(t, fields, "y", 456.75)
}

func TestPlayerDiedObservabilityRecordUsesDevtoolsReason(t *testing.T) {
	path := configureGameObservability(t)
	game := NewWithSeed(13)
	game.SetMatchContext(gameObservabilityMatchID, gameObservabilityTraceID)

	session := newPlayerSession(gameObservabilityPlayerID, physics.Vector2{X: 30, Y: 40})
	session.Lives = 2
	game.playerSessions[session.ID] = session
	game.entities.Players[session.ID] = session.NewShip(session.SpawnPosition)

	if !NewControl(game).ApplyPlayerDefeat("debugger", session.ID) {
		t.Fatal("expected devtools defeat to be applied")
	}
	assertAcceptedObservabilityRecords(t, 1)
	records := readGameObservabilityRecords(t, path)
	if len(records) != 1 {
		t.Fatalf("JSONL records = %d, want 1", len(records))
	}
	record := records[0]
	assertGameObservabilityContext(t, record, observability.EventNamePlayerDied, gameObservabilityPlayerID)
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields = %#v, want object", record["fields"])
	}
	if fields["reason_code"] != "devtools" {
		t.Fatalf("reason_code = %#v, want devtools", fields["reason_code"])
	}
}

func TestRespawnBlockedObservabilityRecords(t *testing.T) {
	tests := []struct {
		name   string
		reason string
	}{
		{name: "session missing", reason: "session_missing"},
		{name: "cooldown or lives exhausted", reason: "respawn_cooldown_or_lives_exhausted"},
		{name: "already active", reason: "already_active"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := configureGameObservability(t)
			game := NewWithSeed(11)
			game.SetMatchContext(gameObservabilityMatchID, gameObservabilityTraceID)

			session := newPlayerSession(gameObservabilityPlayerID, physics.Vector2{X: 10, Y: 20})
			switch test.reason {
			case "session_missing":
			case "respawn_cooldown_or_lives_exhausted":
				session.Lives = 2
				session.RespawnCooldown = 1.25
				game.playerSessions[session.ID] = session
			case "already_active":
				game.playerSessions[session.ID] = session
				game.entities.Players[session.ID] = session.NewShip(session.SpawnPosition)
			}

			game.respawnPlayer(gameObservabilityPlayerID)
			assertAcceptedObservabilityRecords(t, 1)
			records := readGameObservabilityRecords(t, path)
			if len(records) != 1 {
				t.Fatalf("JSONL records = %d, want 1", len(records))
			}
			record := records[0]
			assertGameObservabilityContext(t, record, observability.EventNameRespawnBlocked, gameObservabilityPlayerID)
			fields, ok := record["fields"].(map[string]any)
			if !ok {
				t.Fatalf("fields = %#v, want object", record["fields"])
			}
			if fields["reason_code"] != test.reason {
				t.Fatalf("reason_code = %#v, want %s", fields["reason_code"], test.reason)
			}
			if test.reason == "respawn_cooldown_or_lives_exhausted" {
				assertNumericGameField(t, fields, "lives", 2)
				assertNumericGameField(t, fields, "respawn_cooldown", 1.25)
			}
		})
	}
}

func configureGameObservability(t *testing.T) string {
	t.Helper()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440100",
	}); err != nil {
		t.Fatal(err)
	}
	logging.Configure("info")
	path, err := logging.ConfigureFileOutput(t.TempDir(), "game-server-game-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = logging.CloseFileOutput()
		logging.Configure("warn")
	})
	return path
}

func readGameObservabilityRecords(t *testing.T, path string) []map[string]any {
	t.Helper()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func assertAcceptedObservabilityRecords(t *testing.T, want int) {
	t.Helper()
	status := logging.EventStatus()
	if status.AcceptedCount != uint64(want) || status.RejectedCount != 0 {
		t.Fatalf("event status = %+v, want %d accepted and no rejections", status, want)
	}
}

func assertGameObservabilityContext(t *testing.T, record map[string]any, event observability.EventName, playerID string) {
	t.Helper()
	if record["event"] != string(event) {
		t.Fatalf("event = %#v, want %s", record["event"], event)
	}
	if record["trace_id"] != gameObservabilityTraceID {
		t.Fatalf("trace_id = %#v, want %s", record["trace_id"], gameObservabilityTraceID)
	}
	if record["match_id"] != gameObservabilityMatchID {
		t.Fatalf("match_id = %#v, want %s", record["match_id"], gameObservabilityMatchID)
	}
	if record["player_id"] != playerID {
		t.Fatalf("player_id = %#v, want %s", record["player_id"], playerID)
	}
}

func assertNumericGameField(t *testing.T, fields map[string]any, key string, want float64) {
	t.Helper()
	got, ok := fields[key].(float64)
	if !ok || got != want {
		t.Fatalf("%s = %#v, want %v", key, fields[key], want)
	}
}
