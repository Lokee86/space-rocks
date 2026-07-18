extends RefCounted
class_name ClientMeasurementSnapshotScheduler

const DEFAULT_INTERVAL_SECONDS := 0.25

var coordinator
var interval_seconds := DEFAULT_INTERVAL_SECONDS
var elapsed_seconds := 0.0


func _init(coordinator_ref = null, interval_seconds_ref: float = DEFAULT_INTERVAL_SECONDS) -> void:
	configure(coordinator_ref, interval_seconds_ref)


func configure(coordinator_ref, interval_seconds_ref: float = DEFAULT_INTERVAL_SECONDS) -> void:
	coordinator = coordinator_ref
	interval_seconds = maxf(interval_seconds_ref, 0.0)
	reset()


func process(delta: float) -> void:
	if coordinator == null or !coordinator.is_recording():
		reset()
		return

	elapsed_seconds += maxf(delta, 0.0)
	if elapsed_seconds < interval_seconds:
		return
	elapsed_seconds = 0.0
	if coordinator.pending_request_ids.has("snapshot"):
		return
	coordinator.snapshot()


func reset() -> void:
	elapsed_seconds = 0.0
