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
		"players: %s" % _count_value(metrics, "players"),
		"enemies: %s" % _count_value(metrics, "enemies"),
		"asteroids: %s" % _count_value(metrics, "asteroids"),
		"pickups: %s" % _count_value(metrics, "pickups"),
		"total_asteroids: %s" % _count_value(metrics, "total_asteroids"),
		"bullets: %s" % _count_value(metrics, "bullets"),
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
	_append_measurement_lines(lines, metrics)
	metrics_label.text = "\n".join(lines)


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
