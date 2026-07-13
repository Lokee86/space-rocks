extends RefCounted
class_name AsteroidSync

const AsteroidSyncState = preload("res://scripts/world/asteroid_sync_state.gd")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")
const AsteroidPresentation = preload("res://scripts/entities/asteroid.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")
const DELETED_ASTEROID_ID_CAP := 2048

var asteroids_layer: Node2D
var asteroid_nodes := {}
var initialized_asteroids := {}
var warned_missing_asteroid_scale := {}
var warned_missing_asteroid_variant := {}
var target_asteroid_positions := {}
var asteroid_server_positions := {}
var asteroid_visual_positions := {}
var asteroid_variants := {}
var deleted_asteroid_ids := {}
var _deleted_asteroid_id_order := []


func configure(layer: Node2D) -> void:
	asteroids_layer = layer


func reset() -> void:
	for asteroid_id in asteroid_nodes.keys():
		var asteroid_node: AsteroidPresentation = asteroid_nodes[asteroid_id]
		asteroid_node.queue_free()

	asteroid_nodes.clear()
	initialized_asteroids.clear()
	warned_missing_asteroid_scale.clear()
	warned_missing_asteroid_variant.clear()
	target_asteroid_positions.clear()
	asteroid_server_positions.clear()
	asteroid_visual_positions.clear()
	asteroid_variants.clear()
	deleted_asteroid_ids.clear()
	_deleted_asteroid_id_order.clear()


func _contract_violation(asteroid_node: Node) -> void:
	var actual_script: Script = asteroid_node.get_script()
	var actual_script_path := ""
	if actual_script is Script:
		actual_script_path = actual_script.resource_path
	ClientLogger.event(
		ClientLogger.CATEGORY_WORLD_SYNC,
		ClientLogger.LEVEL_ERROR,
		"asteroid_presentation_contract_violation",
		"Asteroid scene root does not satisfy its presentation contract",
		{
			"actual_class": asteroid_node.get_class(),
			"actual_script": actual_script_path,
		}
	)


func get_asteroid_node(asteroid_id: String) -> AsteroidPresentation:
	if asteroid_nodes.has(asteroid_id):
		return asteroid_nodes[asteroid_id]

	var scene_root: Node = ASTEROID_SCENE.instantiate()
	var asteroid_node := scene_root as AsteroidPresentation
	if asteroid_node == null:
		_contract_violation(scene_root)
		scene_root.queue_free()
		return null
	asteroids_layer.add_child(asteroid_node)
	asteroid_nodes[asteroid_id] = asteroid_node

	return asteroid_node


func apply_asteroid_scale(asteroid_id: String, asteroid_node: AsteroidPresentation, state: Dictionary) -> void:
	if state.has(Packets.FIELD_SCALE):
		asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SCALE])
		return

	if warned_missing_asteroid_scale.has(asteroid_id):
		return

	warned_missing_asteroid_scale[asteroid_id] = true
	push_warning("Asteroid state missing scale for %s" % asteroid_id)


func apply_asteroid(
	asteroid_id: String,
	state: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2,
	create_if_missing: bool = true
) -> void:
	if create_if_missing:
		_clear_deleted_asteroid_id(asteroid_id)
	elif deleted_asteroid_ids.has(asteroid_id):
		return

	if !create_if_missing and !asteroid_nodes.has(asteroid_id):
		return

	var asteroid_node: AsteroidPresentation = get_asteroid_node(asteroid_id)
	if asteroid_node == null:
		return
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
		if state.has(Packets.FIELD_VARIANT):
			asteroid_node.set_asteroid_variant(state[Packets.FIELD_VARIANT])
		elif !warned_missing_asteroid_variant.has(asteroid_id):
			warned_missing_asteroid_variant[asteroid_id] = true
			push_warning("Asteroid state missing variant for %s" % asteroid_id)


func apply(
	server_asteroids: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for asteroid_id in server_asteroids.keys():
		apply_asteroid(asteroid_id, server_asteroids[asteroid_id], local_visual_position, local_server_position)


func remove_asteroid(asteroid_id: String) -> void:
	if not deleted_asteroid_ids.has(asteroid_id):
		deleted_asteroid_ids[asteroid_id] = true
		_deleted_asteroid_id_order.append(asteroid_id)
		while _deleted_asteroid_id_order.size() > DELETED_ASTEROID_ID_CAP:
			deleted_asteroid_ids.erase(_deleted_asteroid_id_order.pop_front())
	if !asteroid_nodes.has(asteroid_id):
		return

	var asteroid_node: AsteroidPresentation = asteroid_nodes[asteroid_id]
	asteroid_node.queue_free()
	asteroid_nodes.erase(asteroid_id)
	warned_missing_asteroid_scale.erase(asteroid_id)
	warned_missing_asteroid_variant.erase(asteroid_id)
	initialized_asteroids.erase(asteroid_id)
	target_asteroid_positions.erase(asteroid_id)
	asteroid_server_positions.erase(asteroid_id)
	asteroid_visual_positions.erase(asteroid_id)
	asteroid_variants.erase(asteroid_id)

func _clear_deleted_asteroid_id(asteroid_id: String) -> void:
	if not deleted_asteroid_ids.has(asteroid_id):
		return
	deleted_asteroid_ids.erase(asteroid_id)
	_deleted_asteroid_id_order.erase(asteroid_id)


func is_deleted(asteroid_id: String) -> bool:
	return deleted_asteroid_ids.has(asteroid_id)


func remove_missing(server_asteroids: Dictionary) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if server_asteroids.has(asteroid_id):
			continue

		remove_asteroid(asteroid_id)


func interpolate(weight: float) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if !target_asteroid_positions.has(asteroid_id):
			continue

		var asteroid_node: AsteroidPresentation = asteroid_nodes[asteroid_id]
		asteroid_node.global_position = asteroid_node.global_position.lerp(
			target_asteroid_positions[asteroid_id],
			weight
		)


func asteroid_target_positions() -> Dictionary:
	var positions := {}
	for asteroid_id in asteroid_visual_positions.keys():
		if not asteroid_server_positions.has(asteroid_id):
			continue
		var asteroid_node: AsteroidPresentation = asteroid_nodes.get(asteroid_id, null)
		var visual_scale := 1.0
		if asteroid_node != null:
			visual_scale = float(asteroid_node.scale.x)
		positions[asteroid_id] = {
			"visual_position": asteroid_visual_positions[asteroid_id],
			"server_position": asteroid_server_positions[asteroid_id],
			"visual_scale": visual_scale,
		}
	return positions
