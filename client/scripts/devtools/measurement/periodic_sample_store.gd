extends RefCounted
class_name ClientMeasurementPeriodicSampleStore

const DEFAULT_CAPACITY := 60

var capacity := DEFAULT_CAPACITY
var _samples: Array = []
var dropped_count := 0


func _init(capacity_ref: int = DEFAULT_CAPACITY) -> void:
	capacity = max(capacity_ref, 1)


func append(sample: Dictionary) -> void:
	_samples.append(sample.duplicate(true))
	while _samples.size() > capacity:
		_samples.pop_front()
		dropped_count += 1


func reset() -> void:
	_samples.clear()
	dropped_count = 0


func snapshot() -> Dictionary:
	return {
		"count": _samples.size(),
		"dropped": dropped_count,
		"samples": _samples.duplicate(true),
	}
