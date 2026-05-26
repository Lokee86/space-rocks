extends Node

const GameplayShellFlow := preload("res://scripts/shell/gameplay_shell_flow.gd")
const GameplayHudFlow := preload("res://scripts/shell/gameplay_hud_flow.gd")
const GameplayMenuFlow := preload("res://scripts/shell/gameplay_menu_flow.gd")
const GameplayBackgroundFlow := preload("res://scripts/shell/gameplay_background_flow.gd")

var connection_service
var scene_root: Node
var player
var bullets: Node2D
var asteroids: Node2D
var hud: Control
var game_over_sound: AudioStreamPlayer
var main_menu: Control
var repeated_background: TextureRect
var repeated_foreground_background: TextureRect
var logger: Callable

var gameplay_shell_flow
var gameplay_hud_flow
var gameplay_menu_flow
var gameplay_background_flow


func configure(
	connection_service_ref,
	scene_root_ref: Node,
	player_ref,
	bullets_ref: Node2D,
	asteroids_ref: Node2D,
	hud_ref: Control,
	game_over_sound_ref: AudioStreamPlayer,
	main_menu_ref: Control,
	repeated_background_ref: TextureRect,
	repeated_foreground_background_ref: TextureRect,
	logger_callable: Callable
) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	player = player_ref
	bullets = bullets_ref
	asteroids = asteroids_ref
	hud = hud_ref
	game_over_sound = game_over_sound_ref
	main_menu = main_menu_ref
	repeated_background = repeated_background_ref
	repeated_foreground_background = repeated_foreground_background_ref
	logger = logger_callable

	gameplay_hud_flow = GameplayHudFlow.new()
	gameplay_hud_flow.configure(hud)
	gameplay_menu_flow = GameplayMenuFlow.new()
	gameplay_menu_flow.configure(hud, connection_service, player)
	gameplay_background_flow = GameplayBackgroundFlow.new()
	gameplay_background_flow.configure(repeated_background, repeated_foreground_background)
	gameplay_shell_flow = GameplayShellFlow.new()
	gameplay_shell_flow.configure(
		connection_service,
		scene_root,
		player,
		bullets,
		asteroids,
		gameplay_hud_flow,
		gameplay_menu_flow,
		gameplay_background_flow,
		game_over_sound
	)
	_connect_gameplay_shell_signal("gameplay_started", Callable(self, "_on_gameplay_started"))
	_connect_gameplay_shell_signal(
		"quit_to_main_menu_requested",
		Callable(self, "_on_gameplay_quit_to_main_menu_requested")
	)


func handle_gameplay_state(packet: Dictionary) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state(packet)


func _process(delta: float) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)


func reset() -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.reset()


func _connect_gameplay_shell_signal(signal_name: StringName, handler: Callable) -> void:
	if gameplay_shell_flow.has_signal(signal_name) && !gameplay_shell_flow.is_connected(signal_name, handler):
		gameplay_shell_flow.connect(signal_name, handler)


func _on_gameplay_started() -> void:
	if main_menu != null:
		main_menu.hide()


func _on_gameplay_quit_to_main_menu_requested() -> void:
	_log("V2 gameplay quit to main menu requested")
	if connection_service != null:
		connection_service.begin_graceful_close()
	reset()
	if main_menu != null:
		main_menu.show()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
