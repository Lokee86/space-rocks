extends RefCounted
class_name AsteroidSync

const AsteroidSyncState = preload("res://scripts/world/asteroid_sync_state.gd")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

var asteroids_layer: Node2D
var asteroid_nodes := {}
var initialized_asteroids := {}
var warned_missing_asteroid_scale := {}
var target_asteroid_positions := {}
var asteroid_server_positions := {}
var asteroid_visual_positions := {}
var asteroid_variants := {}


func configure(layer: Node2D) -> void:
	asteroids_layer = layer


func reset() -> void:
	for asteroid_id in asteroid_nodes.keys():
		asteroid_nodes[asteroid_id].queue_free()

	asteroid_nodes.clear()
	initialized_asteroids.clear()
	warned_missing_asteroid_scale.clear()
	target_asteroid_positions.clear()
	asteroid_server_positions.clear()
	asteroid_visual_positions.clear()
	asteroid_variants.clear()


func get_asteroid_node(asteroid_id: String):
	if asteroid_nodes.has(asteroid_id):
		return asteroid_nodes[asteroid_id]

	var asteroid_node = ASTEROID_SCENE.instantiate()
	asteroids_layer.add_child(asteroid_node)
	asteroid_nodes[asteroid_id] = asteroid_node

	return asteroid_node


func apply_asteroid_scale(asteroid_id: String, asteroid_node: Node2D, state: Dictionary) -> void:
	if state.has(Packets.FIELD_SCALE):
		asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SCALE])
		return

	if warned_missing_asteroid_scale.has(asteroid_id):
		return

	warned_missing_asteroid_scale[asteroid_id] = true
	push_warning("Asteroid state missing scale for %s" % asteroid_id)


func apply(
	server_asteroids: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for asteroid_id in server_asteroids.keys():
		var state: Dictionary = server_asteroids[asteroid_id]
		var asteroid_node = get_asteroid_node(asteroid_id)
		var raw_server_position := AsteroidSyncState.server_position(state)
		var visual_position: Vector2

		if asteroid_server_positions.has(asteroid_id):
			visual_position = asteroid_visual_positions[asteroid_id] + WorldWrapScript.shortest_delta(
				asteroid_server_positions[asteroid_id],
				raw_server_position
			)
			target_asteroid_positions[asteroid_id] = visual_position
			asteroid_server_positions[asteroid_id] = raw_server_position
			asteroid_visual_positions[asteroid_id] = visual_position
		else:
			# First-seen asteroid positions may intentionally be outside wrapped world bounds for offscreen spawns.
			visual_position = local_visual_position + WorldWrapScript.shortest_delta(
				local_server_position,
				raw_server_position
			)
			target_asteroid_positions[asteroid_id] = visual_position
			asteroid_server_positions[asteroid_id] = raw_server_position
			asteroid_visual_positions[asteroid_id] = visual_position

		apply_asteroid_scale(asteroid_id, asteroid_node, state)
		if state.has(Packets.FIELD_VARIANT):
			asteroid_variants[asteroid_id] = int(state[Packets.FIELD_VARIANT])

		if !initialized_asteroids.has(asteroid_id):
			initialized_asteroids[asteroid_id] = true
			asteroid_node.global_position = visual_position
			asteroid_node.set_asteroid_variant(state[Packets.FIELD_VARIANT])


func remove_missing(server_asteroids: Dictionary) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if server_asteroids.has(asteroid_id):
			continue

		asteroid_nodes[asteroid_id].queue_free()
		asteroid_nodes.erase(asteroid_id)
		warned_missing_asteroid_scale.erase(asteroid_id)
		initialized_asteroids.erase(asteroid_id)
		target_asteroid_positions.erase(asteroid_id)
		asteroid_server_positions.erase(asteroid_id)
		asteroid_visual_positions.erase(asteroid_id)
		asteroid_variants.erase(asteroid_id)


func interpolate(weight: float) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if !target_asteroid_positions.has(asteroid_id):
			continue

		var asteroid_node = asteroid_nodes[asteroid_id]
		asteroid_node.global_position = asteroid_node.global_position.lerp(
			target_asteroid_positions[asteroid_id],
			weight
		)


func asteroid_target_positions() -> Dictionary:
	var positions := {}
	for asteroid_id in asteroid_visual_positions.keys():
		if not asteroid_server_positions.has(asteroid_id):
			continue
		var asteroid_node = asteroid_nodes.get(asteroid_id, null)
		var visual_scale := 1.0
		if asteroid_node != null:
			visual_scale = float(asteroid_node.scale.x)
		positions[asteroid_id] = {
			"visual_position": asteroid_visual_positions[asteroid_id],
			"server_position": asteroid_server_positions[asteroid_id],
			"visual_scale": visual_scale,
		}
	return positions


func server_hitbox_draw_entries() -> Array:
	var entries: Array = []
	for asteroid_id in asteroid_visual_positions.keys():
		if !asteroid_nodes.has(asteroid_id):
			continue

		var asteroid_node = asteroid_nodes[asteroid_id]
		if asteroid_node == null or !is_instance_valid(asteroid_node):
			continue

		var visual_scale := 1.0
		if asteroid_node != null:
			visual_scale = float(asteroid_node.scale.x)

		var entry := {
			"kind": "asteroid",
			"id": str(asteroid_id),
			"visual_position": asteroid_visual_positions[asteroid_id],
			"rotation": asteroid_node.rotation if asteroid_node != null else 0.0,
			"scale": visual_scale,
			"variant": int(asteroid_variants.get(asteroid_id, 0)),
		}
		entries.append(entry)

	return entries

