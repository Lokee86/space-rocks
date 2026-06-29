extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")
const GameplayReadModel = preload("res://scripts/protocol/realtime/gameplay_read_model.gd")
const ResyncState = preload("res://scripts/protocol/realtime/resync_state.gd")

var _read_model := GameplayReadModel.new()
var _resync_state := ResyncState.new()
var _presented_once := false

func is_presentable() -> bool:
	return not _required_lanes_need_resync()

func build_compatibility_read_model(world_lane_state, overlay_lane_state, session_lane_state, event_batch_applier = null) -> Dictionary:
	return _read_model.build_gameplay_read_model(world_lane_state, overlay_lane_state, session_lane_state, event_batch_applier)

func can_fanout() -> bool:
	return is_presentable()

func mark_lane_synced(lane: String) -> void:
	_resync_state.clear_resync(lane)

func mark_lane_needs_resync(lane: String, reason: String = ResyncState.REASON_WRONG_BASELINE) -> void:
	match reason:
		ResyncState.REASON_MISSING_BASELINE:
			_resync_state.mark_missing_baseline(lane)
		ResyncState.REASON_STALE_OR_INVALID_SEQUENCE:
			_resync_state.mark_stale_or_invalid_sequence(lane)
		_:
			_resync_state.mark_wrong_baseline(lane)

func has_fanned_out() -> bool:
	return _presented_once

func mark_fanned_out() -> void:
	_presented_once = true

func _required_lanes_need_resync() -> bool:
	return _resync_state.needs_resync(LaneMetadata.LANE_WORLD) or _resync_state.needs_resync(LaneMetadata.LANE_OVERLAY) or _resync_state.needs_resync(LaneMetadata.LANE_SESSION)

