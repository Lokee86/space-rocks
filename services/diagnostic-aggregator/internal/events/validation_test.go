package events

import (
	"errors"
	"strings"
	"testing"
)

const validEvent = `{"event_id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"ingest_accepted","service":"diagnostic-aggregator","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001","fields":{"count":1}}`

func TestDecodeValidEvent(t *testing.T) {
	event, err := Decode([]byte(validEvent))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if event.Event != "ingest_accepted" || event.Fields["count"] != float64(1) {
		t.Fatalf("decoded event = %#v", event)
	}
}

func TestDecodeRejectsInputWithStableCode(t *testing.T) {
	tests := []struct {
		name string
		json string
		code string
	}{
		{"unknown field", validEvent[:len(validEvent)-1] + `,"secret":"value"}`, CodeUnknownField},
		{"trailing json", validEvent + ` {}`, CodeTrailingJSON},
		{"invalid uuid", strings.Replace(validEvent, "550e8400-e29b-41d4-a716-446655440000", "not-a-uuid", 1), CodeInvalidUUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Decode([]byte(test.json))
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Code != test.code {
				t.Fatalf("error = %v, want code %q", err, test.code)
			}
		})
	}
}

func TestValidateRequiredAndConditionalFields(t *testing.T) {
	base, err := Decode([]byte(validEvent))
	if err != nil {
		t.Fatal(err)
	}
	base.EventID = ""
	assertCode(t, base, CodeRequiredFieldMissing)

	base, _ = Decode([]byte(validEvent))
	base.SchemaVersion = 2
	assertCode(t, base, CodeUnsupportedSchema)

	base, _ = Decode([]byte(validEvent))
	base.Level = "notice"
	assertCode(t, base, CodeUnsupportedLevel)

	base, _ = Decode([]byte(validEvent))
	base.Event = "Not-Snake"
	assertCode(t, base, CodeInvalidEventName)

	base, _ = Decode([]byte(validEvent))
	base.AuditRequired = true
	assertCode(t, base, CodeAuditTypeRequired)
}

func TestValidateContractNumericAndSnakeCaseFields(t *testing.T) {
	base, err := Decode([]byte(validEvent))
	if err != nil {
		t.Fatal(err)
	}
	duration := -1.0
	base.DurationMS = &duration
	assertCode(t, base, CodeInvalidDuration)

	base, _ = Decode([]byte(validEvent))
	processID := int64(-1)
	base.ProcessID = &processID
	assertCode(t, base, CodeInvalidProcessID)

	for _, field := range []string{"AuditType", "Action", "ReasonCode"} {
		base, _ = Decode([]byte(validEvent))
		switch field {
		case "AuditType":
			base.AuditType = "Audit-Type"
		case "Action":
			base.Action = "Do Thing"
		case "ReasonCode":
			base.ReasonCode = "UPPER_CASE"
		}
		assertCode(t, base, CodeInvalidFieldName)
	}
}

func assertCode(t *testing.T, event Event, want string) {
	t.Helper()
	var validationErr *ValidationError
	if err := Validate(event); !errors.As(err, &validationErr) || validationErr.Code != want {
		t.Fatalf("Validate() error = %v, want code %q", err, want)
	}
}
