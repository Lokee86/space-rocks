extends RefCounted


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
			return "Invalid room ID."
		"room_not_found":
			return "Room not found."
		"room_full":
			return "Room is full."
		"room_in_game":
			return "Room is already in game."
		"already_in_room":
			return "Already in a room."
		"invalid_room_state":
			return "Room is not joinable."

	if !message.is_empty():
		return message
	return "Could not join room."
