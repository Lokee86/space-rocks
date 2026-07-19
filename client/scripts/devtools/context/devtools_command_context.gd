extends RefCounted
class_name DevtoolsCommandContext

const ClientLogger := preload("res://scripts/logging/logger.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")


var connection_service
var dev_connection_service
var debug_flow
var state_context
var local_respawn_confirmation_marker: Callable
var operation_trace_factory: Callable


func configure(
	debug_flow_ref,
	state_context_ref,
	operation_trace_factory_ref: Callable = Callable()
) -> void:
	debug_flow = debug_flow_ref
	state_context = state_context_ref
	operation_trace_factory = operation_trace_factory_ref


func create_operation_trace(action_name: String) -> ClientOperationTrace:
	return ClientOperationTrace.create(action_name, operation_trace_factory)


func configure_connection(connection_service_ref) -> void:
	connection_service = connection_service_ref


func configure_dev_connection(dev_connection_service_ref) -> void:
	dev_connection_service = dev_connection_service_ref


func configure_local_respawn_confirmation_marker(marker: Callable) -> void:
	local_respawn_confirmation_marker = marker


func process(required_lane_baselines_synced: bool) -> void:
	if debug_flow != null:
		debug_flow.process(required_lane_baselines_synced)


func request_toggle_invincible(target_scope: String = "", target_player_id: String = "") -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.toggle_invincible(target_scope, target_player_id)


func request_toggle_infinite_lives(target_scope: String = "", target_player_id: String = "") -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.toggle_infinite_lives(target_scope, target_player_id)


func request_toggle_freeze_world(freeze_target: String = "") -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.toggle_freeze_world(freeze_target)


func request_toggle_freeze_player(target_scope: String = "", target_player_id: String = "") -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.toggle_freeze_player(target_scope, target_player_id)


func request_clear_bullets() -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.clear_bullets()


func request_clear_asteroids() -> void:
	if !_debug_flow_is_ready():
		return
	debug_flow.clear_asteroids()


func request_kill_player(target_scope: String = "", target_player_id: String = "") -> void:
	if debug_flow == null:
		_emit_dependency_unavailable("debug_flow")
		return
	if target_scope not in ["", DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER]:
		_reject(create_operation_trace("devtools.kill_player"), "debug_kill_player", "invalid_target_scope")
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(create_operation_trace("devtools.kill_player"), "debug_kill_player", "target_required")
		return
	debug_flow.kill_player(target_scope, target_player_id)


func request_set_game_target(target_player_id: String) -> void:
	if state_context == null or !state_context.has_lane_baseline_sync():
		return
	if connection_service == null:
		return
	connection_service.send_packet(Packets.set_target_player_request_packet("player", target_player_id))


func request_clear_game_target() -> void:
	request_set_game_target("")


func request_respawn_player(target_scope: String = DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, target_player_id: String = "") -> void:
	var operation_trace := create_operation_trace("devtools.respawn_player")
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(operation_trace, "debug_respawn_player", "target_required")
		return
	if state_context == null:
		_emit_dependency_unavailable("state_context")
		return
	if !state_context.has_lane_baseline_sync():
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		_emit_dependency_unavailable("dev_connection_service")
		return
	dev_connection_service.send_respawn_player(target_scope, target_player_id, operation_trace)
	var includes_local_player := target_scope == DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS
	if !includes_local_player and state_context != null:
		includes_local_player = target_player_id == state_context.get_local_player_id()
	if includes_local_player:
		if local_respawn_confirmation_marker.is_valid():
			local_respawn_confirmation_marker.call()


func request_respawn_local_player() -> void:
	if state_context == null:
		_emit_dependency_unavailable("state_context")
		return
	if state_context.get_local_player_id() == "":
		_reject(create_operation_trace("devtools.respawn_player"), "debug_respawn_player", "target_required")
		return
	request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, state_context.get_local_player_id())


func request_set_score(target_scope: String, target_player_id: String, score: int) -> void:
	var operation_trace := create_operation_trace("devtools.set_score")
	if !_debug_flow_is_ready():
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(operation_trace, "debug_set_score", "target_required")
		return
	debug_flow.set_score(target_scope, target_player_id, score, operation_trace)


func request_add_score(target_scope: String, target_player_id: String, amount: int) -> void:
	var operation_trace := create_operation_trace("devtools.add_score")
	if !_debug_flow_is_ready():
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(operation_trace, "debug_add_score", "target_required")
		return
	debug_flow.add_score(target_scope, target_player_id, amount, operation_trace)


func request_set_lives(target_scope: String, target_player_id: String, lives: int) -> void:
	var operation_trace := create_operation_trace("devtools.set_lives")
	if !_debug_flow_is_ready():
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(operation_trace, "debug_set_lives", "target_required")
		return
	debug_flow.set_lives(target_scope, target_player_id, lives, operation_trace)


func request_add_lives(target_scope: String, target_player_id: String, amount: int) -> void:
	var operation_trace := create_operation_trace("devtools.add_lives")
	if !_debug_flow_is_ready():
		return
	if target_scope == DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER and target_player_id == "":
		_reject(operation_trace, "debug_add_lives", "target_required")
		return
	debug_flow.add_lives(target_scope, target_player_id, amount, operation_trace)


func _reject(operation_trace: ClientOperationTrace, command_type: String, reason: String) -> void:
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_DEVTOOLS_COMMAND_REJECTED,
		"",
		{"trace_id": operation_trace.trace_id()},
		{"command_type": command_type, "reason": reason}
	)


func _debug_flow_is_ready() -> bool:
	if state_context == null:
		_emit_dependency_unavailable("state_context")
		return false
	if !state_context.has_lane_baseline_sync():
		return false
	if debug_flow == null:
		_emit_dependency_unavailable("debug_flow")
		return false
	return true


func _emit_dependency_unavailable(dependency: String) -> void:
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
		"",
		{},
		{
			"subsystem": "devtools",
			"dependency": dependency,
			"failure_mode": "not_configured",
		}
	)
