extends GutTest

const Packets := preload("res://scripts/networking/packets.gd")


func test_packet_builders_set_expected_type() -> void:
	var cases := [
		[Packets.pause_player_packet(), Packets.TYPE_PAUSE_PLAYER],
		[Packets.resume_player_packet(), Packets.TYPE_RESUME_PLAYER],
		[Packets.toggle_debug_invincible_packet(), Packets.TYPE_TOGGLE_DEBUG_INVINCIBLE],
		[Packets.toggle_debug_infinite_lives_packet(), Packets.TYPE_TOGGLE_DEBUG_INFINITE_LIVES],
		[Packets.toggle_debug_freeze_world_packet(), Packets.TYPE_TOGGLE_DEBUG_FREEZE_WORLD],
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
	]

	for field in required_fields:
		assert_eq(typeof(field), TYPE_STRING)
		assert_false(field.is_empty())
