extends Node2D

const SessionBootController := preload("res://scripts/boot/session_boot_controller.gd")
const MainMenuSessionController := preload("res://scripts/main_menu/main_menu_session_controller.gd")
const SessionNetworkController := preload("res://scripts/session/session_network_controller.gd")
const RoomSessionController := preload("res://scripts/session/room_session_controller.gd")
const GameplaySessionController := preload("res://scripts/session/gameplay_session_controller.gd")
const ClientConfigController := preload("res://scripts/session/client_config_controller.gd")
const AppShutdownController := preload("res://scripts/session/app_shutdown_controller.gd")
const Constants := preload("res://scripts/constants/constants.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

@onready var main_menu: Control = $CanvasLayer/MainMenu
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground
@onready var player = $Player
@onready var bullets: Node2D = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var hud: Control = $CanvasLayer/HUD
@onready var game_over_sound: AudioStreamPlayer = $CanvasLayer/HUD/CenterContainer/GameOverContainer/GameOverSound

var session_boot_controller
var main_menu_session_controller
var session_network_controller
var room_session_controller
var gameplay_session_controller
var client_config_controller
var app_shutdown_controller
var background_controller


func _ready() -> void:
	_log_v2_status("V2 app entry booted")
	get_tree().set_auto_accept_quit(false)
	_setup_boot_and_config()
	app_shutdown_controller = AppShutdownController.new()
	add_child(app_shutdown_controller)
	app_shutdown_controller.configure(session_boot_controller.get_connection_service(), get_tree())
	background_controller = BackgroundController.new()
	add_child(background_controller)
	background_controller.configure(repeated_background, repeated_foreground_background, player)
	gameplay_session_controller = GameplaySessionController.new()
	add_child(gameplay_session_controller)
	gameplay_session_controller.configure(
		session_boot_controller.get_connection_service(),
		self,
		player,
		bullets,
		asteroids,
		hud,
		game_over_sound,
		main_menu,
		session_boot_controller.get_session_context(),
		session_boot_controller.get_shell_boot_flow(),
		Callable(self, "_log_v2_status")
	)
	session_network_controller = SessionNetworkController.new()
	session_network_controller.configure(
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_shell_boot_flow(),
		Callable(self, "_log_v2_status"),
		{}
	)
	session_network_controller.connect_connection_signals()
	session_network_controller.configure_gameplay_session_controller(gameplay_session_controller)
	session_network_controller.connect_gameplay_signals()
	room_session_controller = RoomSessionController.new()
	room_session_controller.configure(
		main_menu,
		canvas_layer,
		session_boot_controller.get_session_context(),
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_shell_boot_flow(),
		Callable(self, "_log_v2_status")
	)
	room_session_controller.configure_client_config_sender(
		Callable(client_config_controller, "send_client_config")
	)
	gameplay_session_controller.configure_room_state_provider(
		Callable(room_session_controller, "current_room_state")
	)
	session_network_controller.configure_room_session_controller(room_session_controller)
	session_network_controller.connect_room_signals()
	main_menu_session_controller = MainMenuSessionController.new()
	main_menu_session_controller.configure(
		main_menu,
		session_boot_controller,
		Callable(self, "_log_v2_status")
	)
	_connect_main_menu_signals()


func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		if app_shutdown_controller != null:
			app_shutdown_controller.request_shutdown()
		else:
			get_tree().quit()


func _setup_boot_and_config() -> void:
	session_boot_controller = SessionBootController.new()
	session_boot_controller.configure(Constants.MULTIPLAYER_WS_URL, Callable(self, "_log_v2_status"))
	add_child(session_boot_controller)
	client_config_controller = ClientConfigController.new()
	client_config_controller.configure(session_boot_controller.get_connection_service(), get_viewport())
	_connect_boot_flow_signal(
		"boot_request_sent",
		Callable(client_config_controller, "send_client_config")
	)


func _connect_main_menu_signals() -> void:
	if main_menu == null:
		push_error("Missing main menu")
		return

	_connect_main_menu_signal("single_player_pressed", _on_single_player_pressed)
	_connect_main_menu_signal("multiplayer_create_requested", _on_multiplayer_create_requested)
	_connect_main_menu_signal("multiplayer_join_requested", _on_multiplayer_join_requested)


func _connect_main_menu_signal(signal_name: StringName, handler: Callable) -> void:
	if main_menu.has_signal(signal_name):
		main_menu.connect(signal_name, handler)


func _connect_boot_flow_signal(signal_name: StringName, handler: Callable) -> void:
	var shell_boot_flow = session_boot_controller.get_shell_boot_flow()
	if shell_boot_flow.has_signal(signal_name) && !shell_boot_flow.is_connected(signal_name, handler):
		shell_boot_flow.connect(signal_name, handler)


func _log_v2_status(message: String) -> void:
	ClientLogger.shell_info(message)


func _on_single_player_pressed() -> void:
	_log_v2_status("V2 app entry single player requested")
	main_menu_session_controller.request_single_player()


func _on_multiplayer_create_requested() -> void:
	_log_v2_status("V2 app entry multiplayer create requested")
	main_menu_session_controller.request_create_room()


func _on_multiplayer_join_requested(room_code: String) -> void:
	_log_v2_status("V2 app entry multiplayer join requested: %s" % room_code)
	main_menu_session_controller.request_join_room(room_code)
