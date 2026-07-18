extends RefCounted
class_name ClientMeasurementLifecycleCounterWindow

const DEFAULT_CAPACITY := 60

var capacity := DEFAULT_CAPACITY
var _elapsed_seconds := 0.0
var _current: Dictionary = {}
var _windows: Array = []


func _init(capacity_ref: int = DEFAULT_CAPACITY) -> void:
	capacity = max(capacity_ref, 1)


func record(entity_kind: String, operation: String, count: int) -> void:
	if count <= 0:
		return
	if not _current.has(entity_kind):
		_current[entity_kind] = {}
	var entity_counts: Dictionary = _current[entity_kind]
	entity_counts[operation] = int(entity_counts.get(operation, 0)) + count


func advance(delta: float) -> void:
	_elapsed_seconds += max(delta, 0.0)
	while _elapsed_seconds >= 1.0:
		_windows.append(_current.duplicate(true))
		_current.clear()
		_elapsed_seconds -= 1.0
		while _windows.size() > capacity:
			_windows.pop_front()


func reset() -> void:
	_elapsed_seconds = 0.0
	_current.clear()
	_windows.clear()


func snapshot() -> Array:
	var result := _windows.duplicate(true)
	if not _current.is_empty():
		result.append(_current.duplicate(true))
	return result
