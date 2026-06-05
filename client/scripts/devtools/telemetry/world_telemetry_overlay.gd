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
	])
	metrics_label.text = "\n".join(lines)


func _count_value(metrics: Dictionary, key: String) -> String:
	if not metrics.has(key):
		return UNAVAILABLE
	var value: Variant = metrics[key]
	if value is int or value is float:
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
