extends RefCounted
class_name AsteroidSync

const AsteroidSyncState = preload("res://scripts/world/asteroid_sync_state.gd")
const AsteroidVariantsScript = preload("res://scripts/generated/asteroids/asteroid_variants.gd")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

const ClientLogger = preload("res://scripts/logging/logger.gd")
const ObservabilityContract = preload("res://scripts/generated/observability/contract_generated.gd")
const AsteroidTrace = preload("res://scripts/networking/realtime/asteroid_trace.gd")
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
var _variant_textures: Array[Texture2D] = []
var deleted_asteroid_ids := {}
var _deleted_asteroid_id_order := []
var _measurement_observer: Callable


func configure(layer: Node2D) -> void:
	asteroids_layer = layer
	_warm_variant_textures()


func _warm_variant_textures() -> void:
	if not _variant_textures.is_empty():
		return
	for variant_index in range(AsteroidVariantsScript.count()):
		var texture_path := AsteroidVariantsScript.texture_path_for_index(variant_index)
		if texture_path.is_empty():
			continue
		var texture := load(texture_path) as Texture2D
		if texture != null:
			_variant_textures.append(texture)


func set_measurement_observer(observer: Callable) -> void:
	_measurement_observer = observer


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
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Asteroid scene root does not satisfy its presentation contract",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "asteroid",
			"resource_kind": "scene",
			"failure_mode": "wrong_scene_root",
			"expected_type": "AsteroidPresentation",
			"actual_type": asteroid_node.get_class(),
			"resource_path": actual_script_path,
		}
	)


func get_asteroid_node(asteroid_id: String) -> AsteroidPresentation:
	if asteroid_nodes.has(asteroid_id):
		return asteroid_nodes[asteroid_id]

	if asteroids_layer == null:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
			"Cannot create asteroid without a configured asteroids layer",
			{},
			{
				"subsystem": "world_sync",
				"entity_kind": "asteroid",
				"resource_kind": "presentation_layer",
				"failure_mode": "missing_layer",
				"expected_type": "Node2D",
				"actual_type": "null",
			}
		)
		return null

	var scene_root: Node = ASTEROID_SCENE.instantiate()
	var asteroid_node := scene_root as AsteroidPresentation
	if asteroid_node == null:
		_contract_violation(scene_root)
		scene_root.queue_free()
		return null
	asteroids_layer.add_child(asteroid_node)
	asteroid_nodes[asteroid_id] = asteroid_node
	AsteroidTrace.record_event("presentation_node_created", {
		"asteroid_id": asteroid_id,
		"deleted_tombstone_present": deleted_asteroid_ids.has(asteroid_id),
	})
	_record_lifecycle("created")

	return asteroid_node


func apply_asteroid_scale(asteroid_id: String, asteroid_node: AsteroidPresentation, state: Dictionary) -> void:
	if state.has(Packets.FIELD_SCALE):
		asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SCALE])
		return

	if warned_missing_asteroid_scale.has(asteroid_id):
		return

	warned_missing_asteroid_scale[asteroid_id] = true
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_STATE_INVALID,
		"Asteroid presentation state is missing a required field",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "asteroid",
			"failure_mode": "missing_state_field",
			"field_name": "scale",
		}
	)


func apply_asteroid(
	asteroid_id: String,
	state: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2,
	create_if_missing: bool = true
) -> void:
	if create_if_missing:
		if deleted_asteroid_ids.has(asteroid_id):
			AsteroidTrace.record_event("presentation_tombstone_cleared", {"asteroid_id": asteroid_id})
		_clear_deleted_asteroid_id(asteroid_id)
	elif deleted_asteroid_ids.has(asteroid_id):
		AsteroidTrace.record_event("presentation_update_blocked_by_tombstone", {"asteroid_id": asteroid_id})
		return

	if !create_if_missing and !asteroid_nodes.has(asteroid_id):
		return

	var asteroid_node: AsteroidPresentation = get_asteroid_node(asteroid_id)
	if asteroid_node == null:
		return
	_record_lifecycle("updated")
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
			ClientLogger.emit_canonical(
				ObservabilityContract.EVENT_CLIENT_PRESENTATION_STATE_INVALID,
				"Asteroid presentation state is missing a required field",
				{},
				{
					"subsystem": "world_sync",
					"entity_kind": "asteroid",
					"failure_mode": "missing_state_field",
					"field_name": "variant",
				}
			)


func apply(
	server_asteroids: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for asteroid_id in server_asteroids.keys():
		apply_asteroid(asteroid_id, server_asteroids[asteroid_id], local_visual_position, local_server_position)


func rebase_to_view_anchor(anchor_visual_position: Vector2, anchor_server_position: Vector2) -> void:
	for asteroid_id in asteroid_server_positions.keys():
		if !asteroid_visual_positions.has(asteroid_id):
			continue
		var copy_offset := WorldWrapScript.visual_copy_offset_to_anchor(
			asteroid_visual_positions[asteroid_id],
			anchor_visual_position,
			anchor_server_position,
			asteroid_server_positions[asteroid_id]
		)
		if copy_offset == Vector2.ZERO:
			continue
		asteroid_visual_positions[asteroid_id] += copy_offset
		if target_asteroid_positions.has(asteroid_id):
			target_asteroid_positions[asteroid_id] += copy_offset
		var asteroid_node: AsteroidPresentation = asteroid_nodes.get(asteroid_id, null)
		if asteroid_node != null:
			asteroid_node.global_position += copy_offset


func remove_asteroid(asteroid_id: String) -> void:
	AsteroidTrace.record_event("presentation_remove", {
		"asteroid_id": asteroid_id,
		"node_existed": asteroid_nodes.has(asteroid_id),
		"initialized": initialized_asteroids.has(asteroid_id),
	})
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
	_record_lifecycle("removed")


func _record_lifecycle(operation: String) -> void:
	if _measurement_observer.is_valid():
		_measurement_observer.call("asteroids", operation, 1)

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
