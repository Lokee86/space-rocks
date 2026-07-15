package observabilityevent

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	uuidPattern     = regexp.MustCompile(UUIDPattern)
	fieldKeyPattern = regexp.MustCompile(FreeFormKeyPattern)
)

// Emitter validates, redacts, serializes, and writes one canonical event.
type Emitter struct {
	config  Config
	service ServiceDefinition

	mu          sync.Mutex
	status      Status
	lastWarning time.Time
}

func New(config Config) (*Emitter, error) {
	service, ok := ServiceDefinitionFor(config.Service)
	if !ok {
		return nil, errors.New("observabilityevent: unknown service key")
	}
	if strings.TrimSpace(config.Environment) == "" || strings.TrimSpace(config.BuildVersion) == "" {
		return nil, errors.New("observabilityevent: environment and build version are required")
	}
	if !uuidPattern.MatchString(config.ServiceInstanceID) {
		return nil, errors.New("observabilityevent: service instance ID must be a UUID")
	}
	if config.Sink == nil {
		return nil, errors.New("observabilityevent: sink is required")
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.NewEventID == nil {
		config.NewEventID = newUUID
	}
	return &Emitter{config: config, service: service}, nil
}

// Emit emits an ordinary canonical event. Bridge-only events are always
// rejected here; adapters must use EmitLegacy.
func (e *Emitter) Emit(request Request) Result {
	definition, ok := EventDefinitionFor(request.Event)
	if !ok {
		return e.reject(RejectionCodeUnknownEvent, "")
	}
	if definition.BridgeOnly {
		return e.reject(RejectionCodeBridgeEventForbidden, "")
	}
	return e.emit(definition, definition.DefaultLevel, definition.Category, request.Message, request.Context, request.Fields)
}

// EmitLegacy emits the generated bridge-only log_message event. It is intended
// only for existing logger adapters while semantic event call sites migrate.
func (e *Emitter) EmitLegacy(request LegacyRequest) Result {
	definition := EventDefinitions[EventNameLogMessage]
	if !IsCanonicalLevel(request.Level) {
		return e.reject(RejectionCodeInvalidFieldType, "level")
	}
	if strings.TrimSpace(request.Category) == "" {
		return e.reject(RejectionCodeInvalidFieldType, "category")
	}
	fields := copyFields(request.Fields)
	if request.LegacyEvent != "" && request.LegacyEvent != string(EventNameLogMessage) {
		if fields == nil {
			fields = Fields{}
		}
		fields["legacy_event"] = request.LegacyEvent
	}
	return e.emit(definition, request.Level, request.Category, request.Message, request.Context, fields)
}

func (e *Emitter) emit(definition EventDefinition, level Level, category, message string, context Context, fields Fields) Result {
	if !serviceAllowed(definition.Services, e.config.Service) {
		return e.reject(RejectionCodeServiceNotAllowed, "")
	}
	if definition.TraceRequired && context.TraceID == "" {
		return e.reject(RejectionCodeTraceRequired, "trace_id")
	}
	if key := invalidContextUUID(context); key != "" {
		return e.reject(RejectionCodeInvalidUuid, key)
	}
	if key := oversizedContextValue(context); key != "" {
		return e.reject(RejectionCodeStringLimitExceeded, key)
	}

	safeFields, redacted, code, key := sanitizeFields(fields)
	if code != "" {
		return e.reject(code, key)
	}
	if utf8.RuneCountInString(message) > MaxStringBytes || len([]byte(message)) > MaxStringBytes || len([]byte(category)) > MaxStringBytes {
		return e.reject(RejectionCodeStringLimitExceeded, "message")
	}

	eventID, err := e.config.NewEventID()
	if err != nil || !uuidPattern.MatchString(eventID) {
		return e.reject(RejectionCodeInvalidUuid, "event_id")
	}
	record := map[string]any{
		"timestamp":           e.config.Now().UTC().Format(time.RFC3339Nano),
		"level":               string(level),
		"event":               string(definition.Name),
		"event_id":            eventID,
		"service":             e.service.EmittedName,
		"environment":         e.config.Environment,
		"build_version":       e.config.BuildVersion,
		"schema_version":      SchemaVersion,
		"service_instance_id": e.config.ServiceInstanceID,
		"category":            category,
		"retention_tier":      string(definition.RetentionTier),
	}
	if message != "" {
		record["message"] = message
	}
	addContext(record, context)
	if e.config.WorkerID != "" {
		record["worker_id"] = e.config.WorkerID
	}
	if e.config.PID != 0 {
		record["pid"] = e.config.PID
	}
	if len(safeFields) > 0 {
		record["fields"] = safeFields
	}
	if len(record) > MaxEventFields {
		return e.reject(RejectionCodeFieldLimitExceeded, "")
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return e.reject(RejectionCodeSerializationFailed, "")
	}
	if len(payload) > MaxEventBytes {
		return e.reject(RejectionCodeEventTooLarge, "")
	}
	if err := e.config.Sink.WriteRecord(payload, safeConsoleLine(level, category, definition.Name, message)); err != nil {
		return e.writeFailure(err)
	}

	e.mu.Lock()
	e.status.AcceptedCount++
	if redacted {
		e.status.RedactedCount++
	}
	e.mu.Unlock()
	return Result{Accepted: true, Redacted: redacted}
}

func sanitizeFields(fields Fields) (Fields, bool, RejectionCode, string) {
	if len(fields) > MaxFreeFormFields {
		return nil, false, RejectionCodeFieldLimitExceeded, ""
	}
	result := make(Fields, len(fields))
	redacted := false
	for key, value := range fields {
		action, matched, ambiguous := redactionAction(key)
		if ambiguous {
			return nil, false, RejectionCodeUnsafeField, key
		}
		if matched {
			switch action {
			case RedactionActionReject:
				return nil, false, RejectionCodeUnsafeField, key
			case RedactionActionRedact:
				result[key] = RedactionReplacementMarker
				redacted = true
				continue
			default:
				return nil, false, RejectionCodeRedactionFailed, key
			}
		}
		if !fieldKeyPattern.MatchString(key) {
			return nil, false, RejectionCodeInvalidFieldKey, key
		}
		normalized, ok, isNull := scalarValue(value)
		if isNull {
			return nil, false, RejectionCodeNullNotAllowed, key
		}
		if !ok {
			return nil, false, RejectionCodeInvalidFieldType, key
		}
		if text, ok := normalized.(string); ok && len([]byte(text)) > MaxFreeFormValueBytes {
			return nil, false, RejectionCodeStringLimitExceeded, key
		}
		result[key] = normalized
	}
	return result, redacted, "", ""
}

func scalarValue(value any) (any, bool, bool) {
	if value == nil {
		return nil, false, true
	}
	switch typed := value.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32:
		return typed, true, false
	case uint64:
		if typed > math.MaxInt64 {
			return nil, false, false
		}
		return typed, true, false
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return nil, false, false
		}
		return typed, true, false
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, false, false
		}
		return typed, true, false
	default:
		return nil, false, false
	}
}

