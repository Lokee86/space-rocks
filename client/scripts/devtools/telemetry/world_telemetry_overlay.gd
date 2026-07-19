extends CanvasLayer

const UNAVAILABLE := "—"

@onready var metrics_label: Label = %MetricsLabel


func _ready() -> void:
	refresh_metrics({})


func refresh_metrics(metrics: Dictionary) -> void:
	if metrics_label == null:
		return

	var lines := PackedStringArray([
		"World",
		"players: %s" % _preferred_count_value(metrics, "server_players", "players"),
		"enemies: %s" % _preferred_count_value(metrics, "server_enemies", "enemies"),
		"asteroids: %s" % _preferred_count_value(metrics, "server_asteroids", "asteroids"),
		"pickups: %s" % _preferred_count_value(metrics, "server_pickups", "pickups"),
		"projectiles: %s" % _preferred_count_value(metrics, "server_projectiles", "bullets"),
		"asteroids_spawned: %s" % _preferred_count_value(metrics, "server_total_asteroids_spawned", "total_asteroids"),
		"",
		"Client",
		"fps: %s" % _timing_or_network_value(metrics, "fps"),
		"frame_ms: %s" % _timing_or_network_value(metrics, "frame_ms"),
		"",
		"Network",
		"rtt_ms: %s" % _timing_or_network_value(metrics, "rtt_ms"),
		"packet_interval_ms: %s" % _timing_or_network_value(metrics, "packet_interval_ms"),
		"jitter_ms: %s" % _timing_or_network_value(metrics, "jitter_ms"),
		"packet_staleness_ms: %s" % _timing_or_network_value(metrics, "packet_staleness_ms"),
		"packet_age_ms: %s" % _timing_or_network_value(metrics, "packet_age_ms"),
		"packets_in: %s" % _timing_or_network_value(metrics, "packets_in"),
		"packets_out: %s" % _timing_or_network_value(metrics, "packets_out"),
		"bytes_in: %s" % _timing_or_network_value(metrics, "bytes_in"),
		"bytes_out: %s" % _timing_or_network_value(metrics, "bytes_out"),
		"last_in_packet_bytes: %s" % _timing_or_network_value(metrics, "last_in_packet_bytes"),
		"last_out_packet_bytes: %s" % _timing_or_network_value(metrics, "last_out_packet_bytes"),
		"max_in_packet_bytes: %s" % _timing_or_network_value(metrics, "max_in_packet_bytes"),
		"max_out_packet_bytes: %s" % _timing_or_network_value(metrics, "max_out_packet_bytes"),
		"decode_failures: %s" % _timing_or_network_value(metrics, "decode_failures"),
		"encode_failures: %s" % _timing_or_network_value(metrics, "encode_failures"),
		"send_failures: %s" % _timing_or_network_value(metrics, "send_failures"),
	])
	_append_server_lines(lines, metrics)
	_append_measurement_lines(lines, metrics)
	metrics_label.text = "\n".join(lines)


func _append_server_lines(lines: PackedStringArray, metrics: Dictionary) -> void:
	if !metrics.has("server_room_count") and !metrics.has("server_match_id"):
		return
	lines.append("")
	lines.append("Server")
	lines.append("match: %s" % _text_value(metrics, "server_match_id"))
	lines.append("rooms: %s" % _count_value(metrics, "server_room_count"))
	lines.append("player_sessions: %s" % _count_value(metrics, "server_player_sessions"))
	lines.append("radial_effects: %s" % _count_value(metrics, "server_radial_effects"))
	lines.append("heap_allocated_bytes: %s" % _count_value(metrics, "server_heap_allocated_bytes"))
	lines.append("heap_in_use_bytes: %s" % _count_value(metrics, "server_heap_in_use_bytes"))
	lines.append("system_bytes: %s" % _count_value(metrics, "server_system_bytes"))
	lines.append("goroutines: %s" % _count_value(metrics, "server_goroutines"))
	lines.append("gc_cycles: %s" % _count_value(metrics, "server_gc_cycles"))
	lines.append("")
	lines.append("Server Network")
	lines.append("packets_out: %s" % _count_value(metrics, "server_packets_out"))
	lines.append("bytes_out: %s" % _count_value(metrics, "server_bytes_out"))
	lines.append("max_packet_bytes: %s" % _count_value(metrics, "server_max_packet_bytes"))
	_append_server_lane_lines(lines, metrics.get("server_lane_metrics", {}))


