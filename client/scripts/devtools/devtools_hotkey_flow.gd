extends RefCounted
class_name DevtoolsHotkeyFlow

var respawn_local_player_route: Callable
var placement_request_route: Callable


func configure(respawn_route: Callable, placement_route: Callable) -> void:
	respawn_local_player_route = respawn_route
	placement_request_route = placement_route


func process(has_received_gameplay_state: bool) -> void:
	if !has_received_gameplay_state:
		return

	if Input.is_action_just_pressed("DevToggle5"):
		if !respawn_local_player_route.is_null():
			respawn_local_player_route.call()

	if Input.is_action_just_pressed("DevToggle6"):
		if placement_request_route.is_null():
			return
		if Input.is_key_pressed(KEY_ALT):
			placement_request_route.call(&"spawn_bullet")
		elif Input.is_key_pressed(KEY_SHIFT):
			placement_request_route.call(&"spawn_asteroid")
		else:
			placement_request_route.call(&"spawn_player")
