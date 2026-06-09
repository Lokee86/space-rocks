extends GutTest

const ShellBootFlow := preload("res://scripts/boot/shell_boot_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


class FakeConnectionService:
	extends RefCounted

	var sent_single_player := 0
	var sent_create_room := 0
	var sent_join_room_codes: Array[String] = []

	func send_start_single_player_request() -> void:
		sent_single_player += 1

	func send_create_room_request() -> void:
		sent_create_room += 1

	func send_join_room_request(room_code: String) -> void:
		sent_join_room_codes.append(room_code)


func test_send_pending_single_player_request_consumes_and_sends() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_single_player()
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_SINGLE_PLAYER)
	assert_true(flow.pending_request_is_single_player())
	assert_false(flow.pending_request_is_multiplayer())

	flow.send_pending_boot_request()

	assert_eq(connection.sent_single_player, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_pending_create_room_is_multiplayer_without_consuming() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_create_room()

	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)
	assert_true(flow.pending_request_is_multiplayer())
	assert_false(flow.pending_request_is_single_player())


func test_pending_join_room_is_multiplayer_without_consuming() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_join_room("ABCD")

	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_JOIN_ROOM)
	assert_true(flow.pending_request_is_multiplayer())
	assert_false(flow.pending_request_is_single_player())
