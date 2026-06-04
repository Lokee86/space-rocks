extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const AsteroidSyncState := preload("res://scripts/world/asteroid_sync_state.gd")


func test_server_position_reads_packet_coordinates() -> void:
	assert_eq(
		AsteroidSyncState.server_position({
			Packets.FIELD_X: 320.5,
			Packets.FIELD_Y: 640.25,
		}),
		Vector2(320.5, 640.25)
	)