func redactionAction(key string) (RedactionAction, bool, bool) {
	candidate := key
	if !RedactionCaseSensitive {
		candidate = strings.ToLower(candidate)
	}
	actions := map[RedactionAction]struct{}{}
	for _, rule := range RedactionExactRules {
		for _, match := range rule.Matches {
			if candidate == match {
				actions[rule.Action] = struct{}{}
			}
		}
	}
	for _, rule := range RedactionFragmentRules {
		for _, match := range rule.Matches {
			if strings.Contains(candidate, match) {
				actions[rule.Action] = struct{}{}
			}
		}
	}
	if len(actions) == 0 {
		return "", false, false
	}
	if len(actions) > 1 {
		return RedactionAmbiguousMatchAction, true, true
	}
	for action := range actions {
		return action, true, false
	}
	return RedactionFailureAction, true, true
}

func addContext(record map[string]any, context Context) {
	values := map[string]string{
		"trace_id": context.TraceID, "session_id": context.SessionID,
		"room_id": context.RoomID, "player_id": context.PlayerID,
		"account_id": context.AccountID, "match_id": context.MatchID,
		"request_id": context.RequestID, "diagnostic_report_id": context.DiagnosticReportID,
		"audit_event_id": context.AuditEventID, "route": context.Route,
		"packet_type": context.PacketType,
	}
	for key, value := range values {
		if value != "" {
			record[key] = value
		}
	}
	if context.DurationMS != nil {
		record["duration_ms"] = *context.DurationMS
	}
}

