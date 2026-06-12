extends Node2D

const SessionBootController := preload("res://scripts/boot/session_boot_controller.gd")
const MainMenuSessionController := preload("res://scripts/main_menu/main_menu_session_controller.gd")
const SessionNetworkController := preload("res://scripts/session/session_network_controller.gd")
const RoomSessionController := preload("res://scripts/session/room_session_controller.gd")
const GameplaySessionController := preload("res://scripts/session/gameplay_session_controller.gd")
const ClientConfigController := preload("res://scripts/session/client_config_controller.gd")
const AppShutdownController := preload("res://scripts/session/app_shutdown_controller.gd")
const AuthSessionController := preload("res://scripts/auth/auth_session_controller.gd")
const AuthApiClient := preload("res://scripts/auth/auth_api_client.gd")
const ApiHttpClient := preload("res://scripts/api/api_http_client.gd")
const MenuFlowController := preload("res://scripts/ui/menu_flow/menu_flow_controller.gd")
const MultiplayerEntryFlow := preload("res://scripts/ui/menu_flow/multiplayer_entry_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

@onready var main_menu: Control = %MainMenu
@onready var canvas_layer: CanvasLayer = $CanvasLayer
@onready var repeated_background: TextureRect = %RepeatedBackground
@onready var repeated_foreground_background: TextureRect = %RepeatedForegroundBackground
@onready var repeated_planet_background: TextureRect = %RepeatedPlanetBackground
@onready var player = $Player
@onready var view_anchor: Node2D = $ViewAnchor
@onready var bullets: Node2D = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var pickups: Node2D = $Pickups
@onready var hud: Control = %HUD

var session_boot_controller
var main_menu_session_controller
var session_network_controller
var room_session_controller
var gameplay_session_controller
var client_config_controller
var app_shutdown_controller
var auth_session_controller
var api_http_client
var auth_api_client
var background_controller
var menu_flow_controller
var multiplayer_entry_flow

func _ready() -> void:

	_log_shell_status("App entry booted")
	get_tree().set_auto_accept_quit(false)

	_setup_boot_and_config()

	app_shutdown_controller = AppShutdownController.new()
	add_child(app_shutdown_controller)
	app_shutdown_controller.configure(session_boot_controller.get_connection_service(), get_tree())

	api_http_client = ApiHttpClient.new()
	add_child(api_http_client)

	auth_api_client = AuthApiClient.new(api_http_client)

	auth_session_controller = AuthSessionController.new()
	add_child(auth_session_controller)
	auth_session_controller.configure(auth_api_client)
	auth_session_controller.auth_state_changed.connect(_on_auth_state_changed)
	auth_session_controller.auth_error.connect(_on_auth_error)
	session_boot_controller.get_connection_service().set_auth_session_controller(auth_session_controller)

	background_controller = BackgroundController.new()
	add_child(background_controller)
	background_controller.configure(repeated_background, repeated_foreground_background, repeated_planet_background, view_anchor)

	gameplay_session_controller = GameplaySessionController.new()
	add_child(gameplay_session_controller)
	gameplay_session_controller.configure(
		session_boot_controller.get_connection_service(),
		self,
		player,
		view_anchor,
		bullets,
		asteroids,
		pickups,
		hud,
		main_menu,
		session_boot_controller.get_session_context(),
		session_boot_controller.get_shell_boot_flow(),
		Callable(self, "_log_shell_status")
	)

	session_network_controller = SessionNetworkController.new()
	session_network_controller.configure(
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_shell_boot_flow(),
		Callable(self, "_log_shell_status"),
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
		Callable(self, "_log_shell_status")
	)
	room_session_controller.configure_client_config_sender(
		Callable(client_config_controller, "send_client_config")
	)

	gameplay_session_controller.configure_room_state_provider(
		Callable(room_session_controller, "current_room_state")
	)
	gameplay_session_controller.configure_room_max_players_provider(
		Callable(room_session_controller, "current_max_players")
	)

	session_network_controller.configure_room_session_controller(room_session_controller)
	session_network_controller.connect_room_signals()

	main_menu_session_controller = MainMenuSessionController.new()
	main_menu_session_controller.configure(
		main_menu,
		session_boot_controller,
		Callable(self, "_log_shell_status")
	)

	menu_flow_controller = MenuFlowController.new()
	menu_flow_controller.configure(
		canvas_layer,
		main_menu,
		Callable(self, "_start_single_player_from_pregame"),
		Callable(auth_session_controller, "request_discord_sign_in")
	)

	multiplayer_entry_flow = MultiplayerEntryFlow.new()
	multiplayer_entry_flow.configure(menu_flow_controller, auth_session_controller)

	_connect_main_menu_signals()
	_connect_auth_signals()
	auth_session_controller.initialize_from_saved_token()
	_make_view_anchor_camera_current()

func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		if app_shutdown_controller != null:
			app_shutdown_controller.request_shutdown()
		else:
			get_tree().quit()


func _setup_boot_and_config() -> void:
	session_boot_controller = SessionBootController.new()
	session_boot_controller.configure(Constants.MULTIPLAYER_WS_URL, Callable(self, "_log_shell_status"))
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

	_connect_main_menu_signal("single_player_requested", _on_single_player_requested)
	if multiplayer_entry_flow != null:
		_connect_main_menu_signal("multiplayer_requested", Callable(multiplayer_entry_flow, "request_multiplayer"))
	_connect_main_menu_signal("logout_requested", _on_logout_requested)


func _connect_auth_signals() -> void:
	if auth_session_controller == null:
		push_error("Missing auth session controller")
		return

	if !auth_session_controller.auth_state_changed.is_connected(_on_auth_state_changed):
		auth_session_controller.auth_state_changed.connect(_on_auth_state_changed)
	if multiplayer_entry_flow != null:
		var multiplayer_state_callable := Callable(multiplayer_entry_flow, "handle_auth_state_changed")
		if !auth_session_controller.auth_state_changed.is_connected(multiplayer_state_callable):
			auth_session_controller.auth_state_changed.connect(multiplayer_state_callable)
	if !auth_session_controller.auth_error.is_connected(_on_auth_error):
		auth_session_controller.auth_error.connect(_on_auth_error)


func _connect_main_menu_signal(signal_name: StringName, handler: Callable) -> void:
	if main_menu.has_signal(signal_name):
		main_menu.connect(signal_name, handler)


func _connect_boot_flow_signal(signal_name: StringName, handler: Callable) -> void:
	var shell_boot_flow = session_boot_controller.get_shell_boot_flow()
	if shell_boot_flow.has_signal(signal_name) && !shell_boot_flow.is_connected(signal_name, handler):
		shell_boot_flow.connect(signal_name, handler)


func _log_shell_status(message: String) -> void:
	ClientLogger.shell_info(message)


func _on_single_player_requested() -> void:
	_log_shell_status("App entry single-player pregame requested")
	if menu_flow_controller != null:
		menu_flow_controller.show_single_player_pregame()


func _on_logout_requested() -> void:
	_log_shell_status("App entry logout requested")
	auth_session_controller.logout()


func _start_single_player_from_pregame() -> void:
	if menu_flow_controller != null:
		menu_flow_controller.clear_for_gameplay()
	if main_menu_session_controller != null:
		main_menu_session_controller.request_single_player()


func _on_auth_state_changed() -> void:
	if main_menu == null || auth_session_controller == null:
		return

	var session = auth_session_controller.get_session()
	if session != null && session.is_signed_in():
		main_menu.show_signed_in(session.display_name)
	else:
		main_menu.show_signed_out()


func _on_auth_error(message: String) -> void:
	ClientLogger.shell_info("Auth error: %s" % message)


func _make_view_anchor_camera_current() -> void:
	if view_anchor == null:
		return
	var camera := view_anchor.get_node_or_null("Camera2D") as Camera2D
	if camera != null:
		camera.make_current()
