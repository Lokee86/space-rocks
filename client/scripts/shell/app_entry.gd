extends Node2D

const SessionBootController := preload("res://scripts/boot/session_boot_controller.gd")
const MainMenuSessionController := preload("res://scripts/main_menu/main_menu_session_controller.gd")

const RoomSessionController := preload("res://scripts/session/room_session_controller.gd")

const ClientConfigController := preload("res://scripts/session/client_config_controller.gd")
const AppShutdownController := preload("res://scripts/session/app_shutdown_controller.gd")
const LocalServerProcessScript := preload("res://scripts/boot/local_server_process.gd")
const LocalAlphaSmokeScript := preload("res://scripts/boot/local_alpha_smoke.gd")
const AuthSessionControllerScript := preload("res://scripts/auth/auth_session_controller.gd")
const AuthApiClientScript := preload("res://scripts/auth/auth_api_client.gd")
const ApiHttpClientScript := preload("res://scripts/api/api_http_client.gd")
const PlayerDataProfileApiClientScript := preload("res://scripts/profile/player_data_profile_api_client.gd")
const ProfileStatsProviderScript := preload("res://scripts/profile/profile_stats_provider.gd")
const MenuFlowControllerScript := preload("res://scripts/ui/menu_flow/menu_flow_controller.gd")
const MultiplayerEntryFlowScript := preload("res://scripts/ui/menu_flow/multiplayer_entry_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const GameplayCameraControllerScript := preload("res://scripts/gameplay/camera/gameplay_camera_controller.gd")
const RuntimeScenarioConfigScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_config.gd")
const RuntimeScenarioDriverScript := preload("res://scripts/devtools/runtime_scenarios/runtime_scenario_driver.gd")

const GAME_SERVER_UNAVAILABLE_MESSAGE := "Could not contact the game server. Please try again."
const GAME_SERVER_AUTH_FAILED_MESSAGE := "Game server authentication failed. Sign in again and retry."

@onready var main_menu: Control = %MainMenu
@onready var user_interface: CanvasLayer = $UserInterface
@onready var gameplay_user_interface: Control = %GameplayUserInterface
@onready var repeated_background: TextureRect = %RepeatedBackground
@onready var repeated_foreground_background: TextureRect = %RepeatedForegroundBackground
@onready var repeated_planet_background: TextureRect = %RepeatedPlanetBackground
@onready var player = $Player
@onready var view_anchor: Node2D = $ViewAnchor
@onready var gameplay_camera: Camera2D = $ViewAnchor/Camera2D
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
var local_server_process
var auth_session_controller
var api_http_client
var player_data_profile_api_client
var profile_stats_provider
var auth_api_client
var background_controller
var menu_flow_controller
var multiplayer_entry_flow
var gameplay_camera_controller
var runtime_scenario_driver

func _ready() -> void:
	var logger_callable := Callable(ClientLogger, "shell_info")
	if !_is_test_process() and !ClientLogger.configure_file_output("user://logs", "client"):
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_OBSERVABILITY_UNAVAILABLE,
			"",
			{},
			{"subsystem": "file_logging", "failure_mode": "configure_failed"}
		)
	ClientLogger.emit_canonical(ObservabilityContract.EVENT_CLIENT_STARTING)

	get_tree().set_auto_accept_quit(false)

	var runtime_scenario_config := RuntimeScenarioConfigScript.from_command_line()
	var runtime_scenario_enabled := bool(runtime_scenario_config.get("enabled", false))
	var local_server_start_error: Error = OK
	if !runtime_scenario_enabled:
		local_server_process = LocalServerProcessScript.new()
		add_child(local_server_process)
		local_server_start_error = local_server_process.start()
	var smoke_phase := _local_alpha_smoke_phase()
	if local_server_start_error != OK:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
			"",
			{},
			{
				"subsystem": "local_packaged_alpha",
				"dependency": "bundled_game_server",
				"failure_mode": "process_start_failed",
				"error_code": str(local_server_start_error),
			}
		)
		if !smoke_phase.is_empty():
			get_tree().quit(2)
			return
	if !smoke_phase.is_empty():
		var smoke = LocalAlphaSmokeScript.new()
		add_child(smoke)
		var smoke_exit_code: int = await smoke.run(smoke_phase)
		local_server_process.stop()
		await get_tree().create_timer(0.1).timeout
		get_tree().quit(smoke_exit_code)
		return

	_setup_boot_and_config(logger_callable)
	_setup_gameplay_camera()

	app_shutdown_controller = AppShutdownController.new()
	add_child(app_shutdown_controller)
	app_shutdown_controller.configure(
		session_boot_controller.get_connection_service(),
		get_tree(),
		local_server_process
	)

	api_http_client = ApiHttpClientScript.new()

	player_data_profile_api_client = PlayerDataProfileApiClientScript.new(api_http_client)
	profile_stats_provider = ProfileStatsProviderScript.new()

	auth_api_client = AuthApiClientScript.new(api_http_client)

	auth_session_controller = AuthSessionControllerScript.new()
	add_child(auth_session_controller)
	auth_session_controller.configure(auth_api_client)
	auth_session_controller.auth_state_changed.connect(_on_auth_state_changed)
	auth_session_controller.auth_error.connect(_on_auth_error)
	session_boot_controller.get_connection_service().set_auth_session_controller(auth_session_controller)
	profile_stats_provider.configure(auth_session_controller, player_data_profile_api_client)

	background_controller = BackgroundController.new()
	add_child(background_controller)
	background_controller.configure(
		repeated_background,
		repeated_foreground_background,
		repeated_planet_background,
		view_anchor,
		gameplay_camera
	)

	gameplay_session_controller = GameplaySessionController.new()
	add_child(gameplay_session_controller)
	gameplay_session_controller.configure(
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_connection_service().get_realtime_packet_pipeline(),
		self,
		player,
		view_anchor,
		bullets,
		asteroids,
		pickups,
		hud,
		gameplay_user_interface,
		main_menu,
		session_boot_controller.get_session_context(),
		session_boot_controller.get_shell_boot_flow(),
		logger_callable,
	)
	gameplay_session_controller.replay_requested.connect(_on_gameplay_replay_requested)
	gameplay_session_controller.return_to_pregame_requested.connect(_on_gameplay_return_to_pregame_requested)

	session_network_controller = SessionNetworkController.new()
	session_network_controller.configure(
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_shell_boot_flow(),
		{}
	)
	session_network_controller.connect_connection_signals()
	session_network_controller.initial_room_operation_failed.connect(_on_initial_room_operation_failed)
	session_network_controller.configure_gameplay_session_controller(gameplay_session_controller)
	session_network_controller.connect_gameplay_signals()

	room_session_controller = RoomSessionController.new()
	room_session_controller.configure(
		main_menu,
		user_interface,
		session_boot_controller.get_session_context(),
		session_boot_controller.get_connection_service(),
		session_boot_controller.get_shell_boot_flow(),
	)
	room_session_controller.configure_client_config_sender(
		Callable(client_config_controller, "send_client_config")
	)

	gameplay_session_controller.configure_room_state_provider(
		Callable(room_session_controller, "current_room_state")
	)
	gameplay_session_controller.configure_match_result_provider(
		Callable(room_session_controller, "current_match_result")
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
		logger_callable,
	)

	menu_flow_controller = MenuFlowControllerScript.new()
	menu_flow_controller.configure(
		user_interface,
		main_menu,
		Callable(self, "_start_single_player_from_pregame"),
		Callable(auth_session_controller, "request_discord_sign_in"),
		Callable(self, "_request_create_room_from_pregame"),
		Callable(self, "_request_join_room_from_pregame"),
		Callable(self, "_logout_from_pregame"),
		auth_session_controller,
		profile_stats_provider
	)
	room_session_controller.configure_lobby_leave_return_destination(
		Callable(menu_flow_controller, "show_multiplayer_pregame")
	)
	room_session_controller.configure_room_transition_completed(
		Callable(menu_flow_controller, "clear_for_room_transition")
	)
	room_session_controller.configure_room_operation_failed(
		Callable(menu_flow_controller, "show_multiplayer_operation_error")
	)

	multiplayer_entry_flow = MultiplayerEntryFlowScript.new()
	multiplayer_entry_flow.configure(menu_flow_controller, auth_session_controller)

	_connect_main_menu_signals()
	_connect_auth_signals()
	if bool(runtime_scenario_config.get("enabled", false)):
		_configure_runtime_scenario_auth(runtime_scenario_config)
	else:
		auth_session_controller.initialize_from_saved_token()
	_make_view_anchor_camera_current()
	ClientLogger.emit_canonical(ObservabilityContract.EVENT_CLIENT_STARTED)
	_start_runtime_scenario(runtime_scenario_config)

