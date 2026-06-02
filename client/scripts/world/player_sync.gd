extends RefCounted
class_name PlayerSync

const Constants = preload("res://scripts/constants/constants.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const Packets = preload("res://scripts/networking/packets/packets.gd")
const VisualSyncPositions = preload("res://scripts/world/visual_sync_positions.gd")

var owner_node: Node2D
var local_player: Player
var player_nodes := {}
var initialized_players := {}
var target_player_positions := {}
var target_player_rotations := {}
var remote_player_visual_positions := {}
var view_target_player_id := ""
var player_hue_presenter := PlayerHuePresenter.new()
var pause_state_tracker


func configure(game_owner: Node2D, player: Player, pause_tracker = null) -> void:
	owner_node = game_owner
	local_player = player
	pause_state_tracker = pause_tracker
	local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX


func reset() -> void:
	player_hue_presenter.reset()
	view_target_player_id = ""
	if local_player != null:
		local_player.hide()

	for player_node in player_nodes.values():
		if player_node != local_player:
			player_node.queue_free()

	player_nodes.clear()
	initialized_players.clear()
	target_player_positions.clear()
	target_player_rotations.clear()
	remote_player_visual_positions.clear()


func set_view_target_player(player_id: String) -> void:
	view_target_player_id = player_id
	_make_local_camera_current()


func clear_view_target_player() -> void:
	view_target_player_id = ""
	_make_local_camera_current()


func focus_camera_on_player(player_id: String) -> bool:
	if !player_nodes.has(player_id):
		return false

	view_target_player_id = player_id
	_make_local_camera_current()

	return true


func _make_local_camera_current() -> bool:
	if local_player == null:
		return false

	var camera := local_player.get_node_or_null("Camera2D") as Camera2D
	if camera == null:
		return false

	camera.make_current()
	return true


func get_player_node(self_id: String, player_id: String):
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		local_player.visible = true
		local_player.z_index = Constants.LOCAL_PLAYER_Z_INDEX
		player_hue_presenter.apply_local_player_hue(local_player)
		player_nodes[player_id] = local_player
		return local_player

	var remote_player = PLAYER_SCENE.instantiate()
	remote_player.z_index = Constants.REMOTE_PLAYER_Z_INDEX
	apply_remote_player_hue(player_id, remote_player)
	owner_node.add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func apply(
	self_id: String,
	server_players: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	if local_player != null:
		player_hue_presenter.apply_local_player_hue(local_player)

	var remote_player_ids := []
	for player_id in server_players.keys():
		if player_id == self_id:
			continue
		remote_player_ids.append(player_id)
	remote_player_ids.sort()
	player_hue_presenter.set_remote_player_order(remote_player_ids)

	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node = get_player_node(self_id, player_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var visual_position := server_position
		var server_rotation: float = state[Packets.FIELD_ROTATION]
		var is_paused: bool = false
		if pause_state_tracker != null:
			is_paused = pause_state_tracker.is_paused(player_id)
		var remote_afterburner_active := false
		if player_id != self_id:
			var thrusting := bool(state.get("thrusting", false))
			remote_afterburner_active = thrusting && !is_paused

		if player_id == self_id:
			visual_position = local_visual_position
		else:
			visual_position = VisualSyncPositions.relative_to_local_visual(
				local_visual_position,
				local_server_position,
				server_position
			)
			correct_remote_visual_copy_mismatch(player_id, player_node, visual_position)
			apply_remote_player_hue(player_id, player_node)
			if player_node.has_method("set_remote_afterburner_visual_active"):
				player_node.set_remote_afterburner_visual_active(remote_afterburner_active)

		target_player_positions[player_id] = visual_position
		target_player_rotations[player_id] = server_rotation
		if player_id != self_id:
			player_node.visible = !is_paused

		if !initialized_players.has(player_id):
			initialized_players[player_id] = true
			player_node.position = visual_position
			player_node.rotation = server_rotation


func correct_remote_visual_copy_mismatch(
	player_id: String,
	player_node: Node2D,
	visual_position: Vector2
) -> void:
	# Remote targets are local-relative, but rendered remotes can briefly stay in
	# an old visual copy; snap cache/render state before interpolation crosses it.
	if !initialized_players.has(player_id):
		return
	if !VisualSyncPositions.is_world_copy_mismatch(player_node.position, visual_position):
		return

	player_node.position = visual_position
	target_player_positions[player_id] = visual_position
	remote_player_visual_positions[player_id] = visual_position


func apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	player_hue_presenter.apply_remote_player_hue(player_id, remote_player)


func get_remote_player_hues(current_self_id: String) -> Dictionary:
	return player_hue_presenter.remote_player_hues_without(current_self_id)


func remove_missing(server_players: Dictionary, self_id: String) -> void:
	for player_id in player_nodes.keys():
		if server_players.has(player_id):
			continue

		remove_player_node(self_id, player_id)


func remove_player_node(self_id: String, player_id: String) -> void:
	if !player_nodes.has(player_id):
		if player_id == self_id:
			local_player.stop_transient_effects()
			local_player.visible = false
		return

	var player_node = player_nodes[player_id]
	if player_id == view_target_player_id:
		clear_view_target_player()

	if player_node == local_player:
		local_player.stop_transient_effects()
		local_player.visible = false
	else:
		if player_node.has_method("stop_transient_effects"):
			player_node.stop_transient_effects()
		player_node.queue_free()

	player_nodes.erase(player_id)
	initialized_players.erase(player_id)
	target_player_positions.erase(player_id)
	target_player_rotations.erase(player_id)
	remote_player_visual_positions.erase(player_id)
	player_hue_presenter.remove_player(player_id)


func interpolate(weight: float, current_self_id: String) -> void:
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(
			target_player_positions[player_id],
			weight
		)
		player_node.rotation = lerp_angle(
			player_node.rotation,
			target_player_rotations[player_id],
			weight
		)
		if player_id == current_self_id:
			remote_player_visual_positions.erase(player_id)
		else:
			remote_player_visual_positions[player_id] = player_node.position

	if view_target_player_id == "":
		return
	if !player_nodes.has(view_target_player_id):
		return

	var view_target_player = player_nodes[view_target_player_id]
	if view_target_player == local_player:
		return

	local_player.global_position = view_target_player.global_position
	local_player.rotation = view_target_player.rotation
	local_player.visible = false


func get_remote_player_visual_positions(current_self_id: String) -> Dictionary:
	var positions := remote_player_visual_positions.duplicate()
	positions.erase(current_self_id)
	return positions
