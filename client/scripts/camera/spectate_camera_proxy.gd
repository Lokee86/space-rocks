extends RefCounted
class_name SpectateCameraProxy

# Spectate currently uses the local Player/Camera2D as the camera carrier.
# While spectating, the proxy copies the selected remote player's transform
# onto local_player and hides local_player.
# Do not move Camera2D out from under Player unless world presentation/view-origin
# is refactored at the same time.
var view_target_player_id := ""


func set_view_target_player(player_id: String) -> void:
	view_target_player_id = player_id


func clear_view_target_player(local_player: Player = null) -> void:
	view_target_player_id = ""
	if local_player != null:
		make_local_camera_current(local_player)


func current_view_target_player_id() -> String:
	return view_target_player_id


func apply_from_lifecycle(player_lifecycle: PlayerSyncLifecycle, local_player: Player) -> void:
	var view_target_player_id := current_view_target_player_id()
	if view_target_player_id == "":
		return
	if !player_lifecycle.has_player_node(view_target_player_id):
		return

	var view_target_player = player_lifecycle.get_player_node(view_target_player_id)
	apply(view_target_player, local_player)


func focus_camera_on_player(player_id: String, player_lifecycle: PlayerSyncLifecycle, local_player: Player) -> bool:
	if !player_lifecycle.has_player_node(player_id):
		return false

	set_view_target_player(player_id)
	make_local_camera_current(local_player)
	return true


func handle_player_removed(player_id: String, local_player: Player) -> void:
	if player_id != current_view_target_player_id():
		return

	clear_view_target_player(local_player)


func make_local_camera_current(local_player: Player) -> bool:
	if local_player == null:
		return false

	var camera := local_player.get_node_or_null("Camera2D") as Camera2D
	if camera == null:
		return false

	camera.make_current()
	return true


func apply(view_target_player: Node2D, local_player: Player) -> void:
	if view_target_player == null:
		return
	if local_player == null:
		return
	if view_target_player == local_player:
		return

	local_player.global_position = view_target_player.global_position
	local_player.rotation = view_target_player.rotation
	local_player.visible = false
