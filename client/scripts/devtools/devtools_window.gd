extends Window

signal toggle_invincible_requested
signal toggle_infinite_lives_requested
signal toggle_freeze_world_requested
signal toggle_freeze_player_requested

@onready var invincible_button: Button = %InvincibleButton
@onready var infinite_lives_button: Button = %InfiniteLivesButton
@onready var freeze_world_button: Button = %FreezeWorldButton
@onready var freeze_player_button: Button = %FreezePlayerButton
@onready var invincible_status_label: Label = %InvincibleStatusLabel
@onready var infinite_lives_status_label: Label = %InfiniteLivesStatusLabel
@onready var world_frozen_status_label: Label = %WorldFrozenStatusLabel
@onready var player_frozen_status_label: Label = %PlayerFrozenStatusLabel


func _ready() -> void:
	if !close_requested.is_connected(_on_close_requested):
		close_requested.connect(_on_close_requested)
	if !invincible_button.pressed.is_connected(_on_invincible_button_pressed):
		invincible_button.pressed.connect(_on_invincible_button_pressed)
	if !infinite_lives_button.pressed.is_connected(_on_infinite_lives_button_pressed):
		infinite_lives_button.pressed.connect(_on_infinite_lives_button_pressed)
	if !freeze_world_button.pressed.is_connected(_on_freeze_world_button_pressed):
		freeze_world_button.pressed.connect(_on_freeze_world_button_pressed)
	if !freeze_player_button.pressed.is_connected(_on_freeze_player_button_pressed):
		freeze_player_button.pressed.connect(_on_freeze_player_button_pressed)


func show_window() -> void:
	popup_centered()


func hide_window() -> void:
	hide()


func toggle_window() -> void:
	if visible:
		hide_window()
	else:
		show_window()


func set_debug_status(status: Dictionary) -> void:
	invincible_status_label.text = "Invincible: %s" % _on_off(status.get("invincible", false))
	infinite_lives_status_label.text = "Infinite lives: %s" % _on_off(status.get("infinite_lives", false))
	world_frozen_status_label.text = "World frozen: %s" % _on_off(status.get("world_frozen", false))
	player_frozen_status_label.text = "Player frozen: %s" % _on_off(status.get("player_frozen", false))


func _on_close_requested() -> void:
	hide_window()


func _on_invincible_button_pressed() -> void:
	toggle_invincible_requested.emit()


func _on_infinite_lives_button_pressed() -> void:
	toggle_infinite_lives_requested.emit()


func _on_freeze_world_button_pressed() -> void:
	toggle_freeze_world_requested.emit()


func _on_freeze_player_button_pressed() -> void:
	toggle_freeze_player_requested.emit()


func _on_off(value) -> String:
	if bool(value):
		return "ON"
	return "OFF"
