extends RefCounted

const Contract := preload("res://scripts/generated/observability/contract_generated.gd")

const CONTEXT_FIELDS := [
	"trace_id", "session_id", "room_id", "player_id", "account_id", "match_id",
	"request_id", "diagnostic_report_id", "audit_event_id", "route", "packet_type", "duration_ms",
]
const UUID_CONTEXT_FIELDS := ["trace_id", "diagnostic_report_id", "audit_event_id"]

var _writer
var _service_key := Contract.SERVICE_CLIENT
var _service: Dictionary
var _service_instance_id := ""
var _environment := "development"
var _build_version := "development"
var _pid := 0
var _uuid_regex := RegEx.new()
var _key_regex := RegEx.new()
var _clock: Callable
var _uuid_generator: Callable
var _warning: Callable
var _last_warning_unix := -10.0
var _status := {
	"accepted_count": 0,
	"rejected_count": 0,
	"redacted_count": 0,
	"write_failure_count": 0,
	"last_rejection_code": "",
	"last_write_error": "",
}


func _init(
	writer = null,
	service_instance_id: String = "",
	environment: String = "development",
	build_version: String = "development",
	pid: int = 0,
	clock: Callable = Callable(),
	uuid_generator: Callable = Callable(),
	warning: Callable = Callable()
) -> void:
	_writer = writer
	_service = Contract.SERVICE_DEFINITIONS[Contract.SERVICE_CLIENT]
	_service_instance_id = service_instance_id if service_instance_id != "" else _new_uuid()
	_environment = environment
	_build_version = build_version
	_pid = pid
	_clock = clock
	_uuid_generator = uuid_generator
	_warning = warning
	_uuid_regex.compile(Contract.UUID_REGEX)
	_key_regex.compile(Contract.FREE_FORM_KEY_REGEX)


func emit(event_name: String, message: String = "", context: Dictionary = {}, fields: Dictionary = {}) -> Dictionary:
	if !Contract.EVENT_DEFINITIONS.has(event_name):
		return _reject(Contract.REJECTION_UNKNOWN_EVENT)
	var definition: Dictionary = Contract.EVENT_DEFINITIONS[event_name]
	if bool(definition.get("bridge_only", false)):
		return _reject(Contract.REJECTION_BRIDGE_EVENT_FORBIDDEN)
	return _emit_definition(definition, str(definition["default_level"]), str(definition["category"]), message, context, fields, true)


func emit_legacy(level: String, category: String, message: String = "", fields: Dictionary = {}, legacy_event: String = "") -> Dictionary:
	return build_legacy(level, category, message, fields, legacy_event, true)


func build_legacy(level: String, category: String, message: String = "", fields: Dictionary = {}, legacy_event: String = "", write_record := false) -> Dictionary:
	if !Contract.CANONICAL_LEVELS.has(level):
		return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, "level")
	if category.is_empty():
		return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, "category")
	var safe_fields := fields.duplicate(true)
	if legacy_event != "" and legacy_event != Contract.EVENT_LOG_MESSAGE:
		safe_fields["legacy_event"] = legacy_event
	return _emit_definition(Contract.EVENT_DEFINITIONS[Contract.EVENT_LOG_MESSAGE], level, category, message, {}, safe_fields, write_record)


func status() -> Dictionary:
	return _status.duplicate(true)


