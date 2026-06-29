extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")

var _lane_synced := {
	LaneMetadata.LANE_WORLD: false,
	LaneMetadata.LANE_OVERLAY: false,
	LaneMetadata.LANE_SESSION: false,
}

func set_lane_synced(lane: String, synced: bool) -> void:
	if not _lane_synced.has(lane):
		return
	_lane_synced[lane] = synced

func mark_world_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_WORLD, true)

func mark_overlay_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_OVERLAY, true)

func mark_session_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_SESSION, true)

func clear_world_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_WORLD, false)

func clear_overlay_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_OVERLAY, false)

func clear_session_baseline_synced() -> void:
	set_lane_synced(LaneMetadata.LANE_SESSION, false)

func is_gameplay_ready() -> bool:
	return _lane_synced[LaneMetadata.LANE_WORLD] and _lane_synced[LaneMetadata.LANE_OVERLAY] and _lane_synced[LaneMetadata.LANE_SESSION]
