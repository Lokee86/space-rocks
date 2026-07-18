extends RefCounted
class_name ProjectileSync

const ProjectileSceneResolver = preload("res://scripts/world/projectiles/projectile_scene_resolver.gd")
const ProjectileSyncState = preload("res://scripts/world/projectile_sync_state.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")
const WorldSyncLogger = preload("res://scripts/logging/logger.gd")
const ObservabilityContract = preload("res://scripts/generated/observability/contract_generated.gd")
const BulletPresentation = preload("res://scripts/entities/bullet.gd")
const TorpedoPresentation = preload("res://scripts/entities/torpedo.gd")
const DELETED_PROJECTILE_ID_CAP := 4096

var audio_flow := GameplayAudioFlow.new()
var bullets_layer: Node2D
var projectile_nodes := {}
var initialized_projectiles := {}
var pooled_projectile_nodes_by_type := {}
var projectile_node_types := {}
var created_projectile_node_count := 0
var reused_projectile_node_count := 0
var released_projectile_node_count := 0
var target_projectile_positions := {}
var target_projectile_rotations := {}
var deleted_projectile_ids := {}
var _deleted_projectile_id_order := []
var _measurement_observer: Callable


func configure(layer: Node2D) -> void:
	bullets_layer = layer


func set_measurement_observer(observer: Callable) -> void:
	_measurement_observer = observer


func has_projectile(bullet_id: String) -> bool:
	return projectile_nodes.has(bullet_id)


func pool_size() -> int:
	var total := 0
	for pool in pooled_projectile_nodes_by_type.values():
		total += pool.size()
	return total

func pool_size_for_type(projectile_type: String) -> int:
	return _pool_for_type(projectile_type).size()

func _projectile_type_from_state(state: Dictionary) -> String:
	if state.has("projectile_type"):
		return str(state["projectile_type"])
	if state.has("type"):
		return str(state["type"])
	return "bullet"


func _pool_for_type(projectile_type: String) -> Array:
	if not pooled_projectile_nodes_by_type.has(projectile_type):
		pooled_projectile_nodes_by_type[projectile_type] = []
	return pooled_projectile_nodes_by_type[projectile_type]


func _contract_violation(projectile_type: String, projectile_node: Node) -> void:
	var actual_script: Script = projectile_node.get_script()
	var actual_script_path := ""
	if actual_script is Script:
		actual_script_path = actual_script.resource_path
	WorldSyncLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Projectile scene root does not satisfy its presentation contract",
		{},
		{
			"subsystem": "world_sync",
			"entity_kind": "projectile",
			"resource_kind": "scene",
			"failure_mode": "wrong_scene_root",
			"expected_type": "TorpedoPresentation" if projectile_type == "torpedo" else "BulletPresentation",
			"actual_type": projectile_node.get_class(),
			"resource_path": actual_script_path,
		}
	)


func _presentation_for_node(projectile_type: String, projectile_node: Node) -> Node2D:
	var node_2d := projectile_node as Node2D
	if node_2d == null:
		_contract_violation(projectile_type, projectile_node)
		return null
	if projectile_type == "torpedo":
		var torpedo_presentation := node_2d as TorpedoPresentation
		if torpedo_presentation == null:
			_contract_violation(projectile_type, projectile_node)
			return null
		return torpedo_presentation
	var bullet_presentation := node_2d as BulletPresentation
	if bullet_presentation == null:
		_contract_violation(projectile_type, projectile_node)
		return null
	return bullet_presentation


func is_deleted(bullet_id: String) -> bool:
	return deleted_projectile_ids.has(bullet_id)


func get_projectile_node(bullet_id: String, state: Dictionary) -> Node2D:
	if projectile_nodes.has(bullet_id):
		return projectile_nodes[bullet_id]

	var projectile_type := _projectile_type_from_state(state)
	return _acquire_projectile_node(bullet_id, projectile_type)

func _acquire_projectile_node(bullet_id: String, projectile_type: String) -> Node2D:
	var pool := _pool_for_type(projectile_type)
	if not pool.is_empty():
		var pooled_node: Node = pool.pop_back()
		var pooled_presentation := _presentation_for_node(projectile_type, pooled_node)
		if pooled_presentation == null:
			pooled_node.queue_free()
			return null
		reused_projectile_node_count += 1
		if projectile_type == "torpedo":
			(pooled_presentation as TorpedoPresentation).reset_from_pool()
		else:
			(pooled_presentation as BulletPresentation).reset_from_pool()
		projectile_nodes[bullet_id] = pooled_presentation
		projectile_node_types[bullet_id] = projectile_type
		_record_lifecycle("created")
		return pooled_presentation

	if bullets_layer == null:
		WorldSyncLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
			"Cannot create projectile without a configured bullets layer",
			{},
			{
				"subsystem": "world_sync",
				"entity_kind": "projectile",
				"resource_kind": "presentation_layer",
				"failure_mode": "missing_layer",
				"expected_type": "Node2D",
				"actual_type": "null",
			}
		)
		return null

	var state_for_scene := {
		Packets.FIELD_PROJECTILE_TYPE: projectile_type
	}
	var scene_node: Node = ProjectileSceneResolver.scene_for_state(state_for_scene).instantiate()
	var presentation_node := _presentation_for_node(projectile_type, scene_node)
	if presentation_node == null:
		scene_node.queue_free()
		return null
	created_projectile_node_count += 1
	bullets_layer.add_child(presentation_node)
	projectile_nodes[bullet_id] = presentation_node
	projectile_node_types[bullet_id] = projectile_type
	_record_lifecycle("created")

	return presentation_node

