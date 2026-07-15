extends RefCounted

var connection_service
var logger: Callable


func _init(connection_service_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	logger = logger_callable


func send_ready_requested(ready: bool) -> void:
	connection_service.send_set_ready_request(ready)


func send_start_game_requested() -> void:
	connection_service.send_start_game_request()


func send_leave_requested() -> void:
	connection_service.send_leave_room_request()
