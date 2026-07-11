package realtime

import "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtimewire"

func CompactWirePacket(packet map[string]any) map[string]any {
	return compactWirePacketFromDescriptors(packet)
}

func compactWireValue(value any, parentKey string) any {
	switch typed := value.(type) {
	case map[string]any:
		expanded := make(map[string]any, len(typed))
		for key, child := range typed {
			compactKey := key
			if mapped, ok := realtimewire.RealtimeWireKeyCompactByReadable["wire."+key]; ok {
				compactKey = mapped
			}
			expanded[compactKey] = compactWireValue(child, key)
		}
		return expanded
	case []any:
		expanded := make([]any, len(typed))
		for i := range typed {
			expanded[i] = compactWireValue(typed[i], parentKey)
		}
		return expanded
	default:
		if compacted, ok := compactWireValueAlias(parentKey, asString(typed)); ok {
			return compacted
		}
		return value
	}
}

func compactWireValueAlias(parentKey, value string) (string, bool) {
	var domains []string
	switch parentKey {
	case "type":
		domains = []string{"packet_type", "event_type"}
	case "lane":
		domains = []string{"lane"}
	case "snapshot_kind":
		domains = []string{"snapshot_kind"}
	default:
		return "", false
	}
	for _, domain := range domains {
		if mapping, ok := realtimewire.RealtimeWireValueCompactByReadable[domain]; ok {
			if compacted, ok := mapping[value]; ok {
				return compacted, true
			}
		}
	}
	return "", false
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