func _emit_definition(definition: Dictionary, level: String, category: String, message: String, context: Dictionary, fields: Dictionary, write_record: bool) -> Dictionary:
	if !definition["services"].has(_service_key):
		return _reject(Contract.REJECTION_SERVICE_NOT_ALLOWED)
	var context_result := _normalize_context(context)
	if !context_result["accepted"]:
		return context_result
	var normalized_context: Dictionary = context_result["value"]
	if bool(definition["trace_required"]) and str(normalized_context.get("trace_id", "")).is_empty():
		return _reject(Contract.REJECTION_TRACE_REQUIRED, "trace_id")
	var fields_result := _sanitize_fields(fields)
	if !fields_result["accepted"]:
		return fields_result
	if message.to_utf8_buffer().size() > Contract.MAX_STRING_BYTES or category.to_utf8_buffer().size() > Contract.MAX_STRING_BYTES:
		return _reject(Contract.REJECTION_STRING_LIMIT_EXCEEDED, "message")
	if !_valid_uuid(_service_instance_id):
		return _reject(Contract.REJECTION_INVALID_UUID, "service_instance_id")

	var event_id := _next_uuid()
	if !_valid_uuid(event_id):
		return _reject(Contract.REJECTION_INVALID_UUID, "event_id")
	var record := {
		"timestamp": _timestamp(),
		"level": level,
		"event": str(definition["name"]),
		"event_id": event_id,
		"service": str(_service["emitted_name"]),
		"environment": _environment,
		"build_version": _build_version,
		"schema_version": Contract.SCHEMA_VERSION,
		"service_instance_id": _service_instance_id,
		"category": category,
		"retention_tier": str(definition["retention_tier"]),
	}
	if !message.is_empty():
		record["message"] = message
	for key in normalized_context:
		record[key] = normalized_context[key]
	if _pid > 0:
		record["pid"] = _pid
	var safe_fields: Dictionary = fields_result["value"]
	if !safe_fields.is_empty():
		record["fields"] = safe_fields
	if record.size() > Contract.MAX_EVENT_FIELDS:
		return _reject(Contract.REJECTION_FIELD_LIMIT_EXCEEDED)
	var payload := JSON.stringify(record)
	if payload.is_empty():
		return _reject(Contract.REJECTION_SERIALIZATION_FAILED)
	if payload.to_utf8_buffer().size() > Contract.MAX_EVENT_BYTES:
		return _reject(Contract.REJECTION_EVENT_TOO_LARGE)

	if write_record and _writer != null:
		var failures_before := int(_writer.get("failure_count")) if _has_property(_writer, "failure_count") else 0
		_writer.write_line(payload)
		var failures_after := int(_writer.get("failure_count")) if _has_property(_writer, "failure_count") else failures_before
		if failures_after > failures_before:
			var last_error := str(_writer.get("last_failure_message")) if _has_property(_writer, "last_failure_message") else "writer degraded"
			return _write_failure(last_error)

	_status["accepted_count"] += 1
	if bool(fields_result["redacted"]):
		_status["redacted_count"] += 1
	return {
		"accepted": true,
		"redacted": bool(fields_result["redacted"]),
		"rejection_code": "",
		"rejected_key": "",
		"write_failed": false,
		"record": record,
		"json": payload,
	}


func _normalize_context(context: Dictionary) -> Dictionary:
	var normalized := {}
	for raw_key in context:
		var key := str(raw_key)
		if !CONTEXT_FIELDS.has(key):
			return _reject(Contract.REJECTION_UNKNOWN_CONTEXT_FIELD, key)
		var value = context[raw_key]
		if value == null:
			return _reject(Contract.REJECTION_NULL_NOT_ALLOWED, key)
		if key == "duration_ms":
			if typeof(value) != TYPE_INT and typeof(value) != TYPE_FLOAT:
				return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, key)
			if typeof(value) == TYPE_FLOAT and (is_nan(value) or is_inf(value)):
				return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, key)
		else:
			if typeof(value) != TYPE_STRING:
				return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, key)
			if UUID_CONTEXT_FIELDS.has(key) and !str(value).is_empty() and !_valid_uuid(str(value)):
				return _reject(Contract.REJECTION_INVALID_UUID, key)
			if str(value).to_utf8_buffer().size() > Contract.MAX_STRING_BYTES:
				return _reject(Contract.REJECTION_STRING_LIMIT_EXCEEDED, key)
		normalized[key] = value
	return {"accepted": true, "value": normalized}


