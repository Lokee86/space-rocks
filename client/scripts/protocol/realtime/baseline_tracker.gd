extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")
const LaneSyncState = preload("res://scripts/protocol/realtime/lane_sync_state.gd")

const LANES := [LaneMetadata.LANE_WORLD, LaneMetadata.LANE_OVERLAY, LaneMetadata.LANE_SESSION]

signal resync_required(lane, baseline_id, last_accepted_sequence, reason)

var _lane_state := {}
var _readiness = null

func _init() -> void:
	for lane in LANES:
		_lane_state[lane] = _make_lane_state()

func bind_readiness(readiness) -> void:
	_readiness = readiness
	_refresh_readiness()

func get_lane_state(lane: String) -> Dictionary:
	return _lane_state.get(lane, _make_lane_state()).duplicate(true)

func is_lane_synced(lane: String) -> bool:
	return _ensure_lane_state(lane).sync_state == LaneSyncState.SYNCED

func get_active_baseline_id(lane: String) -> String:
	var value = _ensure_lane_state(lane).baseline_id
	return value if value is String else ""

func get_last_accepted_sequence(lane: String):
	return _ensure_lane_state(lane).sequence

func record_full_packet(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index: int = 0, chunk_count: int = 1, is_final_chunk: bool = true) -> bool:
	var state := _ensure_lane_state(lane)
	if not _valid_chunk_metadata(chunk_index, chunk_count, is_final_chunk):
		return false
	if not _accept_full_start(state, baseline_id, sequence, chunk_index):
		return false
	_apply_full_chunk(state, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, is_final_chunk)
	_refresh_readiness_for_lane(lane)
	return true

func record_full_chunk(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index: int = 0, chunk_count: int = 1, is_final_chunk: bool = false) -> bool:
	var state := _ensure_lane_state(lane)
	if not _valid_chunk_metadata(chunk_index, chunk_count, is_final_chunk):
		return false
	if not state.pending_full:
		if not _accept_full_start(state, baseline_id, sequence, chunk_index):
			return false
	else:
		if baseline_id != state.baseline_id:
			return false
		if sequence != state.sequence:
			return false
		if chunk_count != state.chunk_count:
			return false
		if chunk_index != state.chunk_index + 1:
			return false
	_apply_full_chunk(state, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, is_final_chunk)
	_refresh_readiness_for_lane(lane)
	return true

func record_delta(lane: String, baseline_id, sequence, snapshot_id = null) -> bool:
	var state := _ensure_lane_state(lane)
	if not state.full_packet_seen or state.baseline_id == null or str(state.baseline_id).is_empty():
		_request_resync(lane, state, "missing_baseline")
		return false
	if state.sync_state == LaneSyncState.UNSYNCED or _is_stale_sequence(state, sequence):
		return false
	if state.baseline_id != baseline_id:
		_request_resync(lane, state, "wrong_baseline")
		return false
	state.sequence = sequence
	state.snapshot_id = snapshot_id
	return true

func mark_lane_unsynced(lane: String) -> void:
	var state := _ensure_lane_state(lane)
	state.sync_state = LaneSyncState.UNSYNCED
	state.resync_needed = true
	_refresh_readiness_for_lane(lane)

func needs_resync(lane: String) -> bool:
	return _ensure_lane_state(lane).resync_needed

func request_resync_for_lane(lane: String, reason: String) -> void:
	_request_resync(lane, _ensure_lane_state(lane), reason)

func _ensure_lane_state(lane: String) -> Dictionary:
	if not _lane_state.has(lane):
		_lane_state[lane] = _make_lane_state()
	return _lane_state[lane]

func _make_lane_state() -> Dictionary:
	return {
		"baseline_id": null,
		"snapshot_id": null,
		"sequence": null,
		"chunk_index": -1,
		"chunk_count": 0,
		"is_final_chunk": false,
		"sync_state": LaneSyncState.UNSYNCED,
		"resync_needed": false,
		"pending_full": false,
		"full_packet_seen": false,
	}

func _is_stale_sequence(state: Dictionary, sequence) -> bool:
	return state.sequence != null and sequence != null and sequence <= state.sequence

func _valid_chunk_metadata(chunk_index: int, chunk_count: int, is_final_chunk: bool) -> bool:
	if chunk_count < 1 or chunk_index < 0 or chunk_index >= chunk_count:
		return false
	return is_final_chunk == (chunk_index == chunk_count - 1)

func _accept_full_start(state: Dictionary, baseline_id, sequence, chunk_index: int) -> bool:
	if sequence == null or chunk_index != 0 or state.pending_full:
		return false
	if state.baseline_id != null and baseline_id != state.baseline_id:
		return state.resync_needed
	return state.sequence == null or sequence > state.sequence

func _apply_full_chunk(state: Dictionary, baseline_id, sequence, snapshot_id, chunk_index: int, chunk_count: int, is_final_chunk: bool) -> void:
	state.baseline_id = baseline_id
	state.snapshot_id = snapshot_id
	state.sequence = sequence
	state.chunk_index = chunk_index
	state.chunk_count = chunk_count
	state.is_final_chunk = is_final_chunk
	state.pending_full = not is_final_chunk
	state.full_packet_seen = true
	state.sync_state = LaneSyncState.SYNCED if is_final_chunk else LaneSyncState.SYNCING
	if is_final_chunk:
		state.resync_needed = false

func _request_resync(lane: String, state: Dictionary, reason: String) -> void:
	var was_pending: bool = state.resync_needed
	var previous_baseline = state.baseline_id if state.baseline_id is String else ""
	var previous_sequence = state.sequence
	state.sync_state = LaneSyncState.UNSYNCED
	state.resync_needed = true
	state.pending_full = false
	_refresh_readiness_for_lane(lane)
	if not was_pending:
		resync_required.emit(lane, previous_baseline, previous_sequence, reason)

func _refresh_readiness() -> void:
	for lane in LANES:
		_refresh_readiness_for_lane(lane)

func _refresh_readiness_for_lane(lane: String) -> void:
	if _readiness != null and _readiness.has_method("set_lane_synced"):
		_readiness.set_lane_synced(lane, is_lane_synced(lane))
