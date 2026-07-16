extends GutTest

const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")


class FixedTraceFactory:
	var ids := [
		"00000000-0000-4000-8000-000000000001",
		"00000000-0000-4000-8000-000000000002",
	]
	var operations: Array[String] = []
	var index := 0

	func create(operation_name: String) -> ClientOperationTrace:
		operations.append(operation_name)
		var trace_id: String = ids[index]
		index += 1
		return ClientOperationTrace.new(operation_name, Callable(self, "_return_id").bind(trace_id))

	func _return_id(trace_id: String) -> String:
		return trace_id


func test_generates_uuid_v4_and_exposes_only_operation_accessors() -> void:
	var trace := ClientOperationTrace.new("discord_sign_in")
	var regex := RegEx.new()
	regex.compile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")

	assert_true(regex.search(trace.trace_id()) != null)
	assert_eq(trace.operation(), "discord_sign_in")
	assert_false(trace.has_method("emit"))
	assert_false(trace.has_method("send_packet"))


func test_injected_uuid_generator_is_deterministic() -> void:
	var trace := ClientOperationTrace.new(
		"saved_token_validation",
		Callable(self, "_fixed_uuid")
	)

	assert_eq(trace.trace_id(), "00000000-0000-4000-8000-000000000099")


func test_factory_creates_independent_traces_without_storing_operation_payloads() -> void:
	var factory := FixedTraceFactory.new()
	var first := ClientOperationTrace.create("create_room", Callable(factory, "create"))
	var second := ClientOperationTrace.create("join_room", Callable(factory, "create"))

	assert_ne(first.trace_id(), second.trace_id())
	assert_eq(factory.operations, ["create_room", "join_room"])
	assert_eq(first.operation(), "create_room")
	assert_eq(second.operation(), "join_room")
	assert_false(first.has_method("room_code"))
	assert_false(first.has_method("local_profile_id"))


func _fixed_uuid() -> String:
	return "00000000-0000-4000-8000-000000000099"