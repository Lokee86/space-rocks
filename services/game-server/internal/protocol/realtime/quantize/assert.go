package quantize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func AssertNoRawFloats(lane string, fieldPath string, value any) {
	for _, leak := range CollectRawFloatLeaks(fieldPath, value) {
		logging.Warn("realtime quantization raw float leak",
			"lane", lane,
			"field_path", leak.FieldPath,
			"value", leak.Value,
		)
	}
}

type RawFloatLeak struct {
	FieldPath string
	Value     any
}

func CollectRawFloatLeaks(fieldPath string, value any) []RawFloatLeak {
	var leaks []RawFloatLeak
	collectRawFloatLeaks(fieldPath, value, &leaks)
	return leaks
}

func collectRawFloatLeaks(fieldPath string, value any, leaks *[]RawFloatLeak) {
	if value == nil {
		return
	}

	switch typed := value.(type) {
	case float32, float64:
		*leaks = append(*leaks, RawFloatLeak{FieldPath: fieldPath, Value: typed})
		return
	case map[string]any:
		for key, child := range typed {
			collectRawFloatLeaks(joinFieldPath(fieldPath, key), child, leaks)
		}
		return
	case []any:
		for index, child := range typed {
			collectRawFloatLeaks(indexFieldPath(fieldPath, index), child, leaks)
		}
		return
	}

	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
		value = rv.Interface()
	}

	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		for index := 0; index < rv.NumField(); index++ {
			field := rt.Field(index)
			if field.PkgPath != "" {
				continue
			}
			childPath := joinFieldPath(fieldPath, fieldPathSegment(field.Name))
			collectRawFloatLeaks(childPath, rv.Field(index).Interface(), leaks)
		}
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return
		}
		iter := rv.MapRange()
		for iter.Next() {
			collectRawFloatLeaks(joinFieldPath(fieldPath, iter.Key().String()), iter.Value().Interface(), leaks)
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < rv.Len(); index++ {
			collectRawFloatLeaks(indexFieldPath(fieldPath, index), rv.Index(index).Interface(), leaks)
		}
	}
}

func fieldPathSegment(name string) string {
	if name == "" {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func FormatRawFloatLeak(fieldPath string, value any) string {
	return fmt.Sprintf("%s=%v", fieldPath, value)
}
