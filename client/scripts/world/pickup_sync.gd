extends RefCounted
class_name PickupSync

const PICKUP_CLASS_POWERUP := "powerup"
const PICKUP_CLASS_WEAPON := "weapon"

const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const PickupPresentation = preload("res://scripts/entities/pickup.gd")

var pickups_layer: Node2D
var audio_flow := GameplayAudioFlow.new()
var pickup_nodes = {}
var pickup_types = {}
var pickup_classes = {}
var initialized_pickups = {}
var target_pickup_positions = {}
var pickup_server_positions = {}
var pickup_visual_positions = {}


func configure(layer: Node2D) -> void:
	pickups_layer = layer


func reset() -> void:
	for pickup_id in pickup_nodes.keys():
		var pickup_node: PickupPresentation = pickup_nodes[pickup_id]
		if pickup_node != null:
			if is_instance_valid(pickup_node):
				pickup_node.queue_free()

	pickup_nodes.clear()
	pickup_types.clear()
	pickup_classes.clear()
	initialized_pickups.clear()
	target_pickup_positions.clear()
	pickup_server_positions.clear()
	pickup_visual_positions.clear()


func _scene_for_class(pickup_class: String) -> PackedScene:
	var pickup_scene: PackedScene = PickupPresentationCatalog.scene_for_class(pickup_class)
	if pickup_scene != null:
		return pickup_scene

	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Cannot resolve pickup presentation scene",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "pickup",
			"resource_kind": "scene",
			"failure_mode": "unknown_pickup_class",
			"expected_type": "known_pickup_class",
			"actual_type": pickup_class,
		}
	)
	return null


func _contract_violation(pickup_class: String, pickup_node: Node) -> void:
	var actual_script: Script = pickup_node.get_script()
	var actual_script_path := ""
	if actual_script is Script:
		actual_script_path = actual_script.resource_path
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Pickup scene root does not satisfy its presentation contract",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "pickup",
			"resource_kind": "scene",
			"failure_mode": "wrong_scene_root",
			"expected_type": "PickupPresentation",
			"actual_type": pickup_node.get_class(),
			"resource_path": actual_script_path,
		}
	)


func get_pickup_node(pickup_id: String, pickup_type: String, pickup_class: String) -> PickupPresentation:
	if pickup_nodes.has(pickup_id):
		var stored_value: Variant = pickup_nodes[pickup_id]
		var stored_node := stored_value as Node
		if stored_node == null:
			ClientLogger.emit_canonical(
				ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
				"Stored pickup value is not a presentation node",
				{},
				{
					"subsystem": "world_sync",
					"entity_kind": "pickup",
					"resource_kind": "stored_presentation",
					"failure_mode": "wrong_node_type",
					"expected_type": "PickupPresentation",
					"actual_type": str(typeof(stored_value)),
				}
			)
			return null
		var pickup_node := stored_node as PickupPresentation
		if pickup_node == null:
			_contract_violation(pickup_class, stored_node)
			return null
		return pickup_node

	if pickups_layer == null:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
			"Cannot create pickup without a configured pickups layer",
			{},
			{
				"subsystem": "world_sync",
				"entity_kind": "pickup",
				"resource_kind": "presentation_layer",
				"failure_mode": "missing_layer",
				"expected_type": "Node2D",
				"actual_type": "null",
			}
		)
		return null

	var pickup_scene = _scene_for_class(pickup_class)
	if pickup_scene == null:
		return null

	var scene_root: Node = pickup_scene.instantiate()
	var pickup_node := scene_root as PickupPresentation
	if pickup_node == null:
		_contract_violation(pickup_class, scene_root)
		scene_root.queue_free()
		return null
	pickups_layer.add_child(pickup_node)
	pickup_node.play_spawn_sound(audio_flow)
	pickup_node.apply_pickup_presentation(pickup_type)
	pickup_nodes[pickup_id] = pickup_node
	pickup_types[pickup_id] = pickup_type
	pickup_classes[pickup_id] = pickup_class

	return pickup_node


func apply(server_pickups: Dictionary, local_visual_position: Vector2, local_server_position: Vector2) -> void:
	for pickup_id in server_pickups.keys():
		var state = server_pickups[pickup_id]
		if not state is Dictionary:
			continue

		var resolved_pickup_id = str(pickup_id)
		var pickup_type = PickupSyncState.pickup_type(state)
		var pickup_class = PickupSyncState.pickup_class(state)
		var pickup_node: PickupPresentation = get_pickup_node(resolved_pickup_id, pickup_type, pickup_class)
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
		pickup_classes[resolved_pickup_id] = pickup_class

		if not initialized_pickups.has(resolved_pickup_id):
			initialized_pickups[resolved_pickup_id] = true
			pickup_node.global_position = visual_position

		pickup_node.apply_lifespan_state(age_seconds, lifespan_seconds)


func remove_missing(server_pickups: Dictionary) -> void:
	var stale_pickup_ids = []

	for pickup_id in pickup_nodes.keys():
		if not server_pickups.has(pickup_id):
			stale_pickup_ids.append(pickup_id)

	for pickup_id in stale_pickup_ids:
		var pickup_node: PickupPresentation = pickup_nodes[pickup_id]
		if pickup_node != null:
			if is_instance_valid(pickup_node):
				pickup_node.queue_free()

		pickup_nodes.erase(pickup_id)
		pickup_types.erase(pickup_id)
		pickup_classes.erase(pickup_id)
		initialized_pickups.erase(pickup_id)
		target_pickup_positions.erase(pickup_id)
		pickup_server_positions.erase(pickup_id)
		pickup_visual_positions.erase(pickup_id)


func pickup_position_entries() -> Dictionary:
	var positions = {}

	for pickup_id in target_pickup_positions.keys():
		var visual_position = target_pickup_positions[pickup_id]

		positions[pickup_id] = {
			"visual_position": visual_position,
			"server_position": pickup_server_positions.get(pickup_id, visual_position),
			"pickup_type": pickup_types.get(pickup_id, ""),
			"pickup_class": pickup_classes.get(pickup_id, ""),
			"node": pickup_nodes.get(pickup_id, null),
		}

	return positions


func interpolate(weight: float) -> void:
	for pickup_id in pickup_nodes.keys():
		if not target_pickup_positions.has(pickup_id):
			continue

		var pickup_node: PickupPresentation = pickup_nodes[pickup_id]
		if pickup_node == null:
			continue
		if not is_instance_valid(pickup_node):
			continue

		var target_position = target_pickup_positions[pickup_id]
		pickup_node.global_position = pickup_node.global_position.lerp(target_position, weight)
		pickup_visual_positions[pickup_id] = pickup_node.global_position
