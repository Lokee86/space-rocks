extends RefCounted
class_name ProjectileSync

const ProjectileSceneResolver = preload("res://scripts/world/projectiles/projectile_scene_resolver.gd")
const ProjectileSyncState = preload("res://scripts/world/projectile_sync_state.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

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


func configure(layer: Node2D) -> void:
	bullets_layer = layer


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


func get_projectile_node(bullet_id: String, state: Dictionary):
	if projectile_nodes.has(bullet_id):
		return projectile_nodes[bullet_id]

	var projectile_type := _projectile_type_from_state(state)
	return _acquire_projectile_node(bullet_id, projectile_type)

func _acquire_projectile_node(bullet_id: String, projectile_type: String):
	var pool := _pool_for_type(projectile_type)
	if not pool.is_empty():
		reused_projectile_node_count += 1
		var pooled_node = pool.pop_back()
		if pooled_node.has_method("reset_from_pool"):
			pooled_node.reset_from_pool()
		else:
			if pooled_node is CanvasItem:
				pooled_node.modulate = Color.WHITE
			if pooled_node is Node2D:
				pooled_node.rotation = 0.0
				pooled_node.scale = Vector2.ONE
		pooled_node.visible = false
		projectile_nodes[bullet_id] = pooled_node
		projectile_node_types[bullet_id] = projectile_type
		return pooled_node

	created_projectile_node_count += 1
	var state_for_scene := {
		Packets.FIELD_PROJECTILE_TYPE: projectile_type
	}
	var bullet_node = ProjectileSceneResolver.scene_for_state(state_for_scene).instantiate()
	bullets_layer.add_child(bullet_node)
	projectile_nodes[bullet_id] = bullet_node
	projectile_node_types[bullet_id] = projectile_type

	return bullet_node

func _play_projectile_firing_sound(projectile_node: Node) -> void:
	var sound := projectile_node.get_node_or_null("FiringSound") as AudioStreamPlayer2D
	if sound == null:
		return
	audio_flow.play_projectile_firing_sound(sound, bullets_layer)

func apply_projectile(
	bullet_id: String,
	state: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	var bullet_node = get_projectile_node(bullet_id, state)
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
	if !projectile_nodes.has(bullet_id):
		return

	_release_projectile_node(bullet_id)

func _release_projectile_node(bullet_id: String) -> void:
	if !projectile_nodes.has(bullet_id):
		return

	released_projectile_node_count += 1
	var projectile_type := str(projectile_node_types.get(bullet_id, "bullet"))
	projectile_node_types.erase(bullet_id)
	var bullet_node = projectile_nodes[bullet_id]
	projectile_nodes.erase(bullet_id)
	initialized_projectiles.erase(bullet_id)
	target_projectile_positions.erase(bullet_id)
	target_projectile_rotations.erase(bullet_id)
	bullet_node.visible = false
	if bullet_node.has_method("reset_for_pool"):
		bullet_node.reset_for_pool()
	_pool_for_type(projectile_type).append(bullet_node)

func remove_missing(server_bullets: Dictionary) -> void:
	for bullet_id in projectile_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		remove_projectile(bullet_id)

func clear_all_projectiles() -> void:
	for projectile_node in projectile_nodes.values():
		projectile_node.queue_free()
	for pool in pooled_projectile_nodes_by_type.values():
		for pooled_projectile_node in pool:
			pooled_projectile_node.queue_free()
		pool.clear()
	projectile_nodes.clear()
	projectile_node_types.clear()
	initialized_projectiles.clear()
	target_projectile_positions.clear()
	target_projectile_rotations.clear()


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