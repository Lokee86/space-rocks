extends GutTest

const Packets := preload("res://scripts/networking/packets.gd")
const PlayerSyncState := preload("res://scripts/networking/player_sync_state.gd")


func test_server_position_reads_packet_coordinates() -> void:
	assert_eq(
		PlayerSyncState.server_position({
			Packets.FIELD_X: 12.5,
			Packets.FIELD_Y: 34.0,
		}),
		Vector2(12.5, 34.0)
	)


func test_server_rotation_reads_packet_rotation() -> void:
	assert_eq(
		PlayerSyncState.server_rotation({
			Packets.FIELD_ROTATION: 1.25,
		}),
		1.25
	)


func test_is_paused_defaults_to_false_when_missing() -> void:
	assert_false(PlayerSyncState.is_paused({}))


func test_is_paused_coerces_packet_value_to_bool() -> void:
	assert_true(PlayerSyncState.is_paused({Packets.FIELD_PAUSED: 1}))
	assert_false(PlayerSyncState.is_paused({Packets.FIELD_PAUSED: 0}))
