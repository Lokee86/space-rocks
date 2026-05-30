extends RefCounted


var connection_service


func configure(connection_service_ref) -> void:
	connection_service = connection_service_ref


func process() -> void:
	if Input.is_action_just_pressed("DevToggle5") && connection_service != null:
		connection_service.send_debug_kill_player_request()
