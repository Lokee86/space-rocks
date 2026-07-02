extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")
const LaneSyncState = preload("res://scripts/protocol/realtime/lane_sync_state.gd")

const LANES := [LaneMetadata.LANE_WORLD, LaneMetadata.LANE_OVERLAY, LaneMetadata.LANE_SESSION]

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

func record_full_packet(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index: int = 0, chunk_count: int = 1, is_final_chunk: bool = true) -> void:
	var state := _ensure_lane_state(lane)
	state.baseline_id = baseline_id
	state.snapshot_id = snapshot_id
	state.sequence = sequence
	state.chunk_index = chunk_index
	state.chunk_count = chunk_count
	state.is_final_chunk = is_final_chunk
	state.sync_state = LaneSyncState.SYNCING if not is_final_chunk else LaneSyncState.SYNCED
	state.resync_needed = false
	state.pending_full = not is_final_chunk
	state.full_packet_seen = true
	_refresh_readiness_for_lane(lane)

func record_full_chunk(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index: int = 0, chunk_count: int = 1, is_final_chunk: bool = false) -> void:
	var state := _ensure_lane_state(lane)
	if state.full_packet_seen and state.baseline_id != baseline_id:
		_mark_resync_needed(state)
		_refresh_readiness_for_lane(lane)
		return
	if _is_stale_sequence(state, sequence):
		return
	state.baseline_id = baseline_id
	state.snapshot_id = snapshot_id
	state.sequence = sequence
	state.chunk_index = chunk_index
	state.chunk_count = chunk_count
	state.is_final_chunk = is_final_chunk
	state.pending_full = not is_final_chunk
	state.full_packet_seen = true
	if is_final_chunk:
		state.sync_state = LaneSyncState.SYNCED
		state.resync_needed = false
	else:
		state.sync_state = LaneSyncState.SYNCING
	_refresh_readiness_for_lane(lane)

func record_delta(lane: String, baseline_id, sequence, snapshot_id = null) -> bool:
	var state := _ensure_lane_state(lane)
	if state.sync_state == LaneSyncState.UNSYNCED:
		return false
	if _is_stale_sequence(state, sequence):
		return false
	if state.baseline_id != baseline_id:
		_mark_resync_needed(state)
		_refresh_readiness_for_lane(lane)
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
	var state := _ensure_lane_state(lane)
	return state.resync_needed

func _ensure_lane_state(lane: String) -> Dictionary:
	if not _lane_state.has(lane):
		_lane_state[lane] = _make_lane_state()
	return _lane_state[lane]

func _make_lane_state() -> Dictionary:
	return {
		"baseline_id": null,
		"snapshot_id": null,
		"sequence": null,
		"chunk_index": 0,
		"chunk_count": 0,
		"is_final_chunk": false,
		"sync_state": LaneSyncState.UNSYNCED,
		"resync_needed": false,
		"pending_full": false,
		"full_packet_seen": false,
	}

func _is_stale_sequence(state: Dictionary, sequence) -> bool:
	if state.sequence == null:
		return false
	return sequence != null and sequence <= state.sequence

func _mark_resync_needed(state: Dictionary) -> void:
	state.sync_state = LaneSyncState.UNSYNCED
	state.resync_needed = true
	state.pending_full = false

func _refresh_readiness() -> void:
	_refresh_readiness_for_lane(LaneMetadata.LANE_WORLD)
	_refresh_readiness_for_lane(LaneMetadata.LANE_OVERLAY)
	_refresh_readiness_for_lane(LaneMetadata.LANE_SESSION)

func _refresh_readiness_for_lane(lane: String) -> void:
	if _readiness == null:
		return
	if _readiness.has_method("set_lane_synced"):
		_readiness.set_lane_synced(lane, is_lane_synced(lane))
