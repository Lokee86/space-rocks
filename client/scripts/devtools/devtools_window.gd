extends Window

@onready var debug_status_label: Label = $MarginContainer/VBoxContainer/DebugStatusLabel


func _ready() -> void:
	if !close_requested.is_connected(_on_close_requested):
		close_requested.connect(_on_close_requested)


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
	if debug_status_label == null:
		return
	debug_status_label.text = "\n".join([
		"Invincible: %s" % _on_off(status.get("invincible", false)),
		"Infinite lives: %s" % _on_off(status.get("infinite_lives", false)),
		"World frozen: %s" % _on_off(status.get("world_frozen", false)),
		"Player frozen: %s" % _on_off(status.get("player_frozen", false)),
	])


func _on_close_requested() -> void:
	hide_window()


func _on_off(value) -> String:
	if bool(value):
		return "ON"
	return "OFF"
