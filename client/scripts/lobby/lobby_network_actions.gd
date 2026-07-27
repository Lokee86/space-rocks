extends RefCounted

var connection_service
var logger: Callable


func _init(connection_service_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	logger = logger_callable


func send_ready_requested(ready: bool) -> void:
	connection_service.send_set_ready_request(ready)


func send_team_assignment_requested(player_id: String, team_id: String) -> void:
	connection_service.send_set_team_assignment_request(player_id, team_id)


func send_start_game_requested() -> void:
	connection_service.send_start_game_request()


func send_add_bot_requested() -> void:
	connection_service.send_add_bot_request()


func send_remove_member_requested(player_id: String) -> void:
	connection_service.send_remove_room_member_request(player_id)


func send_leave_requested() -> void:
	connection_service.send_leave_room_request()
