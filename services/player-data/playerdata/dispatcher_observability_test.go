package playerdata

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/codec"
	"github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const dispatcherTestTraceID = "550e8400-e29b-41d4-a716-446655440012"

type dispatcherFailingStore struct{}

func (dispatcherFailingStore) LoadStats(protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, errors.New("upstream body must not be logged")
}

func (dispatcherFailingStore) RecordMatchResult(protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	return protocol.PlayerDataStats{}, false, errors.New("upstream body must not be logged")
}

func capturePlayerDataEvents(t *testing.T, fn func()) ([]map[string]any, observability.Status) {
	t.Helper()
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440013",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "player-data-test")
	if err != nil {
		t.Fatal(err)
	}

	fn()
	status := logging.EventStatus()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(payload)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	return events, status
}

func TestDispatcherCanonicalMatchResultEventsAreAcceptedAndJSONL(t *testing.T) {
	events, status := capturePlayerDataEvents(t, func() {
		dispatcher := NewDispatcher(NewMemoryStore())
		command := protocol.PlayerDataRecordMatchResult{
			Type:     protocol.PacketTypePlayerDataRecordMatchResult,
			ResultID: "result-1", MatchID: "match-1",
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-1"},
			Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer, TraceID: dispatcherTestTraceID},
		}
		first, err := codec.Encode(command)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dispatcher.Handle(first); err != nil {
			t.Fatal(err)
		}
		duplicate, err := codec.Encode(command)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dispatcher.Handle(duplicate); err != nil {
			t.Fatal(err)
		}
	})
	if status.RejectedCount != 0 || status.AcceptedCount != 2 {
		t.Fatalf("emitter status = %+v, want two accepted events and no rejection", status)
	}
	if len(events) != 2 || events[0]["event"] != string(observability.EventNameMatchResultReportSucceeded) || events[1]["event"] != string(observability.EventNameMatchResultDuplicateSuppressed) {
		t.Fatalf("events = %#v", events)
	}
	for _, event := range events {
		if event["trace_id"] != dispatcherTestTraceID || event["match_id"] != "match-1" {
			t.Fatalf("event context = %#v", event)
		}
		if _, unsafe := event["error"]; unsafe || strings.Contains(string(mustJSON(t, event)), "upstream body") {
			t.Fatalf("unsafe event payload = %#v", event)
		}
	}
}

func TestDispatcherCanonicalFailureEventsAreAcceptedWithoutRawErrors(t *testing.T) {
	events, status := capturePlayerDataEvents(t, func() {
		dispatcher := NewDispatcher(dispatcherFailingStore{})
		load, err := codec.Encode(protocol.PlayerDataLoadStats{
			Type:     protocol.PacketTypePlayerDataLoadStats,
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-1"},
			Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer, TraceID: dispatcherTestTraceID},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dispatcher.Handle(load); err != nil {
			t.Fatal(err)
		}
		record, err := codec.Encode(protocol.PlayerDataRecordMatchResult{
			Type:     protocol.PacketTypePlayerDataRecordMatchResult,
			ResultID: "result-1", MatchID: "match-1",
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-1"},
			Context:  protocol.PlayerDataRequestContext{PlayMode: PlayModeMultiplayer, TraceID: dispatcherTestTraceID},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dispatcher.Handle(record); err != nil {
			t.Fatal(err)
		}
	})
	if status.RejectedCount != 0 || status.AcceptedCount != 2 {
		t.Fatalf("emitter status = %+v, want two accepted events and no rejection", status)
	}
	if len(events) != 2 || events[0]["event"] != string(observability.EventNamePlayerDataReadFailed) || events[1]["event"] != string(observability.EventNamePlayerDataWriteFailed) {
		t.Fatalf("events = %#v", events)
	}
	for _, event := range events {
		fields, ok := event["fields"].(map[string]any)
		if !ok || event["trace_id"] != dispatcherTestTraceID || fields["error_code"] != "operation_failed" || fields["failure_mode"] != "store_failure" {
			t.Fatalf("event fields = %#v", event)
		}
		if strings.Contains(string(mustJSON(t, event)), "upstream body") {
			t.Fatalf("raw error leaked into event = %#v", event)
		}
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
