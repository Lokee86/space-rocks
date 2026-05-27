extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")


func show_status(main_menu: Control, message: String) -> void:
	if main_menu != null && is_instance_valid(main_menu) && main_menu.has_method("show_multiplayer_dialog_status"):
		main_menu.show_multiplayer_dialog_status(message)


func show_room_error(main_menu: Control, packet: Dictionary) -> void:
	var error_code := str(packet.get("error_code", ""))
	var message := str(packet.get("message", ""))
	var friendly_message := friendly_room_error_message(error_code, message)
	show_status(main_menu, friendly_message)


func friendly_room_error_message(error_code: String, message: String) -> String:
	match error_code:
		"invalid_room_code":
			return Constants.DIALOG_STATUS_INVALID_ROOM_ID
		"room_not_found":
			return Constants.DIALOG_STATUS_ROOM_NOT_FOUND
		"room_full":
			return Constants.DIALOG_STATUS_ROOM_FULL
		"room_in_game":
			return Constants.DIALOG_STATUS_ROOM_IN_GAME
		"already_in_room":
			return Constants.DIALOG_STATUS_ALREADY_IN_ROOM
		"invalid_room_state":
			return Constants.DIALOG_STATUS_ROOM_NOT_JOINABLE

	if !message.is_empty():
		return message
	return Constants.DIALOG_STATUS_COULD_NOT_JOIN
