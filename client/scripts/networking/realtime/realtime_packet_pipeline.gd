extends RefCounted
class_name RealtimePacketPipeline

const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")



const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal gameplay_packet_applied(packet)
signal resync_request_required(lane, baseline_id, sequence, reason)

var _router: RealtimeRouter
var _presentation_state: RealtimePresentationState
var _active_match_id := ""
var _pending_match_packets := {}
var _recovery_match_ids := {}
var _recovery_match_order: Array[String] = []
var _recovery_uncertain := false
var _clock: Callable
var _measurement_observer: Callable

func _init(clock: Callable = Callable(Time, "get_ticks_msec")) -> void:
	_clock = clock
	_router = RealtimeRouter.new()
	_presentation_state = RealtimePresentationState.new()
	_presentation_state.update_from_router(_router)
	_bind_router(_router)

func get_router() -> RealtimeRouter:
	return _router

func active_match_id() -> String:
	return _active_match_id

func begin_match(match_id: String) -> void:
	if match_id.is_empty():
		return
	if _active_match_id == match_id:
		return
	_expire_pending_buckets()
	_active_match_id = match_id
	var pending_bucket: Dictionary = _pending_match_packets.get(match_id, {})
	var has_valid_pending_bucket := pending_bucket.has("packets")
	var requires_recovery := _recovery_match_ids.has(match_id) or _recovery_uncertain
	_pending_match_packets.clear()
	_recovery_match_ids.clear()
	_recovery_match_order.clear()
	_recovery_uncertain = false
	_reset_protocol_state()
	if not has_valid_pending_bucket and requires_recovery:
		_request_recovery_resync()
		return
	for pending_packet in pending_bucket.get("packets", []):
		_apply_lane_packet(pending_packet)

func end_match() -> void:
	_active_match_id = ""
	_pending_match_packets.clear()
	_recovery_match_ids.clear()
	_recovery_match_order.clear()
	_recovery_uncertain = false
	_reset_protocol_state()

func is_gameplay_ready() -> bool:
	if _router == null:
		return false
	return _router.is_presentable()

func apply_packet(packet: Dictionary) -> void:
	if _router == null:
		return
	var expanded_packet: Dictionary = packet
	if not expanded_packet.has("type"):
		if expanded_packet.has("t"):
			expanded_packet = CompactLanePacket.expand_packet(packet)
		else:
			return
	var packet_type = expanded_packet.get("type")
	if DescriptorIndex.packet_by_readable_id(str(packet_type)).is_empty():
		return
	var packet_match_id := str(expanded_packet.get("match_id", ""))
	if packet_match_id.is_empty():
		return
	_expire_pending_buckets()
	if _active_match_id.is_empty():
		if _recovery_match_ids.has(packet_match_id):
			return
		if !_pending_match_packets.has(packet_match_id):
			if _pending_match_packets.size() >= RealtimeReceiveLimits.MAX_PENDING_MATCH_BUCKETS:
				_evict_oldest_pending_bucket()
			_pending_match_packets[packet_match_id] = {"packets": [], "estimated_bytes": 0, "created_at": _now_msec()}
		var bucket: Dictionary = _pending_match_packets[packet_match_id]
		var packet_count: int = bucket["packets"].size() + 1
		var estimated_bytes: int = bucket["estimated_bytes"] + _estimate_packet_bytes(expanded_packet)
		if packet_count > RealtimeReceiveLimits.MAX_PACKETS_PER_MATCH:
			_discard_bucket(packet_match_id, "packet_limit")
			return
		if estimated_bytes > RealtimeReceiveLimits.MAX_ESTIMATED_PACKET_BYTES_PER_MATCH:
			_discard_bucket(packet_match_id, "byte_limit")
			return
		bucket["packets"].append(expanded_packet)
		bucket["estimated_bytes"] = estimated_bytes
		return
	if packet_match_id != _active_match_id:
		return
	_apply_lane_packet(expanded_packet)

func reset() -> void:
	_active_match_id = ""
	_pending_match_packets.clear()
	_recovery_match_ids.clear()
	_recovery_match_order.clear()
	_recovery_uncertain = false
	_reset_protocol_state()

func recover_active_match_baseline() -> void:
	if _active_match_id.is_empty():
		return
	_reset_protocol_state()
	_request_recovery_resync("active_match_baseline_recovery")

