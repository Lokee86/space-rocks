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
	DisplayServer.window_set_min_size(Vector2i(Constants.WINDOW_MIN_SIZE))
	DisplayServer.window_set_max_size(Vector2i(Constants.WINDOW_MAX_SIZE))
	_clamp_window_size()
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


func _clamp_window_size() -> void:
	var current_size := DisplayServer.window_get_size()
	var clamped_size := Vector2i(
		clampi(current_size.x, int(Constants.WINDOW_MIN_SIZE.x), int(Constants.WINDOW_MAX_SIZE.x)),
		clampi(current_size.y, int(Constants.WINDOW_MIN_SIZE.y), int(Constants.WINDOW_MAX_SIZE.y))
	)
	if clamped_size != current_size:
		DisplayServer.window_set_size(clamped_size)


func _start_single_player() -> void:
	_start_game("")


func _start_multiplayer(room_id: String) -> void:
	_start_game(room_id)


func _start_game(room_id: String) -> void:
	if game_loop != null:
		return

	game_loop = GAME_LOOP_SCENE.instantiate()
	if game_loop.has_method("set_room_id"):
		game_loop.set_room_id(room_id)
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
	if main_menu != null && main_menu.has_signal("multiplayer_room_requested"):
		main_menu.multiplayer_room_requested.connect(_start_multiplayer)


func set_gameplay_scroll_offset(offset: Vector2) -> void:
	gameplay_scroll_offset = offset


func clear_gameplay_scroll_offset() -> void:
	gameplay_scroll_offset = Vector2.ZERO
