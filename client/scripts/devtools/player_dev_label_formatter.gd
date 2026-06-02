extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")
const MISSING_VALUE := "—"


static func basic_player_text(player_id: String, state: Dictionary) -> String:
	var lines := [
		"ID: %s" % _short_player_id(player_id),
		"Score: %s" % _state_value_text(state, Packets.FIELD_SCORE),
		"Lives: %s" % _state_value_text(state, Packets.FIELD_LIVES),
		"Ship: %s" % _state_value_text(state, Packets.FIELD_SHIP_TYPE),
		"X: %s" % _rounded_state_value_text(state, Packets.FIELD_X),
		"Y: %s" % _rounded_state_value_text(state, Packets.FIELD_Y),
	]
	return "\n".join(lines)


static func network_text(metrics: Dictionary) -> String:
	var lines := [
		"rtt_ms: %s" % _non_negative_metric_text(metrics, "rtt_ms"),
		"packet_interval_ms: %s" % _non_negative_metric_text(metrics, "packet_interval_ms"),
		"jitter_ms: %s" % _non_negative_metric_text(metrics, "jitter_ms"),
		"packet_staleness_ms: %s" % _non_negative_metric_text(metrics, "packet_staleness_ms"),
		"packet_age_ms: %s" % _non_negative_metric_text(metrics, "packet_age_ms"),
	]
	return "\n".join(lines)


static func _short_player_id(player_id: String) -> String:
	if player_id.length() > 8:
		return "%s…" % player_id.substr(0, 8)
	return player_id


static func _state_value_text(state: Dictionary, field_name: String) -> String:
	if not state.has(field_name):
		return MISSING_VALUE
	return str(state[field_name])


static func _rounded_state_value_text(state: Dictionary, field_name: String) -> String:
	if not state.has(field_name):
		return MISSING_VALUE
	return str(int(round(float(state[field_name]))))


static func _non_negative_metric_text(metrics: Dictionary, field_name: String) -> String:
	if not metrics.has(field_name):
		return MISSING_VALUE
	var value = metrics[field_name]
	if value is int or value is float:
		if value < 0:
			return MISSING_VALUE
	return str(value)
