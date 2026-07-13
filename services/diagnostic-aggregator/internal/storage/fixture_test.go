package storage

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEventsFixtureContract(t *testing.T) {
	file, err := os.Open("testdata/events.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	required := []string{
		"event_id", "schema_version", "timestamp", "ingested_at", "level",
		"event", "service", "service_instance_id", "environment", "build_version",
	}
	idFields := map[string]bool{
		"event_id": true, "service_instance_id": true, "trace_id": true,
		"diagnostic_report_id": true, "audit_event_id": true,
	}
	forbidden := map[string]bool{
		"authorization": true, "cookie": true, "access_token": true,
		"refresh_token": true, "oauth_code": true, "client_secret": true,
		"password": true, "raw_profile": true, "raw_packet": true,
	}

	scanner := bufio.NewScanner(file)
	var previous time.Time
	seenEvents := make(map[string]bool)
	traceCounts := make(map[string]int)
	var records []map[string]any
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %d is invalid JSON: %v", lineNumber, err)
		}
		for _, field := range required {
			if _, ok := record[field]; !ok {
				t.Fatalf("line %d is missing required field %q", lineNumber, field)
			}
		}

		timestamp := parseFixtureTime(t, lineNumber, record, "timestamp")
		parseFixtureTime(t, lineNumber, record, "ingested_at")
		if !previous.IsZero() && timestamp.Before(previous) {
			t.Fatalf("line %d timestamp %s is out of order", lineNumber, timestamp)
		}
		previous = timestamp

		eventID := stringField(t, lineNumber, record, "event_id")
		if seenEvents[eventID] {
			t.Fatalf("line %d repeats event_id %q", lineNumber, eventID)
		}
		seenEvents[eventID] = true
		for field := range idFields {
			if value, ok := record[field]; ok {
				assertCanonicalUUID(t, lineNumber, field, value)
			}
		}
		if _, ok := record["trace_id"]; ok {
			traceCounts[stringField(t, lineNumber, record, "trace_id")]++
		}
		assertNoForbiddenKeys(t, lineNumber, record, forbidden, "")
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatal("fixture contains no records")
	}

	repeatedTrace := false
	for _, count := range traceCounts {
		if count > 1 {
			repeatedTrace = true
			break
		}
	}
	if !repeatedTrace {
		t.Fatal("fixture contains no repeated trace chain")
	}

	hasDiagnostic := false
	hasAudit := false
	for _, record := range records {
		if _, ok := record["diagnostic_report_id"]; ok {
			hasDiagnostic = true
		}
		if required, ok := record["audit_required"].(bool); ok && required {
			if _, hasID := record["audit_event_id"]; !hasID {
				t.Fatal("audit-required event is missing audit_event_id")
			}
			hasAudit = true
		}
	}
	if !hasDiagnostic {
		t.Fatal("fixture contains no diagnostic_report_id")
	}
	if !hasAudit {
		t.Fatal("fixture contains no audit-required event with audit_event_id")
	}
}

func parseFixtureTime(t *testing.T, line int, record map[string]any, field string) time.Time {
	t.Helper()
	value := stringField(t, line, record, field)
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("line %d field %q is not RFC3339Nano: %v", line, field, err)
	}
	return parsed
}

func stringField(t *testing.T, line int, record map[string]any, field string) string {
	t.Helper()
	value, ok := record[field].(string)
	if !ok || value == "" {
		t.Fatalf("line %d field %q is not a non-empty string", line, field)
	}
	return value
}

func assertCanonicalUUID(t *testing.T, line int, field string, value any) {
	t.Helper()
	text, ok := value.(string)
	if !ok || len(text) != 36 || text != strings.ToLower(text) {
		t.Fatalf("line %d field %q is not canonical UUID text: %v", line, field, value)
	}
	parts := strings.Split(text, "-")
	if len(parts) != 5 || len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Fatalf("line %d field %q has invalid UUID shape: %q", line, field, text)
	}
	if _, err := hex.DecodeString(strings.ReplaceAll(text, "-", "")); err != nil {
		t.Fatalf("line %d field %q has invalid UUID hex: %q", line, field, text)
	}
}

func assertNoForbiddenKeys(t *testing.T, line int, value any, forbidden map[string]bool, path string) {
	t.Helper()
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if forbidden[strings.ToLower(key)] {
				t.Fatalf("line %d contains forbidden key %q at %s", line, key, path)
			}
			childPath := key
			if path != "" {
				childPath = fmt.Sprintf("%s.%s", path, key)
			}
			assertNoForbiddenKeys(t, line, child, forbidden, childPath)
		}
	case []any:
		for index, child := range value {
			assertNoForbiddenKeys(t, line, child, forbidden, fmt.Sprintf("%s[%d]", path, index))
		}
	}
}
