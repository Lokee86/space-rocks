extends RefCounted
class_name GameplayDebugFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const DevtoolsTargetResolverScript = preload("res://scripts/devtools/devtools_target_resolver.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")

var connection_service
var debug_invincible_enabled := false
var debug_invincible_toggle_was_pressed := false
var operation_trace_factory: Callable


func configure(connection_service_ref, operation_trace_factory_ref: Callable = Callable()) -> void:
	connection_service = connection_service_ref
	operation_trace_factory = operation_trace_factory_ref


func create_operation_trace(action_name: String) -> ClientOperationTrace:
	return ClientOperationTrace.create(action_name, operation_trace_factory)


func reset() -> void:
	debug_invincible_enabled = false
	debug_invincible_toggle_was_pressed = false


func process(required_lane_baselines_synced: bool) -> void:
	var toggle_pressed := Input.is_action_just_pressed("DevToggle1")
	var infinite_lives_toggle_pressed := Input.is_action_just_pressed("DevToggle2")
	var world_freeze_toggle_pressed := Input.is_action_just_pressed("DevToggle3")
	var player_freeze_toggle_pressed := Input.is_action_just_pressed("DevToggle4")
	if !required_lane_baselines_synced || connection_service == null:
		debug_invincible_toggle_was_pressed = toggle_pressed
		return

	if toggle_pressed:
		if !debug_invincible_toggle_was_pressed:
			debug_invincible_toggle_was_pressed = true
			toggle_invincible()
	else:
		debug_invincible_toggle_was_pressed = false

	if infinite_lives_toggle_pressed:
		toggle_infinite_lives()

	if world_freeze_toggle_pressed:
		toggle_freeze_world()

	if player_freeze_toggle_pressed:
		toggle_freeze_player()


func toggle_invincible(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	var operation_trace := create_operation_trace("devtools.toggle_invincible")
	if connection_service == null || !connection_service.has_method("send_packet"):
		_emit_dependency_unavailable()
		return
	debug_invincible_enabled = !debug_invincible_enabled
	_send_command(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_INVINCIBLE, target_scope, target_player_id), operation_trace)


func toggle_infinite_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	var operation_trace := create_operation_trace("devtools.toggle_infinite_lives")
	_send_command(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_INFINITE_LIVES, target_scope, target_player_id), operation_trace)


func kill_player(target_scope: String = "", target_player_id: String = "") -> void:
	var operation_trace := create_operation_trace("devtools.kill_player")
	var packet := Packets.debug_kill_player_packet()
	if target_scope != "":
		packet[Packets.FIELD_TARGET_SCOPE] = target_scope
	if target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	_send_command(packet, operation_trace)


func toggle_freeze_world(freeze_target := "") -> void:
	var operation_trace := create_operation_trace("devtools.toggle_freeze_world")
	var packet: Dictionary
	if freeze_target == "" || freeze_target == "all":
		packet = Packets.toggle_debug_freeze_world_packet()
	else:
		packet = Packets.toggle_debug_freeze_world_target_packet(freeze_target)
	_send_command(packet, operation_trace)


func toggle_freeze_player(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	var operation_trace := create_operation_trace("devtools.toggle_freeze_player")
	_send_command(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_FREEZE_PLAYER, target_scope, target_player_id), operation_trace)


func set_score(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	score: int = 0
) -> void:
	var operation_trace := create_operation_trace("devtools.set_score")
	_send_command(_build_counter_packet(Packets.TYPE_DEBUG_SET_SCORE, target_scope, target_player_id, Packets.FIELD_SCORE, score), operation_trace)


func add_score(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	amount: int = 0
) -> void:
	var operation_trace := create_operation_trace("devtools.add_score")
	_send_command(_build_counter_packet(Packets.TYPE_DEBUG_ADD_SCORE, target_scope, target_player_id, Packets.FIELD_AMOUNT, amount), operation_trace)


func set_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	lives: int = 0
) -> void:
	var operation_trace := create_operation_trace("devtools.set_lives")
	_send_command(_build_counter_packet(Packets.TYPE_DEBUG_SET_LIVES, target_scope, target_player_id, Packets.FIELD_LIVES, lives), operation_trace)


func add_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	amount: int = 0
) -> void:
	var operation_trace := create_operation_trace("devtools.add_lives")
	_send_command(_build_counter_packet(Packets.TYPE_DEBUG_ADD_LIVES, target_scope, target_player_id, Packets.FIELD_AMOUNT, amount), operation_trace)


func clear_bullets() -> void:
	var operation_trace := create_operation_trace("devtools.clear_bullets")
	_send_command(Packets.debug_clear_bullets_packet(), operation_trace)


func clear_asteroids() -> void:
	var operation_trace := create_operation_trace("devtools.clear_asteroids")
	_send_command(Packets.debug_clear_asteroids_packet(), operation_trace)


func _send_command(packet: Dictionary, operation_trace: ClientOperationTrace) -> void:
	var trace_id := operation_trace.trace_id()
	packet[Packets.FIELD_TRACE_ID] = trace_id
	var command_type := str(packet.get(Packets.FIELD_TYPE, ""))
	if connection_service == null || !connection_service.has_method("send_packet"):
		_emit_dependency_unavailable()
		return
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_DEVTOOLS_COMMAND_REQUESTED,
		"",
		{"trace_id": trace_id},
		{"command_type": command_type}
	)
	connection_service.send_packet(packet, trace_id)


func _emit_dependency_unavailable() -> void:
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
		"",
		{},
		{
			"subsystem": "devtools",
			"dependency": "connection_service",
			"failure_mode": "not_configured",
		}
	)


func _build_player_toggle_packet(packet_type: String, target_scope: String, target_player_id: String) -> Dictionary:
	var packet := {
		Packets.FIELD_TYPE: packet_type,
		Packets.FIELD_TARGET_SCOPE: target_scope,
	}
	if target_scope == DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER and target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet


func _build_counter_packet(
	packet_type: String,
	target_scope: String,
	target_player_id: String,
	value_field: String,
	value: int
) -> Dictionary:
	var packet := {
		Packets.FIELD_TYPE: packet_type,
		Packets.FIELD_TARGET_SCOPE: target_scope,
		value_field: value,
	}
	if target_scope == DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER and target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet
