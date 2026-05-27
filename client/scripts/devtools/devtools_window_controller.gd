class_name DevtoolsWindowController
extends RefCounted

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
