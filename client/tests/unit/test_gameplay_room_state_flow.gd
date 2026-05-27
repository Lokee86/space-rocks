extends GutTest

const GameplayRoomStateFlow := preload("res://scripts/gameplay/session/gameplay_room_state_flow.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_room_state_from_packet_uses_packet_room_state() -> void:
	var data := {
		Packets.FIELD_ROOM_STATE: " GameOver ",
	}

	assert_eq(GameplayRoomStateFlow.room_state_from_packet(data, "InGame"), "GameOver")


func test_room_state_from_packet_uses_fallback_when_missing() -> void:
	assert_eq(GameplayRoomStateFlow.room_state_from_packet({}, "InGame"), "InGame")


func test_should_stop_spectating_for_room_game_over() -> void:
	assert_true(GameplayRoomStateFlow.should_stop_spectating_for_room_state("GameOver"))


func test_should_not_stop_spectating_for_non_game_over_room_state() -> void:
	assert_false(GameplayRoomStateFlow.should_stop_spectating_for_room_state("InGame"))
