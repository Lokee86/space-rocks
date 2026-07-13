package observability

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventJSONIncludesRequiredCoreAndOmitsOptionalContext(t *testing.T) {
	event := Event{
		SchemaVersion:     CurrentSchemaVersion,
		EventID:           "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:         time.Unix(1, 0).UTC(),
		Level:             LevelInfo,
		Event:             "service_started",
		Service:           "game-server",
		ServiceInstanceID: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal event JSON: %v", err)
	}

	for _, key := range []string{
		"schema_version",
		"event_id",
		"timestamp",
		"level",
		"event",
		"service",
		"service_instance_id",
	} {
		if _, ok := fields[key]; !ok {
			t.Errorf("required JSON key %q is missing", key)
		}
	}

	for _, key := range []string{
		"trace_id",
		"environment",
		"build_version",
		"category",
		"message",
		"request_id",
		"session_id",
		"room_id",
		"match_id",
		"player_id",
		"account_id",
	} {
		if _, ok := fields[key]; ok {
			t.Errorf("optional JSON key %q should be omitted", key)
		}
	}
}

func TestObservabilitySchemaAndLevels(t *testing.T) {
	if CurrentSchemaVersion != 1 {
		t.Fatalf("CurrentSchemaVersion = %d, want 1", CurrentSchemaVersion)
	}

	levels := map[Level]string{
		LevelDebug:    "debug",
		LevelInfo:     "info",
		LevelWarn:     "warn",
		LevelError:    "error",
		LevelCritical: "critical",
	}
	for level, want := range levels {
		if string(level) != want {
			t.Errorf("level %q = %q, want %q", level, level, want)
		}
	}
}

func TestEventBatchJSONContainsOneEvent(t *testing.T) {
	batch := EventBatch{
		Events: []Event{{
			SchemaVersion:     CurrentSchemaVersion,
			EventID:           "550e8400-e29b-41d4-a716-446655440000",
			Timestamp:         time.Unix(1, 0).UTC(),
			Level:             LevelInfo,
			Event:             "service_started",
			Service:           "game-server",
			ServiceInstanceID: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
		}},
	}

	payload, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("marshal event batch: %v", err)
	}

	var envelope struct {
		Events []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("unmarshal event batch JSON: %v", err)
	}
	if len(envelope.Events) != 1 {
		t.Fatalf("event batch contains %d events, want 1", len(envelope.Events))
	}
}