func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		if app_shutdown_controller != null:
			app_shutdown_controller.request_shutdown()
		else:
			get_tree().quit()
		ClientLogger.close_file_output()


func _exit_tree() -> void:
	if local_server_process != null:
		local_server_process.stop()
	ClientLogger.close_file_output()


func _is_test_process() -> bool:
	for argument in OS.get_cmdline_args():
		if str(argument).contains("gut_cmdln.gd"):
			return true
	return false


func _local_alpha_smoke_phase() -> String:
	if local_server_process == null || !local_server_process.is_required():
		return ""
	for argument in OS.get_cmdline_user_args():
		if argument.begins_with("--local-alpha-smoke="):
			return argument.trim_prefix("--local-alpha-smoke=")
	return ""


func _configure_runtime_scenario_auth(config: Dictionary) -> void:
	var server_url := str(config.get("server_url", ""))
	if !server_url.is_empty():
		session_boot_controller.set_websocket_url_override(server_url)
	var runtime_client_id := str(config.get("client_id", "runtime-client"))
	var token := "runtime-scenario:%s" % runtime_client_id
	auth_session_controller.initialize_ephemeral_session(token, {
		"id": runtime_client_id,
		"display_name": "Scenario %s" % runtime_client_id,
	})
	session_boot_controller.get_connection_service().set_ephemeral_websocket_auth_token(token)