func invalidContextUUID(context Context) string {
	for key, value := range map[string]string{
		"trace_id": context.TraceID, "diagnostic_report_id": context.DiagnosticReportID,
		"audit_event_id": context.AuditEventID,
	} {
		if value != "" && !uuidPattern.MatchString(value) {
			return key
		}
	}
	return ""
}

func oversizedContextValue(context Context) string {
	for key, value := range map[string]string{
		"trace_id": context.TraceID, "session_id": context.SessionID,
		"room_id": context.RoomID, "player_id": context.PlayerID,
		"account_id": context.AccountID, "match_id": context.MatchID,
		"request_id": context.RequestID, "diagnostic_report_id": context.DiagnosticReportID,
		"audit_event_id": context.AuditEventID, "route": context.Route,
		"packet_type": context.PacketType,
	} {
		if len([]byte(value)) > MaxStringBytes {
			return key
		}
	}
	return ""
}

func serviceAllowed(allowed []ServiceKey, service ServiceKey) bool {
	for _, candidate := range allowed {
		if candidate == service {
			return true
		}
	}
	return false
}

func copyFields(fields Fields) Fields {
	if fields == nil {
		return nil
	}
	result := make(Fields, len(fields))
	for key, value := range fields {
		result[key] = value
	}
	return result
}

func safeConsoleLine(level Level, category string, event EventName, message string) string {
	line := fmt.Sprintf("[%s][%s][%s]", category, level, event)
	if message != "" {
		line += " " + message
	}
	return line
}

func (e *Emitter) reject(code RejectionCode, key string) Result {
	e.mu.Lock()
	e.status.RejectedCount++
	e.status.LastRejectionCode = code
	e.warnLocked(code, key)
	e.mu.Unlock()
	return Result{RejectionCode: code, RejectedKey: key}
}

func (e *Emitter) writeFailure(err error) Result {
	e.mu.Lock()
	e.status.WriteFailureCount++
	e.status.LastRejectionCode = RejectionCodeWriteFailed
	e.status.LastWriteError = err.Error()
	e.warnLocked(RejectionCodeWriteFailed, "")
	e.mu.Unlock()
	return Result{RejectionCode: RejectionCodeWriteFailed, WriteFailed: true}
}

func (e *Emitter) warnLocked(code RejectionCode, key string) {
	if e.config.WarningWriter == nil {
		return
	}
	now := e.config.Now()
	if !e.lastWarning.IsZero() && now.Sub(e.lastWarning) < 5*time.Second {
		return
	}
	e.lastWarning = now
	if e.config.Development && key != "" {
		_, _ = fmt.Fprintf(e.config.WarningWriter, "observability event rejected service=%s code=%s key=%s\n", e.service.EmittedName, code, key)
		return
	}
	_, _ = fmt.Fprintf(e.config.WarningWriter, "observability event rejected service=%s code=%s\n", e.service.EmittedName, code)
}

// Status returns a snapshot of emitter-owned counters and last failure state.
func (e *Emitter) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])
	return string(encoded), nil
}
