extends RefCounted

var connection_service
var logger: Callable


func _init(connection_service_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	logger = logger_callable


func send_ready_requested(ready: bool) -> void:
	connection_service.send_set_ready_request(ready)
	_log("Lobby ready requested: %s" % ready)


func send_start_game_requested() -> void:
	connection_service.send_start_game_request()
	_log("Lobby start game requested")


func send_leave_requested() -> void:
	connection_service.send_leave_room_request()
	_log("Lobby leave requested")


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
