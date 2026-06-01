extends RefCounted
class_name BulletSync

signal bullet_spawned

const BulletSyncState = preload("res://scripts/world/bullet_sync_state.gd")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const Packets = preload("res://scripts/networking/packets/packets.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

var bullets_layer: Node2D
var bullet_nodes := {}
var initialized_bullets := {}
var target_bullet_positions := {}
var target_bullet_rotations := {}


func configure(layer: Node2D) -> void:
	bullets_layer = layer


func has_bullet(bullet_id: String) -> bool:
	return bullet_nodes.has(bullet_id)


func get_bullet_node(bullet_id: String):
	if bullet_nodes.has(bullet_id):
		return bullet_nodes[bullet_id]

	var bullet_node = BULLET_SCENE.instantiate()
	bullets_layer.add_child(bullet_node)
	bullet_nodes[bullet_id] = bullet_node

	return bullet_node


func apply(
	server_bullets: Dictionary,
	play_new_bullet_sounds: bool,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	for bullet_id in server_bullets.keys():
		var state: Dictionary = server_bullets[bullet_id]
		var is_new_bullet: bool = !has_bullet(bullet_id)
		var bullet_node = get_bullet_node(bullet_id)
		var server_position := BulletSyncState.server_position(state)
		var visual_position := local_visual_position + WorldWrapScript.shortest_delta(
			local_server_position,
			server_position
		)
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		target_bullet_positions[bullet_id] = visual_position
		target_bullet_rotations[bullet_id] = server_rotation

		if !initialized_bullets.has(bullet_id):
			initialized_bullets[bullet_id] = true
			bullet_node.global_position = visual_position
			bullet_node.rotation = server_rotation

		if is_new_bullet && play_new_bullet_sounds:
			bullet_spawned.emit()


func remove_missing(server_bullets: Dictionary) -> void:
	for bullet_id in bullet_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		bullet_nodes[bullet_id].queue_free()
		bullet_nodes.erase(bullet_id)
		initialized_bullets.erase(bullet_id)
		target_bullet_positions.erase(bullet_id)
		target_bullet_rotations.erase(bullet_id)


func interpolate(weight: float) -> void:
	for bullet_id in bullet_nodes.keys():
		if !target_bullet_positions.has(bullet_id):
			continue

		var bullet_node = bullet_nodes[bullet_id]
		bullet_node.global_position = bullet_node.global_position.lerp(
			target_bullet_positions[bullet_id],
			weight
		)
		bullet_node.rotation = lerp_angle(bullet_node.rotation, target_bullet_rotations[bullet_id], weight)


func bullet_target_positions() -> Dictionary:
	var positions := {}
	for bullet_id in target_bullet_positions.keys():
		if not bullet_nodes.has(bullet_id):
			continue
		var bullet_node = bullet_nodes[bullet_id]
		positions[bullet_id] = {
			"visual_position": bullet_node.global_position,
			"server_position": target_bullet_positions[bullet_id],
		}
	return positions
