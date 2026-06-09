extends Control

const MultiplayerDialogScene := preload("res://scenes/ui/dialogs/multiplayer_dialog.tscn")

signal single_player_pressed
signal sign_in_requested
signal logout_requested
signal multiplayer_create_requested
signal multiplayer_join_requested(room_code: String)

var login_status_label: Label
var logout_button: TextureButton
var multiplayer_button: TextureButton
var multiplayer_label: Label
var sign_in_label: Label
var multiplayer_dialog: Control
var _signed_in := false


func _ready() -> void:
	login_status_label = find_child("LoginStatusLabel", true, false) as Label
	logout_button = find_child("LogoutButton", true, false) as TextureButton
	multiplayer_button = find_child("MultiplayerButton", true, false) as TextureButton
	multiplayer_label = find_child("Multi-player", true, false) as Label
	sign_in_label = find_child("Sign-in", true, false) as Label
	var single_player_button := find_child("SinglePlayerButton", true, false) as TextureButton
	var quit_button := find_child("QuitButton", true, false) as TextureButton

	if login_status_label == null:
		push_error("Missing label: LoginStatusLabel")
	if logout_button == null:
		push_error("Missing button: LogoutButton")
	if multiplayer_button == null:
		push_error("Missing button: MultiplayerButton")
	if multiplayer_label == null:
		push_error("Missing label: Multi-player")
	if sign_in_label == null:
		push_error("Missing label: Sign-in")
	if single_player_button == null:
		push_error("Missing button: SinglePlayerButton")
	else:
		single_player_button.pressed.connect(_on_single_player_pressed)

	if multiplayer_button != null:
		multiplayer_button.pressed.connect(_on_multiplayer_pressed)

	if logout_button != null:
		logout_button.pressed.connect(_on_logout_pressed)

	if quit_button == null:
		push_error("Missing button: QuitButton")
	else:
		quit_button.pressed.connect(_on_quit_pressed)

	show_signed_out()


func _on_single_player_pressed() -> void:
	single_player_pressed.emit()


func _on_multiplayer_pressed() -> void:
	if !_signed_in:
		sign_in_requested.emit()
		return

	_open_multiplayer_dialog()


func _on_logout_pressed() -> void:
	logout_requested.emit()


func _on_quit_pressed() -> void:
	get_tree().quit()


func _open_multiplayer_dialog() -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		multiplayer_dialog.show()
		return

	multiplayer_dialog = MultiplayerDialogScene.instantiate()
	add_child(multiplayer_dialog)
	multiplayer_dialog.create_room_requested.connect(_on_dialog_create_room_requested)
	multiplayer_dialog.join_room_requested.connect(_on_dialog_join_room_requested)
	multiplayer_dialog.canceled.connect(_on_dialog_canceled)
	multiplayer_dialog.show()


func _on_dialog_create_room_requested() -> void:
	multiplayer_create_requested.emit()
	_clear_multiplayer_dialog()


func _on_dialog_join_room_requested(room_code: String) -> void:
	multiplayer_join_requested.emit(room_code)


func _on_dialog_canceled() -> void:
	_clear_multiplayer_dialog()


func show_multiplayer_dialog_status(message: String) -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		if multiplayer_dialog.has_method("show_join_error"):
			multiplayer_dialog.show_join_error(message)
		elif multiplayer_dialog.has_method("set_status"):
			multiplayer_dialog.set_status(message)


func show_signed_out() -> void:
	_signed_in = false
	if login_status_label != null:
		login_status_label.text = "Not Signed In"
	if logout_button != null:
		logout_button.visible = false
	if sign_in_label != null:
		sign_in_label.visible = true
	if multiplayer_label != null:
		multiplayer_label.visible = false


func show_signed_in(display_name: String) -> void:
	_signed_in = true
	if login_status_label != null:
		login_status_label.text = display_name
	if logout_button != null:
		logout_button.visible = true
	if sign_in_label != null:
		sign_in_label.visible = false
	if multiplayer_label != null:
		multiplayer_label.visible = true


func _clear_multiplayer_dialog() -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		multiplayer_dialog.queue_free()
	multiplayer_dialog = null
