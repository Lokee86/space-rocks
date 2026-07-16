extends GutTest

const PendingBootRequest := preload("res://scripts/boot/pending_boot_request.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")
class TraceFactory:
	var ids: Array
	var index := 0

	func _init(ids_ref: Array) -> void:
		ids = ids_ref

	func create(operation_name: String) -> ClientOperationTrace:
		var trace_id: String = ids[index]
		index += 1
		return ClientOperationTrace.new(operation_name, Callable(self, "_return_id").bind(trace_id))

	func _return_id(trace_id: String) -> String:
		return trace_id


func test_create_room_request_is_multiplayer_without_consuming() -> void:
	var pending := PendingBootRequest.new()
	pending.request_create_room()

	assert_eq(pending.current_type(), Constants.BOOT_REQUEST_CREATE_ROOM)
	assert_true(pending.is_multiplayer_request())
	assert_false(pending.is_single_player_request())
	assert_true(pending.has_request())


func test_join_room_request_is_multiplayer_without_consuming() -> void:
	var pending := PendingBootRequest.new()
	pending.request_join_room("ABCD")

	assert_eq(pending.current_type(), Constants.BOOT_REQUEST_JOIN_ROOM)
	assert_true(pending.is_multiplayer_request())
	assert_false(pending.is_single_player_request())
	assert_true(pending.has_request())


func test_single_player_request_is_single_player_without_consuming() -> void:
	var pending := PendingBootRequest.new()
	pending.request_single_player()

	assert_eq(pending.current_type(), Constants.BOOT_REQUEST_SINGLE_PLAYER)
	assert_true(pending.is_single_player_request())
	assert_false(pending.is_multiplayer_request())
	assert_true(pending.has_request())

func test_consume_returns_trace_and_clear_removes_pending_context() -> void:
	var factory := Callable(TraceFactory.new([
		"00000000-0000-4000-8000-000000000001",
	]), "create")
	var pending := PendingBootRequest.new(factory)
	pending.request_join_room("ABCD")

	var trace_id := pending.current_trace_id()
	var request := pending.consume_request()

	assert_eq(request["trace_id"], trace_id)
	assert_ne(trace_id, "")
	assert_eq(request["room_code"], "ABCD")
	assert_eq(pending.current_trace_id(), "")
	assert_false(pending.has_request())


func test_replacing_pending_request_replaces_trace() -> void:
	var factory := Callable(TraceFactory.new([
		"00000000-0000-4000-8000-000000000001",
		"00000000-0000-4000-8000-000000000002",
	]), "create")
	var pending := PendingBootRequest.new(factory)
	pending.request_create_room()
	var first_trace_id := pending.current_trace_id()

	pending.request_single_player("local-profile")
	var second_trace_id := pending.current_trace_id()

	assert_ne(first_trace_id, second_trace_id)
	assert_eq(pending.consume_request()["trace_id"], second_trace_id)
