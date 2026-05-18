extends Node2D

const Constants = preload("res://scripts/constants.gd")
const GAME_LOOP_SCENE := preload("res://scenes/game_loop.tscn")
const MAIN_MENU_SCENE := preload("res://scenes/ui/main_menu.tscn")
const BACKGROUND_DRIFT := Vector2(18.0, 8.0)
const FOREGROUND_DRIFT := Vector2(42.0, 18.0)

@onready var background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var main_menu: Control = $CanvasLayer/MainMenu

var game_loop: Node
var gameplay_scroll_offset := Vector2.ZERO
var drift_time := 0.0


func _ready() -> void:
	_connect_main_menu()


func _process(delta: float) -> void:
	drift_time += delta
	_update_layer_shader(
		background,
		(BACKGROUND_DRIFT * drift_time) + (gameplay_scroll_offset * Constants.BACKGROUND_PARALLAX)
	)
	_update_layer_shader(
		foreground_background,
		Constants.FOREGROUND_BACKGROUND_OFFSET +
			(FOREGROUND_DRIFT * drift_time) +
			(gameplay_scroll_offset * Constants.FOREGROUND_BACKGROUND_PARALLAX)
	)


func _update_layer_shader(layer: TextureRect, scroll_offset: Vector2) -> void:
	if layer == null:
		return

	var background_material := layer.material as ShaderMaterial
	if background_material == null:
		return

	background_material.set_shader_parameter("scroll_offset", scroll_offset)


func _start_single_player() -> void:
	if game_loop != null:
		return

	game_loop = GAME_LOOP_SCENE.instantiate()
	if game_loop.has_signal("return_to_menu_requested"):
		game_loop.return_to_menu_requested.connect(_return_to_main_menu)
	add_child(game_loop)

	if main_menu != null:
		main_menu.queue_free()
		main_menu = null


func _return_to_main_menu() -> void:
	clear_gameplay_scroll_offset()
	if game_loop != null:
		game_loop.queue_free()
		game_loop = null

	if main_menu == null:
		main_menu = MAIN_MENU_SCENE.instantiate()
		canvas_layer.add_child(main_menu)
		_connect_main_menu()


func _connect_main_menu() -> void:
	if main_menu != null && main_menu.has_signal("single_player_pressed"):
		main_menu.single_player_pressed.connect(_start_single_player)


func set_gameplay_scroll_offset(offset: Vector2) -> void:
	gameplay_scroll_offset = offset


func clear_gameplay_scroll_offset() -> void:
	gameplay_scroll_offset = Vector2.ZERO
