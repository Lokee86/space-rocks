extends RefCounted
class_name DevtoolsCommandContext

const ClientLogger := preload("res://scripts/logging/logger.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

var connection_service
var dev_connection_service
var debug_flow
var state_context


func configure(debug_flow_ref, state_context_ref) -> void:
	debug_flow = debug_flow_ref
	state_context = state_context_ref


func configure_connection(connection_service_ref) -> void:
	connection_service = connection_service_ref


func configure_dev_connection(dev_connection_service_ref) -> void:
	dev_connection_service = dev_connection_service_ref


func process(has_received_state: bool) -> void:
	if debug_flow != null:
		debug_flow.process(has_received_state)


func request_toggle_invincible(target_scope: String = "", target_player_id: String = "") -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.toggle_invincible(target_scope, target_player_id)


func request_toggle_infinite_lives(target_scope: String = "", target_player_id: String = "") -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.toggle_infinite_lives(target_scope, target_player_id)


func request_toggle_freeze_world(freeze_target: String = "") -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.toggle_freeze_world(freeze_target)


func request_toggle_freeze_player(target_scope: String = "", target_player_id: String = "") -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.toggle_freeze_player(target_scope, target_player_id)


func request_clear_bullets() -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.clear_bullets()


func request_clear_asteroids() -> void:
	if state_context == null or !state_context.has_gameplay_state() || debug_flow == null:
		return
	debug_flow.clear_asteroids()


func request_set_game_target(target_player_id: String) -> void:
	if state_context == null or !state_context.has_gameplay_state():
		return
	if connection_service == null:
		return
	connection_service.send_packet(Packets.set_target_player_request_packet(target_player_id))


func request_clear_game_target() -> void:
	request_set_game_target("")


func request_respawn_player(target_scope: String = DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, target_player_id: String = "") -> void:
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, target_player_id is empty")
		return
	if state_context == null or !state_context.has_gameplay_state():
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		ClientLogger.game_warn("GameplayDevtoolsContext: respawn request ignored, dev_connection_service is unavailable")
		return
	dev_connection_service.send_respawn_player(target_scope, target_player_id)


func request_respawn_local_player() -> void:
	if state_context == null or state_context.get_local_player_id() == "":
		ClientLogger.game_warn("GameplayDevtoolsContext: local respawn request ignored, local_player_id is empty")
		return
	request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, state_context.get_local_player_id())


func request_set_score(target_scope: String, target_player_id: String, score: int) -> void:
	if state_context == null or !state_context.has_gameplay_state():
		return
	if debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.set_score(target_scope, target_player_id, score)


func request_add_score(target_scope: String, target_player_id: String, amount: int) -> void:
	if state_context == null or !state_context.has_gameplay_state():
		return
	if debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.add_score(target_scope, target_player_id, amount)


func request_set_lives(target_scope: String, target_player_id: String, lives: int) -> void:
	if state_context == null or !state_context.has_gameplay_state():
		return
	if debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.set_lives(target_scope, target_player_id, lives)


func request_add_lives(target_scope: String, target_player_id: String, amount: int) -> void:
	if state_context == null or !state_context.has_gameplay_state():
		return
	if debug_flow == null:
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		return
	debug_flow.add_lives(target_scope, target_player_id, amount)