func _start_runtime_scenario(config: Dictionary) -> void:
	if !bool(config.get("enabled", false)):
		return
	runtime_scenario_driver = RuntimeScenarioDriverScript.new()
	add_child(runtime_scenario_driver)
	runtime_scenario_driver.configure(
		config,
		main_menu_session_controller,
		room_session_controller,
		gameplay_session_controller,
		session_boot_controller.get_connection_service()
	)
	runtime_scenario_driver.start()


func _setup_boot_and_config(logger_callable: Callable) -> void:
	session_boot_controller = SessionBootController.new()
	session_boot_controller.configure(logger_callable)
	add_child(session_boot_controller)

	client_config_controller = ClientConfigController.new()
	client_config_controller.configure(session_boot_controller.get_connection_service(), get_viewport())

	_connect_boot_flow_signal(
		"boot_request_sent",
		Callable(client_config_controller, "send_client_config")
	)


func _setup_gameplay_camera() -> void:
	gameplay_camera_controller = GameplayCameraControllerScript.new()
	add_child(gameplay_camera_controller)
	gameplay_camera_controller.configure(
		gameplay_camera,
		Callable(client_config_controller, "send_client_config"),
		Callable(self, "_is_camera_zoom_active")
	)
	client_config_controller.configure_visible_world_size_provider(
		Callable(gameplay_camera_controller, "visible_world_size")
	)


func _is_camera_zoom_active() -> bool:
	return gameplay_session_controller != null && gameplay_session_controller.is_camera_zoom_active()


func _connect_main_menu_signals() -> void:
	if main_menu == null:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
			"",
			{},
			{
				"subsystem": "app_entry",
				"dependency": "main_menu",
				"failure_mode": "not_configured",
			}
		)
		return

	_connect_main_menu_signal("single_player_requested", _on_single_player_requested)
	if multiplayer_entry_flow != null:
		_connect_main_menu_signal("multiplayer_requested", Callable(multiplayer_entry_flow, "request_multiplayer"))
	_connect_main_menu_signal("logout_requested", _on_logout_requested)


func _connect_auth_signals() -> void:
	if auth_session_controller == null:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CONFIGURATION_INVALID,
			"",
			{},
			{
				"subsystem": "app_entry",
				"configuration_key": "auth_session_controller",
				"reason_code": "missing_required_dependency",
			}
		)
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


func _on_single_player_requested() -> void:
	if menu_flow_controller != null:
		menu_flow_controller.show_single_player_pregame()


func _on_logout_requested() -> void:
	auth_session_controller.logout()


func _on_gameplay_return_to_pregame_requested(session_mode: String) -> void:
	if menu_flow_controller == null:
		return
	if session_mode == Constants.SESSION_MODE_MULTIPLAYER:
		menu_flow_controller.show_multiplayer_pregame()
		return

	menu_flow_controller.show_single_player_pregame()


func _on_gameplay_replay_requested() -> void:
	_start_single_player_from_pregame()


func _start_single_player_from_pregame() -> void:
	var local_profile_id := ""
	if menu_flow_controller != null:
		var single_player_context: Dictionary = menu_flow_controller.get_single_player_context()
		if str(single_player_context.get("identity_kind", "")) == "local_profile":
			local_profile_id = str(single_player_context.get("local_profile_id", ""))
	if menu_flow_controller != null:
		menu_flow_controller.clear_for_gameplay()
	if main_menu_session_controller != null:
		main_menu_session_controller.request_single_player(local_profile_id)


func _request_create_room_from_pregame(config: Dictionary = {}) -> void:
	if main_menu_session_controller != null:
		main_menu_session_controller.request_create_room(config)


func _request_join_room_from_pregame(room_code: String) -> void:
	if main_menu_session_controller != null:
		main_menu_session_controller.request_join_room(room_code)


func _logout_from_pregame() -> void:
	if auth_session_controller != null:
		auth_session_controller.logout()


func _on_auth_state_changed() -> void:
	if main_menu == null || auth_session_controller == null:
		return

	var session = auth_session_controller.get_session()
	if session != null && session.is_signed_in():
		main_menu.show_signed_in(session.display_name)
	else:
		main_menu.show_signed_out()


func _on_auth_error(_message: String) -> void:
	pass


func _on_initial_room_operation_failed(operation: String, error_code: String) -> void:
	if menu_flow_controller == null:
		return
	var message := GAME_SERVER_AUTH_FAILED_MESSAGE if error_code == "authentication_failed" else GAME_SERVER_UNAVAILABLE_MESSAGE
	menu_flow_controller.show_multiplayer_operation_error(operation, message)


func _make_view_anchor_camera_current() -> void:
	if view_anchor == null:
		return
	var camera := view_anchor.get_node_or_null("Camera2D") as Camera2D
	if camera != null:
		camera.make_current()
