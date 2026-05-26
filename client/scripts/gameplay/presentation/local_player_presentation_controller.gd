extends RefCounted
class_name LocalPlayerPresentationController

var player


func configure(player_ref) -> void:
	player = player_ref


func reset() -> void:
	if player != null:
		player.set_afterburner_active(false)


func process(has_received_state: bool) -> void:
	if !has_received_state:
		return
	if player == null || !player.visible:
		return

	player.set_afterburner_active(Input.is_action_pressed(player.move_forward_action))
