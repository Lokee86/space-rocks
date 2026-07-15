package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	playerlogging "github.com/Lokee86/space-rocks/player-data/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func TestPlayerDataObservabilityUnavailableIsAcceptedAndWritten(t *testing.T) {
	if err := playerlogging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: playerlogging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440018",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := playerlogging.ConfigureFileOutput(t.TempDir(), "player-data-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = playerlogging.CloseFileOutput() })

	result := emitPlayerDataObservabilityUnavailable("runtime_degraded")
	if !result.Accepted {
		t.Fatalf("result = %+v", result)
	}
	status := playerlogging.EventStatus()
	if status.RejectedCount != 0 || status.AcceptedCount != 1 {
		t.Fatalf("emitter status = %+v", status)
	}
	if err := playerlogging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(payload))), &event); err != nil {
		t.Fatal(err)
	}
	fields, ok := event["fields"].(map[string]any)
	if !ok || event["event"] != string(observability.EventNameObservabilityUnavailable) || fields["failure_mode"] != "runtime_degraded" {
		t.Fatalf("event = %#v", event)
	}
}
