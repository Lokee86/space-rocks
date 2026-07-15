package observabilityevent

import (
	"log/slog"
	"strings"
	"time"
)

// EmitLegacyArgs normalizes slog-compatible adapter arguments into bounded
// scalar fields before entering the bridge-only emission path.
func (e *Emitter) EmitLegacyArgs(request LegacyRequest, args ...any) Result {
	fields, code, key := normalizeLegacyArgs(args)
	if code != "" {
		return e.reject(code, key)
	}
	request.Fields = fields
	return e.EmitLegacy(request)
}

func normalizeLegacyArgs(args []any) (Fields, RejectionCode, string) {
	fields := Fields{}
	for len(args) > 0 {
		if attr, ok := args[0].(slog.Attr); ok {
			args = args[1:]
			if code, key := appendLegacyAttr(fields, "", attr); code != "" {
				return nil, code, key
			}
			continue
		}
		if len(args) < 2 {
			return nil, RejectionCodeInvalidFieldType, ""
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, RejectionCodeInvalidFieldKey, ""
		}
		value := args[1]
		args = args[2:]
		if code, rejectedKey := appendLegacyValue(fields, key, value); code != "" {
			return nil, code, rejectedKey
		}
	}
	return fields, "", ""
}

func appendLegacyAttr(fields Fields, prefix string, attr slog.Attr) (RejectionCode, string) {
	attr.Value = attr.Value.Resolve()
	key := joinLegacyKey(prefix, attr.Key)
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			if code, rejectedKey := appendLegacyAttr(fields, key, child); code != "" {
				return code, rejectedKey
			}
		}
		return "", ""
	}
	return appendLegacyValue(fields, key, attr.Value)
}

func appendLegacyValue(fields Fields, key string, value any) (RejectionCode, string) {
	if logValuer, ok := value.(slog.LogValuer); ok {
		value = logValuer.LogValue().Resolve()
	}
	if slogValue, ok := value.(slog.Value); ok {
		slogValue = slogValue.Resolve()
		switch slogValue.Kind() {
		case slog.KindString:
			value = slogValue.String()
		case slog.KindBool:
			value = slogValue.Bool()
		case slog.KindInt64:
			value = slogValue.Int64()
		case slog.KindUint64:
			value = slogValue.Uint64()
		case slog.KindFloat64:
			value = slogValue.Float64()
		case slog.KindDuration:
			value = float64(slogValue.Duration()) / float64(time.Millisecond)
		case slog.KindTime:
			value = slogValue.Time().UTC().Format(time.RFC3339Nano)
		case slog.KindAny:
			value = slogValue.Any()
		default:
			return RejectionCodeInvalidFieldType, key
		}
	}

	switch typed := value.(type) {
	case error:
		value = typed.Error()
	case time.Time:
		value = typed.UTC().Format(time.RFC3339Nano)
	case time.Duration:
		value = float64(typed) / float64(time.Millisecond)
	}
	if _, ok, isNull := scalarValue(value); isNull {
		return RejectionCodeNullNotAllowed, key
	} else if !ok {
		return RejectionCodeInvalidFieldType, key
	}
	fields[key] = value
	return "", ""
}

func joinLegacyKey(prefix, key string) string {
	key = strings.TrimSpace(key)
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "_" + key
}
