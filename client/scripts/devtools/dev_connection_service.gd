extends RefCounted

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")

var connection_service
var operation_trace_factory: Callable


func configure(connection_service_ref, operation_trace_factory_ref: Callable = Callable()) -> void:
	connection_service = connection_service_ref
	operation_trace_factory = operation_trace_factory_ref


func is_configured() -> bool:
	return connection_service != null


func send_spawn_from_placement_result(result: Dictionary, operation_trace: ClientOperationTrace = null) -> void:
	operation_trace = _trace_from_result(result, operation_trace, "devtools.spawn")
	if connection_service == null:
		_emit_dependency_unavailable("connection_service")
		return
	if result.is_empty():
		return
	var packet: Dictionary = DevSpawnPacketBuilder.build_from_placement_result(result)
	if packet.is_empty():
		return
	if !connection_service.has_method("send_packet"):
		_emit_dependency_unavailable("connection_service.send_packet")
		return
	_send(packet, operation_trace)


func send_begin_continuous_bullet_stream_from_placement_result(result: Dictionary, operation_trace: ClientOperationTrace = null) -> void:
	operation_trace = _trace_from_result(result, operation_trace, "devtools.begin_continuous_bullet_stream")
	if connection_service == null:
		_emit_dependency_unavailable("connection_service")
		return
	if result.is_empty():
		return
	var packet: Dictionary = DevSpawnPacketBuilder.build_continuous_bullet_stream_from_placement_result(result)
	if packet.is_empty():
		return
	if !connection_service.has_method("send_packet"):
		_emit_dependency_unavailable("connection_service.send_packet")
		return
	_send(packet, operation_trace)


func send_respawn_player(target_scope: String, target_player_id: String, operation_trace: ClientOperationTrace = null) -> void:
	operation_trace = _ensure_trace(operation_trace, "devtools.respawn_player")
	if connection_service == null:
		_emit_dependency_unavailable("connection_service")
		return
	var packet: Dictionary = DevRespawnPacketBuilder.build(target_scope, target_player_id, operation_trace.trace_id())
	if packet.is_empty():
		return
	if !connection_service.has_method("send_packet"):
		_emit_dependency_unavailable("connection_service.send_packet")
		return
	_send(packet, operation_trace)


func _send(packet: Dictionary, operation_trace: ClientOperationTrace) -> void:
	var trace_id := operation_trace.trace_id()
	packet[DevSpawnPacketBuilder.FIELD_TRACE_ID] = trace_id
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_DEVTOOLS_COMMAND_REQUESTED,
		"",
		{"trace_id": trace_id},
		{"command_type": str(packet.get("type", ""))}
	)
	connection_service.send_packet(packet, trace_id)


func _trace_from_result(result: Dictionary, operation_trace: ClientOperationTrace, operation_name: String) -> ClientOperationTrace:
	if operation_trace != null:
		return operation_trace
	var result_trace = result.get("_operation_trace", null)
	if result_trace is ClientOperationTrace:
		return result_trace
	return _ensure_trace(null, operation_name)


func _ensure_trace(operation_trace: ClientOperationTrace, operation_name: String) -> ClientOperationTrace:
	if operation_trace != null:
		return operation_trace
	return ClientOperationTrace.create(operation_name, operation_trace_factory)


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
