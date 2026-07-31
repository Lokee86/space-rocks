extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


func test_packet_builders_set_expected_type() -> void:
	var cases := [
		[Packets.toggle_debug_invincible_packet(), Packets.TYPE_TOGGLE_DEBUG_INVINCIBLE],
		[Packets.toggle_debug_infinite_lives_packet(), Packets.TYPE_TOGGLE_DEBUG_INFINITE_LIVES],
		[Packets.toggle_debug_freeze_world_packet(), Packets.TYPE_TOGGLE_DEBUG_FREEZE_WORLD],
		[Packets.create_room_request_packet("", "", "ffa", "", 0, 0, "arcade_survival", 0, false, 0), Packets.TYPE_CREATE_ROOM_REQUEST],
		[Packets.leave_room_request_packet(), Packets.TYPE_LEAVE_ROOM_REQUEST],
		[Packets.start_game_request_packet(), Packets.TYPE_START_GAME_REQUEST],
		[Packets.return_to_lobby_request_packet(), Packets.TYPE_RETURN_TO_LOBBY_REQUEST],
	]

	for test_case in cases:
		var packet: Variant = test_case[0]

		assert_eq(typeof(packet), TYPE_DICTIONARY)
		assert_eq(packet[Packets.FIELD_TYPE], test_case[1])


func test_required_packet_field_constants_exist() -> void:
	var required_fields := [
		Packets.FIELD_TYPE,
		Packets.FIELD_X,
		Packets.FIELD_Y,
		Packets.FIELD_ID,
		Packets.FIELD_SIZE,
		Packets.FIELD_SHIP_TYPE,
		Packets.FIELD_LIVES,
		Packets.FIELD_RESPAWN_DELAY,
		Packets.FIELD_ROOM_CODE,
		Packets.FIELD_ROOM_STATE,
		Packets.FIELD_MEMBERS,
		Packets.FIELD_READY,
		Packets.FIELD_MAX_PLAYERS,
		Packets.FIELD_ERROR_CODE,
		Packets.FIELD_MESSAGE,
		Packets.FIELD_TRACE_ID,
	]

	for field in required_fields:
		assert_eq(typeof(field), TYPE_STRING)
		assert_false(field.is_empty())


func test_lobby_packet_builders_include_request_fields() -> void:
	var join_packet := Packets.join_room_request_packet("TEST", "")
	assert_eq(join_packet[Packets.FIELD_TYPE], Packets.TYPE_JOIN_ROOM_REQUEST)
	assert_eq(join_packet[Packets.FIELD_ROOM_CODE], "TEST")

	var ready_packet := Packets.set_ready_request_packet(true)
	assert_eq(ready_packet[Packets.FIELD_TYPE], Packets.TYPE_SET_READY_REQUEST)
	assert_eq(ready_packet[Packets.FIELD_READY], true)


func test_pause_request_packet_sets_expected_type_without_paused_field() -> void:
	var packet := Packets.pause_request_packet()

	assert_eq(packet[Packets.FIELD_TYPE], Packets.TYPE_PAUSE_REQUEST)
	assert_false(packet.has(Packets.FIELD_PAUSED))


func test_input_packet_uses_primary_and_secondary_fire_fields() -> void:
	var packet := Packets.input_packet(true, false, true, false, true, false)

	assert_eq(packet[Packets.FIELD_TYPE], Packets.TYPE_INPUT)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_FORWARD], true)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_BACK], false)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_RIGHT], true)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_LEFT], false)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_PRIMARY_FIRE], true)
	assert_eq(packet[Packets.FIELD_INPUT][Packets.FIELD_SECONDARY_FIRE], false)


func test_toggle_debug_freeze_world_target_packet_sets_expected_fields() -> void:
	var packet := Packets.toggle_debug_freeze_world_target_packet("asteroids")

	assert_eq(packet["type"], "toggle_debug_freeze_world")
	assert_eq(packet["freeze_target"], "asteroids")


func test_lobby_packet_type_constants_exist() -> void:
	var packet_types := [
		Packets.TYPE_CREATE_ROOM_REQUEST,
		Packets.TYPE_JOIN_ROOM_REQUEST,
		Packets.TYPE_LEAVE_ROOM_REQUEST,
		Packets.TYPE_SET_READY_REQUEST,
		Packets.TYPE_START_GAME_REQUEST,
		Packets.TYPE_RETURN_TO_LOBBY_REQUEST,
		Packets.TYPE_ROOM_SNAPSHOT,
		Packets.TYPE_ROOM_STATE_CHANGED,
		Packets.TYPE_ROOM_ERROR,
	]

	for packet_type in packet_types:
		assert_eq(typeof(packet_type), TYPE_STRING)
		assert_false(packet_type.is_empty())

func test_initial_room_packet_builders_carry_trace_id() -> void:
	var trace_id := "00000000-0000-4000-8000-000000000021"

	var create_packet := Packets.create_room_request_packet(trace_id, "profile-1", "custom", "owner_assigned", 0, 8, "score_attack", 4, false, 2500)
	assert_eq(create_packet[Packets.FIELD_TRACE_ID], trace_id)
	assert_eq(create_packet[Packets.FIELD_TEAM_STRUCTURE], "custom")
	assert_eq(create_packet[Packets.FIELD_TEAM_ASSIGNMENT_MODE], "owner_assigned")
	assert_eq(create_packet[Packets.FIELD_TEAM_COUNT], 0)
	assert_eq(create_packet[Packets.FIELD_MAX_PLAYERS], 8)
	assert_eq(create_packet[Packets.FIELD_PRESET_ID], "score_attack")
	assert_eq(create_packet[Packets.FIELD_STARTING_LIVES], 4)
	assert_eq(create_packet[Packets.FIELD_INFINITE_LIVES], false)
	assert_eq(create_packet[Packets.FIELD_TARGET_SCORE], 2500)

	var join_packet := Packets.join_room_request_packet("TEST", trace_id)
	assert_eq(join_packet[Packets.FIELD_TRACE_ID], trace_id)

	var single_player_packet := Packets.start_single_player_request_packet("profile-1", trace_id)
	assert_eq(single_player_packet[Packets.FIELD_TRACE_ID], trace_id)

	var empty_trace_packet := Packets.create_room_request_packet("", "", "ffa", "", 0, 0, "arcade_survival", 0, false, 0)
	assert_true(empty_trace_packet.has(Packets.FIELD_TRACE_ID))
	assert_eq(empty_trace_packet[Packets.FIELD_TRACE_ID], "")
