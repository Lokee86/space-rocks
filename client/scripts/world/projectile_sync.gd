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
var target_projectile_positions := {}
var target_projectile_rotations := {}


func configure(layer: Node2D) -> void:
	bullets_layer = layer


func has_projectile(bullet_id: String) -> bool:
	return projectile_nodes.has(bullet_id)


func get_projectile_node(bullet_id: String, state: Dictionary):
	if projectile_nodes.has(bullet_id):
		return projectile_nodes[bullet_id]

	var bullet_node = ProjectileSceneResolver.scene_for_state(state).instantiate()
	bullets_layer.add_child(bullet_node)
	projectile_nodes[bullet_id] = bullet_node

	return bullet_node


func _play_projectile_firing_sound(projectile_node: Node) -> void:
	var sound := projectile_node.get_node_or_null("FiringSound") as AudioStreamPlayer2D
	if sound == null:
		return
	audio_flow.play_projectile_firing_sound(sound, bullets_layer)


func apply(
	server_bullets: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for bullet_id in server_bullets.keys():
		var state: Dictionary = server_bullets[bullet_id]
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
			_play_projectile_firing_sound(bullet_node)


func remove_missing(server_bullets: Dictionary) -> void:
	for bullet_id in projectile_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		projectile_nodes[bullet_id].queue_free()
		projectile_nodes.erase(bullet_id)
		initialized_projectiles.erase(bullet_id)
		target_projectile_positions.erase(bullet_id)
		target_projectile_rotations.erase(bullet_id)


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
