extends GutTest

const GameplayDevtoolsContext := preload("res://scripts/devtools/gameplay_devtools_context.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const PresentationEventCapture := preload("res://tests/unit/logging/presentation_event_capture.gd")


class FakeConnectionService:
	func send_packet(_packet: Dictionary, _trace_id: String = "") -> void:
		pass


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_emits_devtools_enabled_after_successful_configuration() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(FakeConnectionService.new())

	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "devtools_enabled")


func test_does_not_emit_devtools_enabled_when_configuration_is_incomplete() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(null)

	assert_eq(writer.written_lines.size(), 0)


func test_emits_devtools_enabled_exactly_once_across_repeated_configuration() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var context := GameplayDevtoolsContext.new()

	context.configure(FakeConnectionService.new())
	context.configure(FakeConnectionService.new())

	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "devtools_enabled")
