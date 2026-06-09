extends GutTest

const PendingBootRequest := preload("res://scripts/boot/pending_boot_request.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


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