func _sanitize_fields(fields: Dictionary) -> Dictionary:
	if fields.size() > Contract.MAX_FREE_FORM_FIELDS:
		return _reject(Contract.REJECTION_FIELD_LIMIT_EXCEEDED)
	var safe := {}
	var redacted := false
	for raw_key in fields:
		var key := str(raw_key)
		var redaction := _redaction_action(key)
		if bool(redaction["ambiguous"]) or (bool(redaction["matched"]) and redaction["action"] == "reject"):
			return _reject(Contract.REJECTION_UNSAFE_FIELD, key)
		if bool(redaction["matched"]) and redaction["action"] == "redact":
			safe[key] = Contract.REDACTION_REPLACEMENT_MARKER
			redacted = true
			continue
		if _key_regex.search(key) == null:
			return _reject(Contract.REJECTION_INVALID_FIELD_KEY, key)
		var value = fields[raw_key]
		if value == null:
			return _reject(Contract.REJECTION_NULL_NOT_ALLOWED, key)
		if typeof(value) not in [TYPE_STRING, TYPE_BOOL, TYPE_INT, TYPE_FLOAT]:
			return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, key)
		if typeof(value) == TYPE_FLOAT and (is_nan(value) or is_inf(value)):
			return _reject(Contract.REJECTION_INVALID_FIELD_TYPE, key)
		if typeof(value) == TYPE_STRING and str(value).to_utf8_buffer().size() > Contract.MAX_FREE_FORM_VALUE_BYTES:
			return _reject(Contract.REJECTION_STRING_LIMIT_EXCEEDED, key)
		safe[key] = value
	return {"accepted": true, "value": safe, "redacted": redacted}


func _redaction_action(key: String) -> Dictionary:
	var candidate := key if Contract.REDACTION_CASE_SENSITIVE else key.to_lower()
	var actions := []
	for rule in Contract.REDACTION_EXACT_RULES:
		if rule["matches"].has(candidate) and !actions.has(rule["action"]):
			actions.append(rule["action"])
	for rule in Contract.REDACTION_FRAGMENT_RULES:
		for fragment in rule["matches"]:
			if candidate.contains(str(fragment)) and !actions.has(rule["action"]):
				actions.append(rule["action"])
	if actions.is_empty():
		return {"matched": false, "ambiguous": false, "action": ""}
	if actions.size() > 1:
		return {"matched": true, "ambiguous": true, "action": Contract.REDACTION_AMBIGUOUS_MATCH_ACTION}
	return {"matched": true, "ambiguous": false, "action": actions[0]}


func _reject(code: String, key: String = "") -> Dictionary:
	_status["rejected_count"] += 1
	_status["last_rejection_code"] = code
	_warn_rejection(code, key)
	return {"accepted": false, "redacted": false, "rejection_code": code, "rejected_key": key, "write_failed": false}


func _write_failure(error: String) -> Dictionary:
	_status["write_failure_count"] += 1
	_status["last_rejection_code"] = Contract.REJECTION_WRITE_FAILED
	_status["last_write_error"] = error
	_warn_rejection(Contract.REJECTION_WRITE_FAILED, "")
	return {"accepted": false, "redacted": false, "rejection_code": Contract.REJECTION_WRITE_FAILED, "rejected_key": "", "write_failed": true}


func _warn_rejection(code: String, _key: String) -> void:
	var now := Time.get_unix_time_from_system()
	if now - _last_warning_unix < 5.0:
		return
	_last_warning_unix = now
	var line := "observability event rejected service=%s code=%s" % [_service["emitted_name"], code]
	if _warning.is_valid():
		_warning.call(line)
	else:
		push_warning(line)


func _timestamp() -> String:
	if _clock.is_valid():
		return str(_clock.call())
	var unix := Time.get_unix_time_from_system()
	var seconds := int(floor(unix))
	var millis := int(floor((unix - seconds) * 1000.0))
	return "%s.%03d000000Z" % [Time.get_datetime_string_from_unix_time(seconds, true), millis]


func _next_uuid() -> String:
	if _uuid_generator.is_valid():
		return str(_uuid_generator.call())
	return _new_uuid()


func _new_uuid() -> String:
	var bytes := Crypto.new().generate_random_bytes(16)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	var hex := bytes.hex_encode()
	return "%s-%s-%s-%s-%s" % [hex.substr(0, 8), hex.substr(8, 4), hex.substr(12, 4), hex.substr(16, 4), hex.substr(20, 12)]


func _valid_uuid(value: String) -> bool:
	return _uuid_regex.search(value) != null


func _has_property(object: Object, property_name: String) -> bool:
	for property in object.get_property_list():
		if property["name"] == property_name:
			return true
	return false
