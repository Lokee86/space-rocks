extends RefCounted
class_name ClientMeasurementSession

const TimingSummary := preload("res://scripts/devtools/measurement/timing_summary.gd")
const PeriodicSampleStore := preload("res://scripts/devtools/measurement/periodic_sample_store.gd")
const LifecycleCounterWindow := preload("res://scripts/devtools/measurement/lifecycle_counter_window.gd")

signal finalized(result: Dictionary)

var _clock: Callable
var _recording := false
var _metadata: Dictionary = {}
var _start_msec := -1
var _end_msec := -1
var _frame_elapsed_seconds := 0.0
var _sample_elapsed_seconds := 0.0
var _last_result: Dictionary = {}
var _periodic_sample_provider: Callable
var _frame_timing := TimingSummary.new()
var _lane_timing: Dictionary = {}
var _presentation_timing := TimingSummary.new()
var _periodic_samples := PeriodicSampleStore.new()
var _lifecycle_window := LifecycleCounterWindow.new()
var _lifecycle_churn: Dictionary = {}
var _entity_counts: Dictionary = {}
var _latest_network_metrics: Dictionary = {}
var _latest_long_match_store_metrics: Dictionary = {}
var _missing_metrics: Dictionary = {}


func _init(clock: Callable = Callable(Time, "get_ticks_msec")) -> void:
	_clock = clock


func set_periodic_sample_provider(provider: Callable) -> void:
	_periodic_sample_provider = provider


func start(metadata: Dictionary = {}) -> void:
	if _recording:
		stop("replaced")
	_clear_run()
	_recording = true
	_metadata = metadata.duplicate(true)
	_start_msec = _now_msec()
	_end_msec = -1


func stop(reason: String = "") -> Dictionary:
	if not _recording:
		return _last_result.duplicate(true)
	_end_msec = _now_msec()
	_recording = false
	var status := "partial" if not reason.is_empty() else "completed"
	var result := _build_snapshot(status, reason)
	_last_result = result.duplicate(true)
	finalized.emit(result.duplicate(true))
	return result


func reset() -> Dictionary:
	var result := {}
	if _recording:
		result = stop("reset")
	else:
		result = _last_result.duplicate(true)
	_clear_run()
	_recording = false
	_metadata = {}
	_start_msec = -1
	_end_msec = -1
	_last_result = {}
	return result


func is_recording() -> bool:
	return _recording


func process_frame(delta: float) -> void:
	if not _recording:
		return
	var frame_delta: float = maxf(delta, 0.0)
	_frame_timing.record(frame_delta * 1000.0)
	_frame_elapsed_seconds += frame_delta
	_sample_elapsed_seconds += frame_delta
	_lifecycle_window.advance(frame_delta)
	while _sample_elapsed_seconds >= 1.0:
		_sample_elapsed_seconds -= 1.0
		if _periodic_sample_provider.is_valid():
			var sample = _periodic_sample_provider.call()
			if sample is Dictionary:
				record_periodic_sample(sample)


func record_lane_application(lane: String, duration_ms: float) -> void:
	if not _recording:
		return
	if not _lane_timing.has(lane):
		_lane_timing[lane] = TimingSummary.new()
	_lane_timing[lane].record(duration_ms)


func record_presentation(duration_ms: float) -> void:
	if _recording:
		_presentation_timing.record(duration_ms)


func record_lifecycle(entity_kind: String, operation: String, count: int = 1) -> void:
	if not _recording or count <= 0:
		return
	if not _lifecycle_churn.has(entity_kind):
		_lifecycle_churn[entity_kind] = {"created": 0, "removed": 0, "updated": 0}
	var entity_churn: Dictionary = _lifecycle_churn[entity_kind]
	entity_churn[operation] = int(entity_churn.get(operation, 0)) + count
	_lifecycle_window.record(entity_kind, operation, count)


func record_periodic_sample(sample: Dictionary) -> void:
	if not _recording:
		return
	var copied := sample.duplicate(true)
	_periodic_samples.append(copied)
	if copied.has("entity_counts"):
		_entity_counts = copied["entity_counts"].duplicate(true)
	if copied.has("network_metrics"):
		_latest_network_metrics = copied["network_metrics"].duplicate(true)
	if copied.has("long_match_store_metrics"):
		_latest_long_match_store_metrics = copied["long_match_store_metrics"].duplicate(true)
	for metric in copied.get("missing_metrics", []):
		_missing_metrics[str(metric)] = true


func snapshot() -> Dictionary:
	if _recording:
		return _build_snapshot("recording", "")
	return _last_result.duplicate(true)


func _build_snapshot(status: String, partial_reason: String) -> Dictionary:
	var lane_snapshot := {}
	for lane in _lane_timing.keys():
		lane_snapshot[lane] = _lane_timing[lane].snapshot()
	var duration_msec := _duration_msec()
	return {
		"status": status,
		"partial_reason": partial_reason,
		"start": _start_msec,
		"end": _end_msec if _end_msec >= 0 else _now_msec(),
		"duration": duration_msec,
		"metadata": _metadata.duplicate(true),
		"frame_timing": _frame_timing.snapshot(),
		"lane_application_timing": lane_snapshot,
		"presentation_timing": _presentation_timing.snapshot(),
		"entity_counts": _entity_counts.duplicate(true),
		"lifecycle_churn": {
			"cumulative": _lifecycle_churn.duplicate(true),
			"recent_window": _lifecycle_window.snapshot(),
		},
		"resource_samples": _periodic_samples.snapshot(),
		"network_metrics": _latest_network_metrics.duplicate(true),
		"long_match_store_metrics": _latest_long_match_store_metrics.duplicate(true),
		"missing_metrics": _missing_metrics.keys(),
	}


func _duration_msec() -> float:
	var elapsed_from_clock := 0.0
	if _start_msec >= 0:
		var end_msec := _end_msec if _end_msec >= 0 else _now_msec()
		elapsed_from_clock = max(float(end_msec - _start_msec), 0.0)
	return max(elapsed_from_clock, _frame_elapsed_seconds * 1000.0)


func _clear_run() -> void:
	_frame_elapsed_seconds = 0.0
	_sample_elapsed_seconds = 0.0
	_frame_timing.reset()
	_lane_timing.clear()
	_presentation_timing.reset()
	_periodic_samples.reset()
	_lifecycle_window.reset()
	_lifecycle_churn.clear()
	_entity_counts.clear()
	_latest_network_metrics.clear()
	_latest_long_match_store_metrics.clear()
	_missing_metrics.clear()


func _now_msec() -> int:
	return int(_clock.call()) if _clock.is_valid() else Time.get_ticks_msec()
