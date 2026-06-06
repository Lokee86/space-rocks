extends RefCounted
class_name PickupSync

const PICKUP_ONE_UP_SCENE = preload("res://scenes/pickups/1_up.tscn")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

var pickups_layer = null
var audio_flow := GameplayAudioFlow.new()
var pickup_nodes = {}
var pickup_types = {}
var initialized_pickups = {}
var target_pickup_positions = {}
var pickup_server_positions = {}
var pickup_visual_positions = {}


func configure(layer: Node2D) -> void:
	pickups_layer = layer


func reset() -> void:
	for pickup_id in pickup_nodes.keys():
		var pickup_node = pickup_nodes[pickup_id]
		if pickup_node != null:
			if is_instance_valid(pickup_node):
				pickup_node.queue_free()

	pickup_nodes.clear()
	pickup_types.clear()
	initialized_pickups.clear()
	target_pickup_positions.clear()
	pickup_server_positions.clear()
	pickup_visual_positions.clear()


func _scene_for_type(pickup_type: String):
	if pickup_type == "1_up":
		return PICKUP_ONE_UP_SCENE

	print("PickupSync: unknown pickup type=%s" % pickup_type)
	return null


func get_pickup_node(pickup_id: String, pickup_type: String):
	if pickup_nodes.has(pickup_id):
		return pickup_nodes[pickup_id]

	if pickups_layer == null:
		print("PickupSync: cannot create pickup; pickups_layer is null")
		return null

	var pickup_scene = _scene_for_type(pickup_type)
	if pickup_scene == null:
		return null

	var pickup_node = pickup_scene.instantiate()
	pickups_layer.add_child(pickup_node)
	if pickup_node.has_method("play_spawn_sound"):
		pickup_node.play_spawn_sound(audio_flow)
	pickup_nodes[pickup_id] = pickup_node
	pickup_types[pickup_id] = pickup_type

	return pickup_node


func apply(server_pickups: Dictionary, local_visual_position: Vector2, local_server_position: Vector2) -> void:
	for pickup_id in server_pickups.keys():
		var state = server_pickups[pickup_id]
		if not state is Dictionary:
			continue

		var resolved_pickup_id = str(pickup_id)
		var pickup_type = PickupSyncState.pickup_type(state)
		var pickup_node = get_pickup_node(resolved_pickup_id, pickup_type)
		if pickup_node == null:
			continue

		var age_seconds = PickupSyncState.age_seconds(state)
		var lifespan_seconds = PickupSyncState.lifespan_seconds(state)

		var raw_server_position = PickupSyncState.server_position(state)
		var visual_position = Vector2.ZERO

		if pickup_server_positions.has(resolved_pickup_id):
			var previous_visual_position = pickup_visual_positions[resolved_pickup_id]
			var previous_server_position = pickup_server_positions[resolved_pickup_id]
			var server_delta = WorldWrapScript.shortest_delta(previous_server_position, raw_server_position)
			visual_position = previous_visual_position + server_delta
		else:
			var spawn_delta = WorldWrapScript.shortest_delta(local_server_position, raw_server_position)
			visual_position = local_visual_position + spawn_delta

		target_pickup_positions[resolved_pickup_id] = visual_position
		pickup_server_positions[resolved_pickup_id] = raw_server_position
		pickup_visual_positions[resolved_pickup_id] = visual_position
		pickup_types[resolved_pickup_id] = pickup_type

		if not initialized_pickups.has(resolved_pickup_id):
			initialized_pickups[resolved_pickup_id] = true
			pickup_node.global_position = visual_position

		if pickup_node.has_method("apply_lifespan_state"):
			pickup_node.apply_lifespan_state(age_seconds, lifespan_seconds)


func remove_missing(server_pickups: Dictionary) -> void:
	var stale_pickup_ids = []

	for pickup_id in pickup_nodes.keys():
		if not server_pickups.has(pickup_id):
			stale_pickup_ids.append(pickup_id)

	for pickup_id in stale_pickup_ids:
		var pickup_node = pickup_nodes[pickup_id]
		if pickup_node != null:
			if is_instance_valid(pickup_node):
				pickup_node.queue_free()

		pickup_nodes.erase(pickup_id)
		pickup_types.erase(pickup_id)
		initialized_pickups.erase(pickup_id)
		target_pickup_positions.erase(pickup_id)
		pickup_server_positions.erase(pickup_id)
		pickup_visual_positions.erase(pickup_id)


func pickup_target_positions() -> Dictionary:
	var positions = {}

	for pickup_id in target_pickup_positions.keys():
		var visual_position = target_pickup_positions[pickup_id]
		var server_position = visual_position
		if pickup_server_positions.has(pickup_id):
			server_position = pickup_server_positions[pickup_id]

		positions[pickup_id] = {
			"visual_position": visual_position,
			"server_position": server_position,
		}

	return positions


func interpolate(weight: float) -> void:
	for pickup_id in pickup_nodes.keys():
		if not target_pickup_positions.has(pickup_id):
			continue

		var pickup_node = pickup_nodes[pickup_id]
		if pickup_node == null:
			continue
		if not is_instance_valid(pickup_node):
			continue

		var target_position = target_pickup_positions[pickup_id]
		pickup_node.global_position = pickup_node.global_position.lerp(target_position, weight)
		pickup_visual_positions[pickup_id] = pickup_node.global_position