func _play_projectile_firing_sound(projectile_node: Node) -> void:
	var sound := projectile_node.get_node_or_null("FiringSound") as AudioStreamPlayer2D
	if sound == null:
		return
	audio_flow.play_projectile_firing_sound(sound, bullets_layer)

func apply_projectile(
	bullet_id: String,
	state: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2,
	create_if_missing: bool = true
) -> void:
	if create_if_missing:
		_clear_deleted_projectile_id(bullet_id)
	elif deleted_projectile_ids.has(bullet_id):
		return

	if !create_if_missing and !projectile_nodes.has(bullet_id):
		return

	var bullet_node: Node2D = get_projectile_node(bullet_id, state)
	if bullet_node == null:
		return
	_record_lifecycle("updated")
	var server_position := ProjectileSyncState.server_position(state)
	var visual_position := local_visual_position + WorldWrapScript.shortest_delta(
		local_server_position,
		server_position
	)
	var server_rotation: float = state[Packets.FIELD_ROTATION]

	target_projectile_positions[bullet_id] = visual_position
	target_projectile_rotations[bullet_id] = server_rotation

	if !initialized_projectiles.has(bullet_id):
		initialized_projectiles[bullet_id] = true
		bullet_node.global_position = visual_position
		bullet_node.rotation = server_rotation
		bullet_node.visible = true
		_play_projectile_firing_sound(bullet_node)

func apply(
	server_bullets: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for bullet_id in server_bullets.keys():
		apply_projectile(bullet_id, server_bullets[bullet_id], local_visual_position, local_server_position)

func remove_projectile(bullet_id: String) -> void:
	if not deleted_projectile_ids.has(bullet_id):
		deleted_projectile_ids[bullet_id] = true
		_deleted_projectile_id_order.append(bullet_id)
		while _deleted_projectile_id_order.size() > DELETED_PROJECTILE_ID_CAP:
			deleted_projectile_ids.erase(_deleted_projectile_id_order.pop_front())
	if !projectile_nodes.has(bullet_id):
		return

	_release_projectile_node(bullet_id)

func _release_projectile_node(bullet_id: String) -> void:
	if !projectile_nodes.has(bullet_id):
		return

	released_projectile_node_count += 1
	var projectile_type := str(projectile_node_types.get(bullet_id, "bullet"))
	projectile_node_types.erase(bullet_id)
	var bullet_node: Node2D = projectile_nodes[bullet_id]
	projectile_nodes.erase(bullet_id)
	initialized_projectiles.erase(bullet_id)
	target_projectile_positions.erase(bullet_id)
	target_projectile_rotations.erase(bullet_id)
	bullet_node.visible = false
	if projectile_type == "torpedo":
		(bullet_node as TorpedoPresentation).reset_for_pool()
	else:
		(bullet_node as BulletPresentation).reset_for_pool()
	_pool_for_type(projectile_type).append(bullet_node)
	_record_lifecycle("removed")

func remove_missing(server_bullets: Dictionary) -> void:
	for bullet_id in projectile_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		remove_projectile(bullet_id)


func _record_lifecycle(operation: String) -> void:
	if _measurement_observer.is_valid():
		_measurement_observer.call("bullets", operation, 1)

func reset() -> void:
	for projectile_node in projectile_nodes.values():
		projectile_node.queue_free()
	for pool in pooled_projectile_nodes_by_type.values():
		for pooled_projectile_node in pool:
			pooled_projectile_node.queue_free()
	projectile_nodes.clear()
	projectile_node_types.clear()
	initialized_projectiles.clear()
	pooled_projectile_nodes_by_type.clear()
	target_projectile_positions.clear()
	target_projectile_rotations.clear()
	deleted_projectile_ids.clear()
	_deleted_projectile_id_order.clear()

func _clear_deleted_projectile_id(bullet_id: String) -> void:
	if not deleted_projectile_ids.has(bullet_id):
		return
	deleted_projectile_ids.erase(bullet_id)
	_deleted_projectile_id_order.erase(bullet_id)


func metrics_snapshot() -> Dictionary:
	var by_type := {}
	for type in pooled_projectile_nodes_by_type.keys():
		by_type[type] = _pool_for_type(type).size()
	return {
		"active_projectile_nodes": projectile_nodes.size(),
		"pooled_projectile_nodes": pool_size(),
		"pooled_projectile_nodes_by_type": by_type,
		"created_projectile_node_count": created_projectile_node_count,
		"reused_projectile_node_count": reused_projectile_node_count,
		"released_projectile_node_count": released_projectile_node_count,
	}


func interpolate(weight: float) -> void:
	for bullet_id in projectile_nodes.keys():
		if !target_projectile_positions.has(bullet_id):
			continue

		var bullet_node = projectile_nodes[bullet_id]
		bullet_node.global_position = bullet_node.global_position.lerp(
			target_projectile_positions[bullet_id],
			weight
		)
		bullet_node.rotation = lerp_angle(bullet_node.rotation, target_projectile_rotations[bullet_id], weight)

func projectile_target_positions() -> Dictionary:
	var positions := {}
	for bullet_id in target_projectile_positions.keys():
		if not projectile_nodes.has(bullet_id):
			continue
		var bullet_node = projectile_nodes[bullet_id]
		positions[bullet_id] = {
			"visual_position": bullet_node.global_position,
			"server_position": target_projectile_positions[bullet_id],
		}
	return positions