func _append_server_lane_lines(lines: PackedStringArray, lane_metrics_value) -> void:
	if !(lane_metrics_value is Dictionary) or lane_metrics_value.is_empty():
		return
	var lane_names: Array = lane_metrics_value.keys()
	lane_names.sort()
	for lane_name in lane_names:
		var lane_metrics = lane_metrics_value[lane_name]
		if !(lane_metrics is Dictionary):
			continue
		lines.append("%s p/b/max: %s/%s/%s" % [
			str(lane_name),
			_count_value(lane_metrics, "packet_count"),
			_count_value(lane_metrics, "encoded_bytes_total"),
			_count_value(lane_metrics, "maximum_encoded_bytes"),
		])


func _preferred_count_value(metrics: Dictionary, preferred_key: String, fallback_key: String) -> String:
	var preferred := _count_value(metrics, preferred_key)
	if preferred != UNAVAILABLE:
		return preferred
	return _count_value(metrics, fallback_key)


func _append_measurement_lines(lines: PackedStringArray, metrics: Dictionary) -> void:
	if !metrics.has("measurement_status"):
		return
	lines.append("")
	lines.append("Measurement")
	lines.append("status: %s" % _text_value(metrics, "measurement_status"))
	lines.append("run: %s" % _text_value(metrics, "measurement_run_id"))
	lines.append("elapsed_s: %s" % _seconds_value(metrics, "measurement_elapsed_ms"))
	lines.append("client_frame_ms avg/max: %s / %s" % [
		_decimal_value(metrics, "measurement_client_frame_average_ms"),
		_decimal_value(metrics, "measurement_client_frame_maximum_ms"),
	])
	lines.append("server_tick_ms avg/max: %s / %s" % [
		_decimal_value(metrics, "measurement_server_tick_average_ms"),
		_decimal_value(metrics, "measurement_server_tick_maximum_ms"),
	])
	lines.append("entities p/a/pr/pk: %s/%s/%s/%s" % [
		_count_value(metrics, "measurement_server_players"),
		_count_value(metrics, "measurement_server_asteroids"),
		_count_value(metrics, "measurement_server_projectiles"),
		_count_value(metrics, "measurement_server_pickups"),
	])
	lines.append("client packets/bytes: %s / %s" % [
		_count_value(metrics, "measurement_client_packets"),
		_count_value(metrics, "measurement_client_bytes"),
	])
	lines.append("server packets/bytes: %s / %s" % [
		_count_value(metrics, "measurement_server_packets"),
		_count_value(metrics, "measurement_server_bytes"),
	])
	lines.append("snapshot_age_ms: %s" % _timing_or_network_value(metrics, "measurement_snapshot_age_ms"))


func _count_value(metrics: Dictionary, key: String) -> String:
	if not metrics.has(key):
		return UNAVAILABLE
	var value: Variant = metrics[key]
	if value is int or value is float:
		if value < 0:
			return UNAVAILABLE
		return str(value)
	return UNAVAILABLE


func _timing_or_network_value(metrics: Dictionary, key: String) -> String:
	if not metrics.has(key):
		return UNAVAILABLE
	var value: Variant = metrics[key]
	if value is int or value is float:
		if value < 0:
			return UNAVAILABLE
		return str(value)
	return UNAVAILABLE


func _decimal_value(metrics: Dictionary, key: String) -> String:
	if !metrics.has(key):
		return UNAVAILABLE
	var value = metrics[key]
	if !(value is int or value is float) or value < 0:
		return UNAVAILABLE
	return "%.2f" % float(value)


func _seconds_value(metrics: Dictionary, key: String) -> String:
	if !metrics.has(key):
		return UNAVAILABLE
	var value = metrics[key]
	if !(value is int or value is float) or value < 0:
		return UNAVAILABLE
	return "%.1f" % (float(value) / 1000.0)


func _text_value(metrics: Dictionary, key: String) -> String:
	var value := str(metrics.get(key, ""))
	return value if !value.is_empty() else UNAVAILABLE
