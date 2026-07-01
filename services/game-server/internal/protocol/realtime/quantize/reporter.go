package quantize

import "github.com/Lokee86/space-rocks/server/internal/logging"

func ReportUnmappedFloatFallback(lane string, fieldPath string, originalValue float64, fallback Policy) {
	logging.Warn("realtime float quantization fallback",
		"lane", lane,
		"field_path", fieldPath,
		"original_value", originalValue,
		"fallback_policy", fallback.Name,
	)
}
