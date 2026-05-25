extends GutTest

const Packets := preload("res://scripts/networking/packets.gd")
const BulletSyncState := preload("res://scripts/networking/bullet_sync_state.gd")


func test_server_position_reads_packet_coordinates() -> void:
	assert_eq(
		BulletSyncState.server_position({
			Packets.FIELD_X: 420.5,
			Packets.FIELD_Y: 840.25,
		}),
		Vector2(420.5, 840.25)
	)
