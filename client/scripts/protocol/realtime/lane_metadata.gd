extends RefCounted

const LANE_WORLD := "world"
const LANE_OVERLAY := "overlay"
const LANE_SESSION := "session"
const LANE_EVENT := "event"
const LANE_CONTROL := "control"
const LANE_DEBUG := "debug"
const LANE_TELEMETRY := "telemetry"

const PACKET_FAMILY_WORLD := ["world_full", "world_delta"]
const PACKET_FAMILY_OVERLAY := ["overlay_full", "overlay_delta"]
const PACKET_FAMILY_SESSION := ["session_full", "session_delta"]
const PACKET_FAMILY_EVENT := ["event_batch"]
const PACKET_FAMILY_CONTROL := ["resync_request", "resync_required"]

const PACKET_TYPE_TO_LANE := {
	"world_full": LANE_WORLD,
	"world_delta": LANE_WORLD,
	"overlay_full": LANE_OVERLAY,
	"overlay_delta": LANE_OVERLAY,
	"session_full": LANE_SESSION,
	"session_delta": LANE_SESSION,
	"event_batch": LANE_EVENT,
	"resync_request": LANE_CONTROL,
	"resync_required": LANE_CONTROL,
}

