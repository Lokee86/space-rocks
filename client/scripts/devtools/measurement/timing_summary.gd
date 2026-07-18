extends RefCounted
class_name ClientMeasurementTimingSummary

const HISTOGRAM_BUCKETS := [
	1.0, 2.0, 4.0, 8.0, 16.0, 33.0, 50.0, 100.0,
	250.0, 500.0, 1000.0, 2000.0, 5000.0, 10000.0,
]

var count := 0
var total_ms := 0.0
var minimum_ms := 0.0
var maximum_ms := 0.0
var _buckets: Array = []


func _init() -> void:
	_buckets.resize(HISTOGRAM_BUCKETS.size() + 1)
	_buckets.fill(0)


func record(duration_ms: float) -> void:
	var value: float = maxf(duration_ms, 0.0)
	count += 1
	total_ms += value
	if count == 1:
		minimum_ms = value
		maximum_ms = value
	else:
		minimum_ms = min(minimum_ms, value)
		maximum_ms = max(maximum_ms, value)

	var bucket_index := HISTOGRAM_BUCKETS.size()
	for index in HISTOGRAM_BUCKETS.size():
		if value <= HISTOGRAM_BUCKETS[index]:
			bucket_index = index
			break
	_buckets[bucket_index] += 1


func reset() -> void:
	count = 0
	total_ms = 0.0
	minimum_ms = 0.0
	maximum_ms = 0.0
	_buckets.fill(0)


func snapshot() -> Dictionary:
	return {
		"count": count,
		"total": total_ms,
		"average": total_ms / count if count > 0 else 0.0,
		"minimum": minimum_ms,
		"maximum": maximum_ms,
		"p95": _percentile(0.95),
		"p99": _percentile(0.99),
	}


func _percentile(fraction: float) -> float:
	if count == 0:
		return 0.0
	var target := int(ceil(float(count) * fraction))
	var seen := 0
	for index in _buckets.size():
		seen += int(_buckets[index])
		if seen >= target:
			if index < HISTOGRAM_BUCKETS.size():
				return HISTOGRAM_BUCKETS[index]
			return maximum_ms
	return maximum_ms
