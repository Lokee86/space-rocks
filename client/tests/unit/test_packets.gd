extends GutTest

const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_packet_builders_set_expected_type() -> void:
	var cases := [
		[Packets.pause_player_packet(), Packets.TYPE_PAUSE_PLAYER],
		[Packets.resume_player_packet(), Packets.TYPE_RESUME_PLAYER],
		[Packets.toggle_debug_invincible_packet(), Packets.TYPE_TOGGLE_DEBUG_INVINCIBLE],
		[Packets.toggle_debug_infinite_lives_packet(), Packets.TYPE_TOGGLE_DEBUG_INFINITE_LIVES],
		[Packets.toggle_debug_freeze_world_packet(), Packets.TYPE_TOGGLE_DEBUG_FREEZE_WORLD],
		[Packets.create_room_request_packet(), Packets.TYPE_CREATE_ROOM_REQUEST],
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
		Packets.FIELD_MEMBER_ID,
		Packets.FIELD_LOCAL_MEMBER_ID,
		Packets.FIELD_READY,
		Packets.FIELD_MAX_PLAYERS,
		Packets.FIELD_ERROR_CODE,
		Packets.FIELD_MESSAGE,
	]

	for field in required_fields:
		assert_eq(typeof(field), TYPE_STRING)
		assert_false(field.is_empty())


func test_lobby_packet_builders_include_request_fields() -> void:
	var join_packet := Packets.join_room_request_packet("TEST")
	assert_eq(join_packet[Packets.FIELD_TYPE], Packets.TYPE_JOIN_ROOM_REQUEST)
	assert_eq(join_packet[Packets.FIELD_ROOM_CODE], "TEST")

	var ready_packet := Packets.set_ready_request_packet(true)
	assert_eq(ready_packet[Packets.FIELD_TYPE], Packets.TYPE_SET_READY_REQUEST)
	assert_eq(ready_packet[Packets.FIELD_READY], true)


func test_pause_request_packet_sets_expected_type_without_paused_field() -> void:
	var packet := Packets.pause_request_packet()

	assert_eq(packet[Packets.FIELD_TYPE], Packets.TYPE_PAUSE_REQUEST)
	assert_false(packet.has(Packets.FIELD_PAUSED))


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
