extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
const AsteroidSync := preload("res://scripts/world/asteroid_sync.gd")
const ProjectileSync := preload("res://scripts/world/projectile_sync.gd")


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


func test_dispatcher_routes_asteroids_lifecycle_create_to_asteroid_create_handling() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var asteroid_sync := _new_asteroid_sync()
	add_child_autofree(dispatcher)
	dispatcher.asteroids_lifecycle_received.connect(func(packet: Dictionary) -> void:
		if str(packet.get("action", "")) != "create":
			return
		asteroid_sync.apply_asteroid(
			str(packet[Packets.FIELD_ID]),
			packet,
			Vector2.ZERO,
			Vector2.ZERO
		)
	)

	dispatcher.dispatch({
		Packets.FIELD_TYPE: "asteroids_lifecycle",
		"action": "create",
		Packets.FIELD_ID: "asteroid-1",
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
		Packets.FIELD_ROTATION: 0.25,
		Packets.FIELD_SCALE: 1.0,
		Packets.FIELD_VARIANT: 2,
	})

	assert_true(asteroid_sync.asteroid_nodes.has("asteroid-1"))
	assert_true(asteroid_sync.initialized_asteroids.has("asteroid-1"))
	assert_eq(asteroid_sync.asteroid_variants["asteroid-1"], 2)


func test_dispatcher_routes_asteroids_lifecycle_delete_to_asteroid_despawn_handling() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var asteroid_sync := _new_asteroid_sync()
	add_child_autofree(dispatcher)
	dispatcher.asteroids_lifecycle_received.connect(func(packet: Dictionary) -> void:
		match str(packet.get("action", "")):
			"create":
				asteroid_sync.apply_asteroid(
					str(packet[Packets.FIELD_ID]),
					packet,
					Vector2.ZERO,
					Vector2.ZERO
				)
			"delete":
				asteroid_sync.remove_asteroid(str(packet[Packets.FIELD_ID]))
	)

	dispatcher.dispatch({
		Packets.FIELD_TYPE: "asteroids_lifecycle",
		"action": "create",
		Packets.FIELD_ID: "asteroid-1",
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
		Packets.FIELD_ROTATION: 0.25,
		Packets.FIELD_SCALE: 1.0,
		Packets.FIELD_VARIANT: 2,
	})
	dispatcher.dispatch({
		Packets.FIELD_TYPE: "asteroids_lifecycle",
		"action": "delete",
		Packets.FIELD_ID: "asteroid-1",
	})

	assert_false(asteroid_sync.asteroid_nodes.has("asteroid-1"))
	assert_true(asteroid_sync.is_deleted("asteroid-1"))


func _new_asteroid_sync() -> AsteroidSync:
	var asteroid_sync := AsteroidSync.new()
	var asteroids_layer := Node2D.new()
	add_child_autofree(asteroids_layer)
	asteroid_sync.configure(asteroids_layer)
	return asteroid_sync


func test_dispatcher_routes_bullets_lifecycle_create_to_projectile_create_handling() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var projectile_sync := _new_projectile_sync()
	add_child_autofree(dispatcher)
	dispatcher.bullets_lifecycle_received.connect(func(packet: Dictionary) -> void:
		if str(packet.get("action", "")) != "create":
			return
		projectile_sync.apply_projectile(
			str(packet[Packets.FIELD_ID]),
			packet,
			Vector2.ZERO,
			Vector2.ZERO
		)
	)

	dispatcher.dispatch({
		Packets.FIELD_TYPE: "bullets_lifecycle",
		"action": "create",
		Packets.FIELD_ID: "bullet-1",
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
		Packets.FIELD_ROTATION: 0.25,
		Packets.FIELD_PROJECTILE_TYPE: "torpedo",
	})

	assert_true(projectile_sync.projectile_nodes.has("bullet-1"))
	assert_eq(projectile_sync.projectile_node_types["bullet-1"], "torpedo")
	assert_eq(projectile_sync.projectile_nodes["bullet-1"].name, "Torpedo")


func test_dispatcher_routes_bullets_lifecycle_delete_to_projectile_despawn_handling() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var projectile_sync := _new_projectile_sync()
	add_child_autofree(dispatcher)
	dispatcher.bullets_lifecycle_received.connect(func(packet: Dictionary) -> void:
		match str(packet.get("action", "")):
			"create":
				projectile_sync.apply_projectile(
					str(packet[Packets.FIELD_ID]),
					packet,
					Vector2.ZERO,
					Vector2.ZERO
				)
			"delete":
				projectile_sync.remove_projectile(str(packet[Packets.FIELD_ID]))
	)

	dispatcher.dispatch({
		Packets.FIELD_TYPE: "bullets_lifecycle",
		"action": "create",
		Packets.FIELD_ID: "bullet-1",
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
		Packets.FIELD_ROTATION: 0.25,
		Packets.FIELD_PROJECTILE_TYPE: "torpedo",
	})
	dispatcher.dispatch({
		Packets.FIELD_TYPE: "bullets_lifecycle",
		"action": "delete",
		Packets.FIELD_ID: "bullet-1",
	})

	assert_false(projectile_sync.projectile_nodes.has("bullet-1"))
	assert_true(projectile_sync.is_deleted("bullet-1"))


func test_dispatcher_routes_bullets_lifecycle_preserves_projectile_kind_type() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var projectile_sync := _new_projectile_sync()
	add_child_autofree(dispatcher)
	dispatcher.bullets_lifecycle_received.connect(func(packet: Dictionary) -> void:
		if str(packet.get("action", "")) != "create":
			return
		projectile_sync.apply_projectile(
			str(packet[Packets.FIELD_ID]),
			packet,
			Vector2.ZERO,
			Vector2.ZERO
		)
	)

	dispatcher.dispatch({
		Packets.FIELD_TYPE: "bullets_lifecycle",
		"action": "create",
		Packets.FIELD_ID: "bullet-1",
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
		Packets.FIELD_ROTATION: 0.25,
		Packets.FIELD_PROJECTILE_TYPE: "torpedo",
	})

	assert_eq(projectile_sync.projectile_node_types["bullet-1"], "torpedo")
	assert_eq(projectile_sync.projectile_nodes["bullet-1"].name, "Torpedo")


func _new_projectile_sync() -> ProjectileSync:
	var projectile_sync := ProjectileSync.new()
	var bullets_layer := Node2D.new()
	add_child_autofree(bullets_layer)
	projectile_sync.configure(bullets_layer)
	return projectile_sync



func test_unknown_route_emits_once_without_recording_packet_dictionary() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var unknown_packets: Array = []
	dispatcher.unknown_packet_received.connect(func(packet: Dictionary) -> void:
		unknown_packets.append(packet)
	)
	dispatcher.dispatch({"type": "unknown_packet", "secret_payload": "must_not_be_logged"})

	assert_eq(unknown_packets.size(), 1)
