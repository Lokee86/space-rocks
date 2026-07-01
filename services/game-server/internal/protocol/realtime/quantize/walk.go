package quantize

import (
	"fmt"
	"reflect"
)

func QuantizeValue(lane string, fieldPath string, value any) (any, error) {
	return quantizeValue(lane, fieldPath, value)
}

func quantizeValue(lane string, fieldPath string, value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch typed := value.(type) {
	case float32:
		return quantizeFloat(lane, fieldPath, float64(typed))
	case float64:
		return quantizeFloat(lane, fieldPath, typed)
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		nil:
		return value, nil
	case map[string]any:
		return quantizeMap(lane, fieldPath, typed)
	case []any:
		return quantizeSlice(lane, fieldPath, typed)
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return value, nil
		}
		return quantizeReflectMap(lane, fieldPath, rv)
	case reflect.Slice, reflect.Array:
		return quantizeReflectSlice(lane, fieldPath, rv)
	case reflect.Float32, reflect.Float64:
		return quantizeFloat(lane, fieldPath, rv.Convert(reflect.TypeOf(float64(0))).Float())
	default:
		return value, nil
	}
}

func quantizeFloat(lane string, fieldPath string, value float64) (any, error) {
	policy, ok := LookupPolicy(fieldPath)
	if !ok {
		ReportUnmappedFloatFallback(lane, fieldPath, value, mustPolicy(PolicyFloatGeneric))
		policy = mustPolicy(PolicyFloatGeneric)
	}

	encoded, err := EncodeFloat(policy, value)
	if err != nil {
		return nil, fmt.Errorf("quantize %s: %w", fieldPath, err)
	}

	return encoded, nil
}

func quantizeMap(lane string, fieldPath string, value map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(value))
	for key, child := range value {
		childPath := joinFieldPath(fieldPath, key)
		quantized, err := quantizeValue(lane, childPath, child)
		if err != nil {
			return nil, err
		}
		out[key] = quantized
	}
	return out, nil
}

func quantizeSlice(lane string, fieldPath string, value []any) ([]any, error) {
	out := make([]any, len(value))
	for index, child := range value {
		childPath := indexFieldPath(fieldPath, index)
		quantized, err := quantizeValue(lane, childPath, child)
		if err != nil {
			return nil, err
		}
		out[index] = quantized
	}
	return out, nil
}

func quantizeReflectMap(lane string, fieldPath string, rv reflect.Value) (map[string]any, error) {
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		childPath := joinFieldPath(fieldPath, key)
		quantized, err := quantizeValue(lane, childPath, iter.Value().Interface())
		if err != nil {
			return nil, err
		}
		out[key] = quantized
	}
	return out, nil
}

func quantizeReflectSlice(lane string, fieldPath string, rv reflect.Value) ([]any, error) {
	out := make([]any, rv.Len())
	for index := 0; index < rv.Len(); index++ {
		childPath := indexFieldPath(fieldPath, index)
		quantized, err := quantizeValue(lane, childPath, rv.Index(index).Interface())
		if err != nil {
			return nil, err
		}
		out[index] = quantized
	}
	return out, nil
}

func joinFieldPath(parent string, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func indexFieldPath(parent string, index int) string {
	if parent == "" {
		return fmt.Sprintf("[%d]", index)
	}
	return fmt.Sprintf("%s[%d]", parent, index)
}
