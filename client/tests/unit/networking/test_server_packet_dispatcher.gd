extends GutTest

const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")


func test_dispatcher_emits_asteroid_delta_received() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var asteroid_packets: Array = []
	add_child_autofree(dispatcher)
	dispatcher.asteroid_delta_received.connect(func(packet: Dictionary) -> void:
		asteroid_packets.append(packet)
	)

	dispatcher.dispatch({"type": "asteroid_delta"})

	assert_eq(asteroid_packets.size(), 1)
	assert_eq(asteroid_packets[0]["type"], "asteroid_delta")


func test_dispatcher_emits_bullet_delta_received_without_world_delta() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var world_packets: Array = []
	var bullet_packets: Array = []
	add_child_autofree(dispatcher)
	dispatcher.world_delta_received.connect(func(packet: Dictionary) -> void:
		world_packets.append(packet)
	)
	dispatcher.bullet_delta_received.connect(func(packet: Dictionary) -> void:
		bullet_packets.append(packet)
	)

	dispatcher.dispatch({"type": "bullet_delta"})

	assert_eq(bullet_packets.size(), 1)
	assert_eq(bullet_packets[0]["type"], "bullet_delta")
	assert_true(world_packets.is_empty())
