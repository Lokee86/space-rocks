extends Node2D

const ShellState := preload("res://scripts/shell/shell_state.gd")

@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu

var shell_state: ShellState


func _ready() -> void:
	shell_state = ShellState.new()
	_connect_main_menu()
	print("V2 game shell booted: %s" % shell_state.current())


func _connect_main_menu() -> void:
	if main_menu.has_signal("single_player_pressed"):
		var single_player_callable := Callable(self, "_on_single_player_pressed")
		if not main_menu.is_connected("single_player_pressed", single_player_callable):
			main_menu.connect("single_player_pressed", single_player_callable)

	if main_menu.has_signal("multiplayer_create_requested"):
		var create_callable := Callable(self, "_on_multiplayer_create_requested")
		if not main_menu.is_connected("multiplayer_create_requested", create_callable):
			main_menu.connect("multiplayer_create_requested", create_callable)

	if main_menu.has_signal("multiplayer_join_requested"):
		var join_callable := Callable(self, "_on_multiplayer_join_requested")
		if not main_menu.is_connected("multiplayer_join_requested", join_callable):
			main_menu.connect("multiplayer_join_requested", join_callable)


func _on_single_player_pressed() -> void:
	print("V2 game shell single player pressed")


func _on_multiplayer_create_requested() -> void:
	print("V2 game shell multiplayer create requested")


func _on_multiplayer_join_requested(room_code: String) -> void:
	print("V2 game shell multiplayer join requested: %s" % room_code)
