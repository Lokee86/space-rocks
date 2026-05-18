extends Control

const GAME_SCENE := "res://scenes/game.tscn"
const BACKGROUND_DRIFT := Vector2(18.0, 8.0)
const FOREGROUND_DRIFT := Vector2(42.0, 18.0)
const FOREGROUND_OFFSET := Vector2(480.0, 270.0)

@onready var new_game_button: TextureButton = $CenterContainer/VBoxContainer/HBoxContainer/NewGameButton
@onready var quit_button: TextureButton = $CenterContainer/VBoxContainer/HBoxContainer/QuitButton
@onready var background: TextureRect = $ParallaxBackground2/BackgroundLayer/RepeatedBackground
@onready var foreground_background: TextureRect = $ParallaxBackground2/ForegroundBackgroundLayer/RepeatedBackground

var drift_time := 0.0


func _ready() -> void:
	new_game_button.pressed.connect(_start_new_game)
	quit_button.pressed.connect(_quit)


func _process(delta: float) -> void:
	drift_time += delta
	_update_layer_shader(background, BACKGROUND_DRIFT * drift_time)
	_update_layer_shader(foreground_background, FOREGROUND_OFFSET + (FOREGROUND_DRIFT * drift_time))


func _start_new_game() -> void:
	get_tree().change_scene_to_file(GAME_SCENE)


func _quit() -> void:
	get_tree().quit()


func _update_layer_shader(layer: TextureRect, scroll_offset: Vector2) -> void:
	var background_material := layer.material as ShaderMaterial
	if background_material == null:
		return

	background_material.set_shader_parameter("scroll_offset", scroll_offset)
