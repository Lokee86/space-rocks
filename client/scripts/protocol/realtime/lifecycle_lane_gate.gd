extends RefCounted
class_name LifecycleLaneGate

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")

const DECISION_APPLY := "apply"
const DECISION_QUEUE := "queue"
const DECISION_REJECT := "reject"

const MAX_PENDING_BASELINE_BUCKETS := 4
const MAX_PENDING_PACKETS_PER_LANE := 8

const _LIFECYCLE_LANES := [
	LaneMetadata.LANE_ASTEROIDS_LIFECYCLE,
	LaneMetadata.LANE_BULLETS_LIFECYCLE,
]

var _latest_applied_sequence := {}
var _pending_by_baseline := {}
var _pending_baseline_order := []
var _pending_sequences := {}

func _init() -> void:
	for lane in _LIFECYCLE_LANES:
		_latest_applied_sequence[lane] = null
		_pending_sequences[lane] = {}

func submit(lane, packet, world_synced, active_world_baseline_id) -> Dictionary:
	if not _latest_applied_sequence.has(lane):
		return _reject(lane, packet, null, "unsupported_lane")
	if not packet is Dictionary:
		return _reject(lane, packet, null, "invalid_packet")

	var sequence = packet.get("sequence")
	if sequence == null:
		return _reject(lane, packet, null, "missing_sequence")
	if typeof(sequence) != TYPE_INT:
		return _reject(lane, packet, sequence, "sequence_not_integer")

	var baseline_id = packet.get("baseline_id")
	if baseline_id == null or baseline_id == "":
		return _reject(lane, packet, sequence, "missing_baseline_id")

	var latest_sequence = _latest_applied_sequence[lane]
	if latest_sequence != null and sequence <= latest_sequence:
		return _reject(lane, packet, sequence, "stale_sequence")
	if _pending_sequences[lane].has(sequence):
		return _reject(lane, packet, sequence, "duplicate_pending_sequence")

	var decision := {
		"status": DECISION_QUEUE,
		"lane": lane,
		"packet": packet,
		"sequence": sequence,
	}
	if world_synced and baseline_id == active_world_baseline_id:
		decision.status = DECISION_APPLY
	else:
		_store_pending(lane, packet, sequence, baseline_id, active_world_baseline_id)
	return decision

func mark_applied(lane, sequence) -> void:
	if not _latest_applied_sequence.has(lane):
		return
	if typeof(sequence) != TYPE_INT:
		return
	var latest_sequence = _latest_applied_sequence[lane]
	if latest_sequence == null or sequence > latest_sequence:
		_latest_applied_sequence[lane] = sequence

func take_pending_for_baseline(baseline_id) -> Array:
	var baseline_bucket: Dictionary = _pending_by_baseline.get(baseline_id, {})
	var drained := []
	for lane in _LIFECYCLE_LANES:
		var lane_packets: Array = baseline_bucket.get(lane, [])
		var sorted_packets := lane_packets.duplicate()
		sorted_packets.sort_custom(func(left, right): return left.get("sequence") < right.get("sequence"))
		for packet in sorted_packets:
			var sequence = packet.get("sequence")
			drained.append({"lane": lane, "packet": packet, "sequence": sequence})
			_pending_sequences[lane].erase(sequence)
	if _pending_by_baseline.has(baseline_id):
		_pending_by_baseline.erase(baseline_id)
		_pending_baseline_order.erase(baseline_id)
	return drained

func discard_obsolete_baselines(active_baseline_id) -> void:
	var active_number = _parse_world_baseline_number(active_baseline_id)
	if active_number == null:
		return
	for baseline_id in _pending_baseline_order.duplicate():
		var baseline_number = _parse_world_baseline_number(baseline_id)
		if baseline_number != null and baseline_number < active_number:
			_discard_baseline(baseline_id)

func reset() -> void:
	for lane in _LIFECYCLE_LANES:
		_latest_applied_sequence[lane] = null
		_pending_sequences[lane].clear()
	_pending_by_baseline.clear()
	_pending_baseline_order.clear()

func _parse_world_baseline_number(baseline_id):
	if not baseline_id is String or not baseline_id.begins_with("world-baseline-"):
		return null
	var suffix: String = baseline_id.substr("world-baseline-".length())
	if suffix.is_empty() or not suffix.is_valid_int():
		return null
	return int(suffix)

func _reject(lane, packet, sequence, reason: String) -> Dictionary:
	return {
		"status": DECISION_REJECT,
		"lane": lane,
		"packet": packet,
		"sequence": sequence,
		"reason": reason,
	}

func _store_pending(lane, packet, sequence, baseline_id, active_world_baseline_id) -> void:
	if not _pending_by_baseline.has(baseline_id):
		_trim_baseline_buckets(active_world_baseline_id)
		_pending_by_baseline[baseline_id] = {}
		_pending_baseline_order.append(baseline_id)
	if not _pending_by_baseline[baseline_id].has(lane):
		_pending_by_baseline[baseline_id][lane] = []
	var lane_packets: Array = _pending_by_baseline[baseline_id][lane]
	if lane_packets.size() >= MAX_PENDING_PACKETS_PER_LANE:
		var discarded_packet: Dictionary = lane_packets.pop_front()
		_pending_sequences[lane].erase(discarded_packet.get("sequence"))
	lane_packets.append(packet)
	_pending_sequences[lane][sequence] = true

func _trim_baseline_buckets(active_world_baseline_id) -> void:
	while _pending_by_baseline.size() >= MAX_PENDING_BASELINE_BUCKETS:
		var discarded_baseline = _oldest_discardable_baseline(active_world_baseline_id)
		if discarded_baseline == null:
			return
		_discard_baseline(discarded_baseline)

func _oldest_discardable_baseline(active_world_baseline_id):
	for baseline_id in _pending_baseline_order:
		if baseline_id != active_world_baseline_id:
			return baseline_id
	return null

func _discard_baseline(baseline_id) -> void:
	var baseline_bucket: Dictionary = _pending_by_baseline.get(baseline_id, {})
	for lane in baseline_bucket:
		for packet in baseline_bucket[lane]:
			_pending_sequences[lane].erase(packet.get("sequence"))
	_pending_by_baseline.erase(baseline_id)
	_pending_baseline_order.erase(baseline_id)
