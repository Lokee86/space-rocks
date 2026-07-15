extends GutTest

const Emitter := preload("res://scripts/logging/observability_emitter.gd")
const Contract := preload("res://scripts/generated/observability/contract_generated.gd")
const INSTANCE_ID := "550e8400-e29b-41d4-a716-446655440010"
const EVENT_ID := "550e8400-e29b-41d4-a716-446655440011"
const TRACE_ID := "550e8400-e29b-41d4-a716-446655440012"

class Sink extends RefCounted:
	var records: Array[String] = []
	var failure_count := 0
	var last_failure_message := ""

	func write_line(record: String) -> void:
		records.append(record)

class FailingSink extends Sink:
	func write_line(_record: String) -> void:
		failure_count += 1
		last_failure_message = "disk unavailable"


func runtime(writer: RefCounted = Sink.new()) -> RefCounted:
	return Emitter.new(
		writer,
		INSTANCE_ID,
		"test",
		"test-build",
		42,
		func(): return "2026-07-14T01:02:03.000000000Z",
		func(): return EVENT_ID,
		func(_line): pass
	)


func test_emits_only_canonical_envelope() -> void:
	var sink := Sink.new()
	var result: Dictionary = runtime(sink).emit(Contract.EVENT_BUILD_VERSION_LOADED, "loaded", {}, {"mode": "test"})
	assert_true(result["accepted"])
	var record = JSON.parse_string(sink.records[0])
	for key in record:
		assert_true(Contract.ALLOWED_TOP_LEVEL_FIELDS.has(key), "unexpected field: %s" % key)
	for key in Contract.REQUIRED_FIELDS:
		assert_true(record.has(key), "missing required field: %s" % key)
	assert_eq(record["service"], "client")
	assert_eq(record["event_id"], EVENT_ID)


func test_bridge_is_confined_to_dedicated_legacy_path() -> void:
	var emitter = runtime()
	assert_eq(emitter.emit(Contract.EVENT_LOG_MESSAGE)["rejection_code"], Contract.REJECTION_BRIDGE_EVENT_FORBIDDEN)
	assert_true(emitter.emit_legacy("info", "network", "legacy")["accepted"])


func test_redaction_and_rejection_use_generated_policy() -> void:
	var sink := Sink.new()
	var emitter = runtime(sink)
	var redacted: Dictionary = emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"private_profile": "sensitive"})
	assert_true(redacted["accepted"])
	assert_true(redacted["redacted"])
	assert_eq(redacted["record"]["fields"]["private_profile"], Contract.REDACTION_REPLACEMENT_MARKER)
	var before := sink.records.size()
	var rejected: Dictionary = emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"password": "do-not-leak"})
	assert_eq(rejected["rejection_code"], Contract.REJECTION_UNSAFE_FIELD)
	assert_eq(sink.records.size(), before)


func test_stable_rejection_codes_cover_trace_uuid_and_limits() -> void:
	var emitter = runtime()
	assert_eq(emitter.emit(Contract.EVENT_CLIENT_CONNECTED)["rejection_code"], Contract.REJECTION_TRACE_REQUIRED)
	assert_eq(emitter.emit(Contract.EVENT_CLIENT_CONNECTED, "", {"trace_id": "bad"})["rejection_code"], Contract.REJECTION_INVALID_UUID)
	assert_eq(emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"Bad-Key": 1})["rejection_code"], Contract.REJECTION_INVALID_FIELD_KEY)
	assert_eq(emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"items": []})["rejection_code"], Contract.REJECTION_INVALID_FIELD_TYPE)
	assert_eq(emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"item": null})["rejection_code"], Contract.REJECTION_NULL_NOT_ALLOWED)
	assert_eq(emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, {"item": "x".repeat(4097)})["rejection_code"], Contract.REJECTION_STRING_LIMIT_EXCEEDED)
	var fields := {}
	for index in range(33):
		fields["item_%d" % index] = index
	assert_eq(emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED, "", {}, fields)["rejection_code"], Contract.REJECTION_FIELD_LIMIT_EXCEEDED)
	var invalid_identity = Emitter.new(Sink.new(), "not-a-uuid", "test", "test-build", 42)
	assert_eq(invalid_identity.emit(Contract.EVENT_BUILD_VERSION_LOADED)["rejection_code"], Contract.REJECTION_INVALID_UUID)


func test_writer_failure_is_non_raising_and_visible() -> void:
	var emitter = runtime(FailingSink.new())
	var result: Dictionary = emitter.emit(Contract.EVENT_BUILD_VERSION_LOADED)
	assert_true(result["write_failed"])
	assert_eq(result["rejection_code"], Contract.REJECTION_WRITE_FAILED)
	assert_eq(emitter.status()["write_failure_count"], 1)
	assert_string_contains(emitter.status()["last_write_error"], "disk unavailable")


func test_shared_cross_language_fixtures_have_matching_outcomes() -> void:
	var fixture_file := FileAccess.open("res://tests/fixtures/observability_emitter_cases.json", FileAccess.READ)
	assert_ne(fixture_file, null)
	var fixture = JSON.parse_string(fixture_file.get_as_text())
	assert_true(fixture is Dictionary)
	for test_case in fixture["cases"]:
		var emitter = runtime(Sink.new())
		var result: Dictionary
		if test_case["mode"] == "legacy":
			result = emitter.emit_legacy(
				str(test_case.get("level", "info")),
				str(test_case.get("category", "fixture")),
				str(test_case.get("message", "")),
				test_case.get("fields", {})
			)
		else:
			result = emitter.emit(
				str(test_case["event"]),
				str(test_case.get("message", "")),
				test_case.get("context", {}),
				test_case.get("fields", {})
			)
		assert_eq(result["accepted"], test_case["accepted"], test_case["id"])
		if test_case.has("redacted"):
			assert_eq(result["redacted"], test_case["redacted"], test_case["id"])
		if test_case.has("rejection_code"):
			assert_eq(result["rejection_code"], test_case["rejection_code"], test_case["id"])
