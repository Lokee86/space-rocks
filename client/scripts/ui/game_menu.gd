extends Control
class_name GameMenu

signal primary_action_requested
signal lobby_requested
signal spectate_requested
signal menu_requested
signal resume_requested
signal quit_requested

@onready var primary_action_button: BaseButton = _find_button(["PrimaryActionButton", "ResumeButton", "LeftButton"])
@onready var menu_button: BaseButton = _find_button(["MenuButton", "QuitButton"])


const SESSION_MODE_MULTIPLAYER := "multiplayer"
const PRIMARY_ACTION_RESUME := "resume"
const PRIMARY_ACTION_LOBBY := "lobby"
const PRIMARY_ACTION_SPECTATE := "spectate"
const PRIMARY_ACTION_WAITING := "waiting"

var primary_action := PRIMARY_ACTION_RESUME


func _ready() -> void:
	if primary_action_button != null:
		primary_action_button.pressed.connect(_on_primary_action_pressed)
	else:
		push_error("Game menu is missing PrimaryActionButton.")

	if menu_button != null:
		menu_button.pressed.connect(_on_menu_pressed)
	else:
		push_error("Game menu is missing MenuButton.")


func set_primary_text(text) -> void:
	var button_text := str(text)
	_set_button_label_visible(primary_action_button, _primary_label_name(button_text))


func set_primary_enabled(enabled) -> void:
	if primary_action_button == null:
		return

	primary_action_button.disabled = !bool(enabled)


func set_menu_text(_text) -> void:
	_set_button_label_visible(menu_button, "Menu")


func configure_for_state(session_mode: String, game_over: bool, room_state: String, has_spectate_targets := false) -> void:
	var normalized_session_mode := _normalized_state(session_mode)
	var room_game_over := _normalized_state(room_state) == "gameover"

	if normalized_session_mode == SESSION_MODE_MULTIPLAYER:
		if game_over:
			if room_game_over:
				primary_action = PRIMARY_ACTION_LOBBY
				set_primary_text("Lobby")
				set_primary_enabled(true)
			elif has_spectate_targets:
				primary_action = PRIMARY_ACTION_SPECTATE
				set_primary_text("Spectate")
				set_primary_enabled(true)
			else:
				primary_action = PRIMARY_ACTION_WAITING
				set_primary_text("Waiting")
				set_primary_enabled(false)
		else:
			primary_action = PRIMARY_ACTION_RESUME
			set_primary_text("Resume")
			set_primary_enabled(true)
		set_menu_text("Main Menu")
		return

	primary_action = PRIMARY_ACTION_RESUME
	set_primary_text("Resume")
	set_primary_enabled(!game_over)
	set_menu_text("Main Menu")


func _on_primary_action_pressed() -> void:
	primary_action_requested.emit()
	if primary_action == PRIMARY_ACTION_LOBBY:
		lobby_requested.emit()
		return
	if primary_action == PRIMARY_ACTION_SPECTATE:
		spectate_requested.emit()
		return

	resume_requested.emit()


func _on_menu_pressed() -> void:
	menu_requested.emit()
	quit_requested.emit()


func _find_button(names: Array) -> BaseButton:
	for button_name in names:
		var button := find_child(button_name, true, false) as BaseButton
		if button != null:
			return button

	return null


func _normalized_state(value) -> String:
	return str(value).strip_edges().replace("_", "").to_lower()


func _primary_label_name(text: String) -> String:
	if _normalized_state(text) == PRIMARY_ACTION_LOBBY:
		return "Lobby"
	if _normalized_state(text) == PRIMARY_ACTION_SPECTATE:
		return "Spectate"
	if _normalized_state(text) == PRIMARY_ACTION_WAITING:
		return "Waiting"

	return "Resume"


func _set_button_label_visible(button: BaseButton, label_name: String) -> void:
	if button == null:
		return

	var fallback_label: Label = null
	var found_target := false
	for child in button.find_children("*", "Label", true, false):
		var label := child as Label
		if fallback_label == null:
			fallback_label = label
		var is_target := str(label.name) == label_name
		label.visible = is_target
		if is_target:
			found_target = true

	if !found_target && fallback_label != null:
		fallback_label.visible = true
