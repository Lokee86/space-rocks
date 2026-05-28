class_name DevtoolsWindowController
extends RefCounted

signal toggle_invincible_requested
signal toggle_infinite_lives_requested
signal toggle_freeze_world_requested
signal toggle_freeze_player_requested

const DevtoolsWindowScene := preload("res://scenes/devtools/devtools_window.tscn")

var window: Window
var parent: Node
var latest_debug_status := {}


func ensure_window() -> Window:
	if window != null && is_instance_valid(window):
		return window

	window = DevtoolsWindowScene.instantiate()
	parent = Engine.get_main_loop().root
	parent.add_child(window)
	_connect_window_signals()
	window.set_debug_status(latest_debug_status)
	return window


func show_window() -> void:
	ensure_window().show_window()


func hide_window() -> void:
	if window != null && is_instance_valid(window):
		window.hide_window()


func toggle_window() -> void:
	var devtools_window := ensure_window()
	if devtools_window.visible:
		devtools_window.hide_window()
	else:
		devtools_window.show_window()


func apply_debug_status(status: Dictionary) -> void:
	latest_debug_status = status
	if window != null && is_instance_valid(window):
		window.set_debug_status(latest_debug_status)


func _connect_window_signals() -> void:
	if !window.toggle_invincible_requested.is_connected(_on_toggle_invincible_requested):
		window.toggle_invincible_requested.connect(_on_toggle_invincible_requested)
	if !window.toggle_infinite_lives_requested.is_connected(_on_toggle_infinite_lives_requested):
		window.toggle_infinite_lives_requested.connect(_on_toggle_infinite_lives_requested)
	if !window.toggle_freeze_world_requested.is_connected(_on_toggle_freeze_world_requested):
		window.toggle_freeze_world_requested.connect(_on_toggle_freeze_world_requested)
	if !window.toggle_freeze_player_requested.is_connected(_on_toggle_freeze_player_requested):
		window.toggle_freeze_player_requested.connect(_on_toggle_freeze_player_requested)


func _on_toggle_invincible_requested() -> void:
	toggle_invincible_requested.emit()


func _on_toggle_infinite_lives_requested() -> void:
	toggle_infinite_lives_requested.emit()


func _on_toggle_freeze_world_requested() -> void:
	toggle_freeze_world_requested.emit()


func _on_toggle_freeze_player_requested() -> void:
	toggle_freeze_player_requested.emit()
