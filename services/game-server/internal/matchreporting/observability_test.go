package matchreporting

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func TestPlayerDataPacketBoundaryEventsAreAcceptedAndWritten(t *testing.T) {
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440016",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "game-server-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logging.CloseFileOutput() })

	command := protocol.PlayerDataRecordMatchResult{
		Type:     protocol.PacketTypePlayerDataRecordMatchResult,
		ResultID: "match-1:player-1", MatchID: "match-1",
		Identity: protocol.PlayerDataIdentity{AccountID: "acct-1"},
		Context:  protocol.PlayerDataRequestContext{TraceID: "550e8400-e29b-41d4-a716-446655440017"},
	}
	emitPlayerDataPacketBoundaryEvent(command, errors.New("decode failed"))
	command.Type = "unknown_packet"
	emitPlayerDataPacketBoundaryEvent(command, errors.New("unknown packet type"))

	status := logging.EventStatus()
	if status.RejectedCount != 0 || status.AcceptedCount != 2 {
		t.Fatalf("emitter status = %+v, want two accepted events and no rejection", status)
	}
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) != 2 {
		t.Fatalf("JSONL lines = %d, want 2", len(lines))
	}
	wantEvents := []string{string(observability.EventNamePacketDecodeFailed), string(observability.EventNamePacketRouteUnknown)}
	for index, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event["event"] != wantEvents[index] || event["trace_id"] != command.Context.TraceID || event["match_id"] != "match-1" {
			t.Fatalf("event = %#v", event)
		}
		if strings.Contains(line, "decode failed") || strings.Contains(line, "unknown packet type") {
			t.Fatalf("raw error leaked into event = %s", line)
		}
	}
}
