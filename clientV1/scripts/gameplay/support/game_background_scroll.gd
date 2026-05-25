extends RefCounted


func update_scroll_offset(
	shell,
	has_initial_spawn: bool,
	is_spectating: bool,
	camera_follow,
	player: Node2D
) -> void:
	if !has_initial_spawn:
		return

	if shell != null && shell.has_method("set_gameplay_scroll_offset"):
		if is_spectating && camera_follow != null && camera_follow.camera != null:
			shell.set_gameplay_scroll_offset(camera_follow.camera.global_position)
		else:
			shell.set_gameplay_scroll_offset(player.global_position)


func clear_scroll_offset(shell) -> void:
	if shell != null && shell.has_method("clear_gameplay_scroll_offset"):
		shell.clear_gameplay_scroll_offset()
