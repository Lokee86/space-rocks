extends RefCounted
class_name PlayerSyncPresentation


func apply_remote_player_presentation(
	player_id: String,
	self_id: String,
	player_node: Node,
	is_paused: bool,
	remote_afterburner_active: bool
) -> void:
	if player_id != self_id:
		player_node.visible = !is_paused

	if player_node.has_method("set_remote_afterburner_visual_active"):
		player_node.set_remote_afterburner_visual_active(remote_afterburner_active)
