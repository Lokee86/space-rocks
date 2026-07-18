extends RefCounted
class_name ClientMeasurementOverlaySnapshot

const NANOSECONDS_PER_MILLISECOND := 1000000.0
const UNAVAILABLE := -1

var coordinator
var client_measurement_context
var _last_server_snapshot: Dictionary = {}
var _last_server_snapshot_msec := -1


func configure(coordinator_ref, client_measurement_context_ref) -> void:
	coordinator = coordinator_ref
	client_measurement_context = client_measurement_context_ref
	reset()


func reset() -> void:
	_last_server_snapshot.clear()
	_last_server_snapshot_msec = -1


func snapshot(now_msec: int = -1) -> Dictionary:
	var now := now_msec if now_msec >= 0 else Time.get_ticks_msec()
	var state := _coordinator_state()
	var client_snapshot := _client_snapshot()
	var server_snapshot := _dictionary(state.get("latest_server_snapshot", {}))
	_track_server_snapshot(
		server_snapshot,
		now,
		int(state.get("latest_server_snapshot_received_at_msec", -1))
	)

	var pending_request_ids := _dictionary(state.get("pending_request_ids", {}))
	var recording := bool(state.get("recording", false))
	var status := "idle"
	if pending_request_ids.has("start"):
		status = "starting"
	elif pending_request_ids.has("stop"):
		status = "stopping"
	elif recording:
		status = "recording"
	elif !_dictionary(state.get("last_tooling_error", {})).is_empty():
		status = "error"

	var frame_timing := _dictionary(client_snapshot.get("frame_timing", {}))
	var server_ticks := _dictionary(server_snapshot.get("ticks", {}))
	var entities := _latest_server_entities(server_snapshot)
	var server_packet_totals := _server_packet_totals(server_snapshot)
	var client_network := _dictionary(client_snapshot.get("network_metrics", {}))

	return {
		"measurement_status": status,
		"measurement_run_id": str(state.get("active_run_id", "")),
		"measurement_elapsed_ms": _number(client_snapshot, "duration"),
		"measurement_client_frame_average_ms": _number(frame_timing, "average"),
		"measurement_client_frame_maximum_ms": _number(frame_timing, "maximum"),
		"measurement_server_tick_average_ms": _duration_milliseconds(server_ticks.get("average", null)),
		"measurement_server_tick_maximum_ms": _duration_milliseconds(server_ticks.get("maximum", null)),
		"measurement_server_players": _number(entities, "players"),
		"measurement_server_asteroids": _number(entities, "asteroids"),
		"measurement_server_projectiles": _number(entities, "projectiles"),
		"measurement_server_pickups": _number(entities, "pickups"),
		"measurement_client_packets": _sum_numbers(client_network, ["packets_in", "packets_out"]),
		"measurement_client_bytes": _sum_numbers(client_network, ["bytes_in", "bytes_out"]),
		"measurement_server_packets": server_packet_totals["packets"],
		"measurement_server_bytes": server_packet_totals["bytes"],
		"measurement_snapshot_age_ms": now - _last_server_snapshot_msec if _last_server_snapshot_msec >= 0 else UNAVAILABLE,
	}


func _coordinator_state() -> Dictionary:
	if coordinator != null and coordinator.has_method("get_state"):
		var result = coordinator.get_state()
		if result is Dictionary:
			return result.duplicate(true)
	return {}


func _client_snapshot() -> Dictionary:
	if client_measurement_context != null and client_measurement_context.has_method("snapshot"):
		var result = client_measurement_context.snapshot()
		if result is Dictionary:
			return result.duplicate(true)
	return {}


func _track_server_snapshot(server_snapshot: Dictionary, now_msec: int, received_at_msec: int) -> void:
	if server_snapshot.is_empty():
		_last_server_snapshot.clear()
		_last_server_snapshot_msec = -1
		return
	if received_at_msec >= 0:
		_last_server_snapshot = server_snapshot.duplicate(true)
		_last_server_snapshot_msec = received_at_msec
		return
	if server_snapshot != _last_server_snapshot:
		_last_server_snapshot = server_snapshot.duplicate(true)
		_last_server_snapshot_msec = now_msec


func _latest_server_entities(server_snapshot: Dictionary) -> Dictionary:
	var samples = server_snapshot.get("samples", [])
	if !(samples is Array) or samples.is_empty():
		return {}
	var latest = samples[samples.size() - 1]
	if !(latest is Dictionary):
		return {}
	return _dictionary(latest.get("entities", {}))


func _server_packet_totals(server_snapshot: Dictionary) -> Dictionary:
	if server_snapshot.is_empty():
		return {"packets": UNAVAILABLE, "bytes": UNAVAILABLE}
	var packet_count := 0
	var encoded_bytes := 0
	var packets = server_snapshot.get("packets", [])
	if packets is Array:
		for packet in packets:
			if !(packet is Dictionary):
				continue
			packet_count += int(packet.get("packet_count", 0))
			encoded_bytes += int(packet.get("encoded_bytes_total", 0))
	return {"packets": packet_count, "bytes": encoded_bytes}


func _sum_numbers(values: Dictionary, keys: Array) -> int:
	var total := 0
	var found := false
	for key in keys:
		var value = values.get(key, null)
		if value is int or value is float:
			total += int(value)
			found = true
	return total if found else UNAVAILABLE


func _duration_milliseconds(value) -> float:
	if value is int or value is float:
		return float(value) / NANOSECONDS_PER_MILLISECOND
	return UNAVAILABLE


func _number(values: Dictionary, key: String):
	var value = values.get(key, null)
	if value is int or value is float:
		return value
	return UNAVAILABLE


func _dictionary(value) -> Dictionary:
	return value if value is Dictionary else {}