func _reset_protocol_state() -> void:
	_router = RealtimeRouter.new()
	_bind_router(_router)
	_presentation_state.update_from_router(_router)

func _request_recovery_resync(reason: String = "pending_bucket_discarded") -> void:
	for lane in ["world", "overlay", "session"]:
		_router.baseline_tracker.request_resync_for_lane(lane, reason)

func _now_msec() -> int:
	return int(_clock.call())

func _estimate_packet_bytes(packet: Dictionary) -> int:
	return JSON.stringify(packet).to_utf8_buffer().size()

func _expire_pending_buckets() -> void:
	var now := _now_msec()
	for match_id in _pending_match_packets.keys():
		var bucket: Dictionary = _pending_match_packets[match_id]
		if now - int(bucket["created_at"]) >= RealtimeReceiveLimits.PENDING_BUCKET_LIFETIME_MSEC:
			_discard_bucket(match_id, "expired")

func _evict_oldest_pending_bucket() -> void:
	var oldest_match_id := ""
	var oldest_created_at := 0
	for match_id in _pending_match_packets.keys():
		var created_at: int = _pending_match_packets[match_id]["created_at"]
		if oldest_match_id.is_empty() or created_at < oldest_created_at:
			oldest_match_id = match_id
			oldest_created_at = created_at
	if not oldest_match_id.is_empty():
		_discard_bucket(oldest_match_id, "evicted")

func _discard_bucket(match_id: String, reason: String) -> void:
	var bucket: Dictionary = _pending_match_packets.get(match_id, {})
	_pending_match_packets.erase(match_id)
	_mark_recovery(match_id)
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_REALTIME_PENDING_STATE_DISCARDED,
		"Realtime pending state discarded",
		{"match_id": match_id},
		{
			"reason": reason,
			"packet_count": bucket.get("packets", []).size(),
			"estimated_bytes": bucket.get("estimated_bytes", 0),
			"recovery_required": true,
		}
	)

func _mark_recovery(match_id: String) -> void:
	if _recovery_match_ids.has(match_id):
		return
	if _recovery_match_order.size() >= RealtimeReceiveLimits.MAX_PENDING_MATCH_BUCKETS:
		var oldest_match_id: String = _recovery_match_order.pop_front()
		_recovery_match_ids.erase(oldest_match_id)
		_recovery_uncertain = true
	_recovery_match_ids[match_id] = true
	_recovery_match_order.append(match_id)

func _bind_router(router: RealtimeRouter) -> void:
	router.resync_request_required.connect(_on_resync_request_required)

func _on_resync_request_required(lane, baseline_id, sequence, reason) -> void:
	resync_request_required.emit(lane, baseline_id, sequence, reason)

func get_presentation_state() -> RealtimePresentationState:
	return _presentation_state


func set_measurement_observer(observer: Callable) -> void:
	_measurement_observer = observer

func _apply_lane_packet(packet: Dictionary) -> void:
	var started_usec := Time.get_ticks_usec()
	_router.route_lane_packet(packet)
	_presentation_state.update_from_router(_router)
	gameplay_packet_applied.emit(packet)
	if _measurement_observer.is_valid():
		_measurement_observer.call(_measurement_lane(packet), float(Time.get_ticks_usec() - started_usec) / 1000.0)


func _measurement_lane(packet: Dictionary) -> String:
	var explicit_lane := str(packet.get("lane", ""))
	if not explicit_lane.is_empty():
		return explicit_lane
	var packet_type := str(packet.get("type", "unknown"))
	if packet_type.begins_with("world") or packet_type.contains("ship") or packet_type.contains("asteroid") or packet_type.contains("bullet"):
		return "world"
	if packet_type.begins_with("overlay"):
		return "overlay"
	if packet_type.begins_with("session"):
		return "session"
	if packet_type.begins_with("event"):
		return "event"
	return packet_type

func apply_world_full(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_world_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_ship_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_ships_lifecycle(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_asteroid_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_bullet_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_asteroids_lifecycle(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_bullets_lifecycle(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_overlay_full(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_overlay_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_session_full(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_session_delta(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_event_batch(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_resync_request(packet: Dictionary) -> void:
	apply_packet(packet)

func apply_resync_required(packet: Dictionary) -> void:
	apply_packet(packet)
