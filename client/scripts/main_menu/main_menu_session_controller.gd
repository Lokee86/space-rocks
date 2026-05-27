extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")
const MultiplayerDialogStatusPresenter := preload("res://scripts/lobby/multiplayer_dialog_status_presenter.gd")

var main_menu: Control
var session_boot_controller
var multiplayer_dialog_status_presenter
var logger: Callable


func configure(
	main_menu_ref: Control,
	session_boot_controller_ref,
	logger_callable: Callable
) -> void:
	main_menu = main_menu_ref
	session_boot_controller = session_boot_controller_ref
	logger = logger_callable
	multiplayer_dialog_status_presenter = MultiplayerDialogStatusPresenter.new()


func request_single_player() -> void:
	session_boot_controller.request_single_player()


func request_create_room() -> void:
	session_boot_controller.request_create_room()


func request_join_room(room_code: String) -> void:
	var stripped_room_code := room_code.strip_edges()
	if stripped_room_code.is_empty():
		_log("Multiplayer join rejected: empty room code")
		multiplayer_dialog_status_presenter.show_status(
			main_menu,
			Constants.DIALOG_STATUS_MUST_ENTER_ID
		)
		return

	session_boot_controller.request_join_room(stripped_room_code)


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
