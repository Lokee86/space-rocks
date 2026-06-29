extends RefCounted

const REASON_WRONG_BASELINE := "wrong_baseline"
const REASON_MISSING_BASELINE := "missing_baseline"
const REASON_STALE_OR_INVALID_SEQUENCE := "stale_or_invalid_sequence"

const LANES := ["world", "overlay", "session"]

var _lane_resync := {}

func _init() -> void:
	for lane in LANES:
		_lane_resync[lane] = _make_lane_resync()

func needs_resync(lane: String) -> bool:
	return _ensure_lane_resync(lane).needed

func get_reason(lane: String):
	return _ensure_lane_resync(lane).reason

func mark_wrong_baseline(lane: String) -> void:
	_mark_lane_resync(lane, REASON_WRONG_BASELINE)

func mark_missing_baseline(lane: String) -> void:
	_mark_lane_resync(lane, REASON_MISSING_BASELINE)

func mark_stale_or_invalid_sequence(lane: String) -> void:
	_mark_lane_resync(lane, REASON_STALE_OR_INVALID_SEQUENCE)

func clear_resync(lane: String) -> void:
	var lane_resync := _ensure_lane_resync(lane)
	lane_resync.needed = false
	lane_resync.reason = null

func clear_all() -> void:
	for lane in LANES:
		clear_resync(lane)

func _mark_lane_resync(lane: String, reason: String) -> void:
	var lane_resync := _ensure_lane_resync(lane)
	lane_resync.needed = true
	lane_resync.reason = reason

func _ensure_lane_resync(lane: String) -> Dictionary:
	if not _lane_resync.has(lane):
		_lane_resync[lane] = _make_lane_resync()
	return _lane_resync[lane]

func _make_lane_resync() -> Dictionary:
	return {
		"needed": false,
		"reason": null,
	}

