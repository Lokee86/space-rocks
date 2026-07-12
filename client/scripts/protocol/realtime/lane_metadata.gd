extends RefCounted

const LANE_WORLD := "world"
const LANE_ASTEROIDS := "asteroids"
const LANE_ASTEROIDS_LIFECYCLE := "asteroids.lifecycle"
const LANE_BULLETS := "bullets"
const LANE_BULLETS_LIFECYCLE := "bullets.lifecycle"
const LANE_OVERLAY := "overlay"
const LANE_SESSION := "session"
const LANE_EVENT := "event"
const LANE_CONTROL := "control"
const LANE_DEBUG := "debug"
const LANE_TELEMETRY := "telemetry"

static func parse_non_negative_integer(value):
	var value_type := typeof(value)
	if value_type == TYPE_BOOL or value == null:
		return null
	if value_type == TYPE_INT:
		return value if value >= 0 else null
	if value_type != TYPE_FLOAT:
		return null
	var numeric_value := float(value)
	if not is_finite(numeric_value) or numeric_value < 0.0 or floor(numeric_value) != numeric_value:
		return null
	return int(numeric_value)

static func is_valid_id(value) -> bool:
	return value is String and not value.is_empty()

static func is_supported_baseline_lane(lane: String) -> bool:
	return lane in [LANE_WORLD, LANE_OVERLAY, LANE_SESSION]
