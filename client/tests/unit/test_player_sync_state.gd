extends GutTest

const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_server_position_reads_packet_coordinates() -> void:
	assert_eq(
		Vector2(12.5, 34.0),
		Vector2(12.5, 34.0)
	)


func test_server_rotation_reads_packet_rotation() -> void:
	assert_eq(
		1.25,
		1.25
	)
