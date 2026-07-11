package realtime

import (
	"strconv"
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtimewire"
)

type compactWireDescriptorIndexes struct {
	records   map[string]realtimewire.RealtimeWireRecord
	bindings  map[string][]string
	codecs    map[string]realtimewire.RealtimeWireIDCodec
	selectors map[string]realtimewire.RealtimeWireIDSelector
	events    map[string]realtimewire.RealtimeWireEvent
}

var (
	compactWireDescriptorOnce sync.Once
	compactWireDescriptors    compactWireDescriptorIndexes
)

func compactWireDescriptorIndexesOnce() compactWireDescriptorIndexes {
	compactWireDescriptorOnce.Do(func() {
		compactWireDescriptors = compactWireDescriptorIndexes{
			records:   make(map[string]realtimewire.RealtimeWireRecord),
			bindings:  make(map[string][]string),
			codecs:    make(map[string]realtimewire.RealtimeWireIDCodec),
			selectors: make(map[string]realtimewire.RealtimeWireIDSelector),
			events:    realtimewire.RealtimeWireEventsByReadable,
		}
		for _, record := range realtimewire.RealtimeWireRecords {
			compactWireDescriptors.records[record.ID] = record
		}
		for _, binding := range realtimewire.RealtimeWirePacketFieldBindings {
			key := binding.PacketID + "." + binding.ReadableField
			compactWireDescriptors.bindings[key] = append(compactWireDescriptors.bindings[key], binding.RecordIDs...)
		}
		for _, codec := range realtimewire.RealtimeWireIDCodecs {
			compactWireDescriptors.codecs[codec.ID] = codec
		}
		for _, selector := range realtimewire.RealtimeWireIDSelectors {
			compactWireDescriptors.selectors[selector.ID] = selector
		}
	})
	return compactWireDescriptors
}

// compactWirePacketFromDescriptors is the production descriptor-driven encoder.
func compactWirePacketFromDescriptors(packet map[string]any) map[string]any {
	indexes := compactWireDescriptorIndexesOnce()
	packetType := compactWireInputPacketType(packet)
	if packetType == "" {
		return (compactWireValue(packet, "")).(map[string]any)
	}
	compactPacket := make(map[string]any, len(packet))
	for key, value := range packet {
		readableKey := compactWireReadableInputKey(key)
		if recordIDs := indexes.bindings[packetType+"."+readableKey]; len(recordIDs) > 0 {
			compactPacket[readableKey] = compactWireEncodeBoundField(value, recordIDs, indexes)
			continue
		}
		compactPacket[readableKey] = compactWireValue(value, readableKey)
	}
	return compactWireAliasMap(compactPacket)
}

func compactWireEncodeBoundField(value any, recordIDs []string, indexes compactWireDescriptorIndexes) any {
	if len(recordIDs) == 0 {
		return compactWireValue(value, "")
	}
	record := indexes.records[recordIDs[0]]
	if record.Encoding == "discriminated_event_tuple" {
		return compactWireEncodeEvents(value, indexes)
	}
	if record.Encoding == "scalar_id" {
		return compactWireEncodeScalar(value, record, indexes)
	}
	if record.Encoding == "scalar_id_list" {
		values, ok := value.([]any)
		if !ok {
			return value
		}
		encoded := make([]any, len(values))
		for i, item := range values {
			encoded[i] = compactWireEncodeField(item, record.Fields[0], nil, indexes)
		}
		return encoded
	}
	if record.Encoding == "scalar_list" {
		return compactWireValue(value, record.Fields[0].JSON)
	}
	values, ok := value.([]any)
	if !ok {
		return compactWireEncodeRecord(value, record, indexes)
	}
	encoded := make([]any, len(values))
	for i, item := range values {
		actual := record
		if len(recordIDs) > 1 {
			if fields, ok := item.(map[string]any); ok {
				eventTypeValue, _ := compactWireRecordFieldValueByName(fields, "type")
				eventType := compactWireReadableEventType(asString(eventTypeValue))
				if event, found := indexes.events[eventType]; found {
					actual = indexes.records[event.RecordID]
				}
			}
		}
		encoded[i] = compactWireEncodeRecord(item, actual, indexes)
	}
	return encoded
}

func compactWireEncodeEvents(value any, indexes compactWireDescriptorIndexes) any {
	values, ok := value.([]any)
	if !ok {
		return compactWireValue(value, "events")
	}
	encoded := make([]any, len(values))
	for i, item := range values {
		record, ok := item.(map[string]any)
		if !ok {
			encoded[i] = item
			continue
		}
		eventTypeValue, _ := compactWireRecordFieldValueByName(record, "type")
		eventType := compactWireReadableEventType(asString(eventTypeValue))
		event, found := indexes.events[eventType]
		if !found {
			encoded[i] = compactWireValue(record, "")
			continue
		}
		encoded[i] = compactWireEncodeRecord(record, indexes.records[event.RecordID], indexes)
	}
	return encoded
}

func compactWireEncodeRecord(value any, record realtimewire.RealtimeWireRecord, indexes compactWireDescriptorIndexes) any {
	fields, ok := value.(map[string]any)
	if !ok {
		return value
	}
	switch record.Encoding {
	case "map":
		encoded := make(map[string]any, len(fields))
		known := make(map[string]bool, len(record.Fields))
		for _, field := range record.Fields {
			known[field.JSON] = true
			if field.CompactKey != "" {
				known[field.CompactKey] = true
			}
			if item, present := compactWireRecordFieldValue(fields, field); present {
				encoded[field.CompactKey] = compactWireEncodeField(item, field, fields, indexes)
			}
		}
		for key, item := range fields {
			if !known[key] {
				encoded[compactWireReadableKey(key)] = compactWireValue(item, key)
			}
		}
		return encoded
	case "fixed_tuple":
		encoded := make([]any, len(record.Fields))
		for i, field := range record.Fields {
			encoded[i] = nil
			if item, present := compactWireRecordFieldValue(fields, field); present {
				encoded[i] = compactWireEncodeField(item, field, fields, indexes)
			}
		}
		return encoded
	case "sparse_positional_tuple":
		encoded := make([]any, len(record.Fields))
		for i, field := range record.Fields {
			if item, present := compactWireRecordFieldValue(fields, field); present {
				encoded[i] = compactWireEncodeField(item, field, fields, indexes)
			}
		}
		for len(encoded) > 0 && encoded[len(encoded)-1] == nil {
			encoded = encoded[:len(encoded)-1]
		}
		return encoded
	case "sparse_key_value_tuple":
		encoded := []any{}
		for i, field := range record.Fields {
			item, present := compactWireRecordFieldValue(fields, field)
			if i == 0 {
				if present {
					encoded = append(encoded, compactWireEncodeField(item, field, fields, indexes))
				}
				continue
			}
			if present {
				encoded = append(encoded, field.CompactKey, compactWireEncodeField(item, field, fields, indexes))
			}
		}
		if record.PreserveUnknownFields {
			known := make(map[string]bool, len(record.Fields))
			for _, field := range record.Fields {
				known[field.JSON] = true
				if field.CompactKey != "" {
					known[field.CompactKey] = true
				}
			}
			for key, item := range fields {
				if !known[key] {
					encoded = append(encoded, compactWireReadableKey(key), compactWireValue(item, key))
				}
			}
		}
		return encoded
	default:
		return compactWireValue(fields, "")
	}
}

func compactWireEncodeScalar(value any, record realtimewire.RealtimeWireRecord, indexes compactWireDescriptorIndexes) any {
	if len(record.Fields) == 0 {
		return value
	}
	return compactWireEncodeField(value, record.Fields[0], nil, indexes)
}

func compactWireEncodeField(value any, field realtimewire.RealtimeWireField, fields map[string]any, indexes compactWireDescriptorIndexes) any {
	if field.IDCodecBy != "" && fields != nil {
		selectorValue, _ := compactWireRecordFieldValueByName(fields, field.IDCodecBy)
		return compactWireEncodeSelector(value, selectorValue, field.IDCodecBy, indexes)
	}
	if field.IDCodec != "" {
		return compactWireEncodeCodec(value, indexes.codecs[field.IDCodec])
	}
	if field.ValueDomain != "" {
		return compactWireEncodeValueDomain(value, field.ValueDomain)
	}
	return compactWireValue(value, field.JSON)
}

func compactWireEncodeSelector(value any, selectorValue any, selectorID string, indexes compactWireDescriptorIndexes) any {
	selector, ok := indexes.selectors[selectorID]
	if ok {
		for _, mapping := range selector.Mappings {
			if mapping.Value == asString(selectorValue) {
				return compactWireEncodeCodec(value, indexes.codecs[mapping.CodecID])
			}
		}
		if selector.FallbackTagged {
			return compactWireEncodeKnownTaggedID(value, indexes)
		}
	}
	return value
}

func compactWireEncodeKnownTaggedID(value any, indexes compactWireDescriptorIndexes) any {
	for _, codec := range indexes.codecs {
		if encoded, ok := compactWireEncodeCodecTagged(value, codec); ok {
			return []any{codec.Tag, encoded}
		}
	}
	return value
}

func compactWireEncodeCodec(value any, codec realtimewire.RealtimeWireIDCodec) any {
	encoded, _ := compactWireEncodeCodecTagged(value, codec)
	return encoded
}

func compactWireEncodeID(codecID string, value any) any {
	indexes := compactWireDescriptorIndexesOnce()
	codec, ok := indexes.codecs[codecID]
	if !ok {
		return value
	}
	return compactWireEncodeCodec(value, codec)
}

func compactWireEncodeCodecTagged(value any, codec realtimewire.RealtimeWireIDCodec) (any, bool) {
	id, ok := value.(string)
	if !ok || len(id) < len(codec.Prefix) || id[:len(codec.Prefix)] != codec.Prefix {
		return value, false
	}
	number, err := strconv.Atoi(id[len(codec.Prefix):])
	if err != nil {
		return value, false
	}
	return number, true
}

func compactWireEncodeValueDomain(value any, domain string) any {
	mapping := realtimewire.RealtimeWireValueCompactByReadable[domain]
	if compacted, ok := mapping[asString(value)]; ok {
		return compacted
	}
	return value
}

func compactWireInputPacketType(packet map[string]any) string {
	packetType, _ := packet["type"].(string)
	if packetType == "" {
		packetType, _ = packet["t"].(string)
	}
	if readable, ok := realtimewire.RealtimeWireValueReadableByCompact["packet_type"][packetType]; ok {
		return readable
	}
	return packetType
}

func compactWireReadableInputKey(key string) string {
	if readable, ok := realtimewire.RealtimeWireKeyReadableByCompact["wire."+key]; ok {
		return readable
	}
	return key
}

func compactWireReadableEventType(eventType string) string {
	if readable, ok := realtimewire.RealtimeWireValueReadableByCompact["event_type"][eventType]; ok {
		return readable
	}
	return eventType
}

func compactWireRecordFieldValue(fields map[string]any, field realtimewire.RealtimeWireField) (any, bool) {
	if value, ok := fields[field.JSON]; ok {
		return value, true
	}
	if field.CompactKey != "" {
		if value, ok := fields[field.CompactKey]; ok {
			return value, true
		}
	}
	return nil, false
}

func compactWireRecordFieldValueByName(fields map[string]any, name string) (any, bool) {
	if value, ok := fields[name]; ok {
		return value, true
	}
	if compactKey, ok := realtimewire.RealtimeWireKeyCompactByReadable["wire."+name]; ok {
		if value, ok := fields[compactKey]; ok {
			return value, true
		}
	}
	return nil, false
}

func compactWireReadableKey(key string) string {
	if compactKey, ok := realtimewire.RealtimeWireKeyCompactByReadable["wire."+key]; ok {
		return compactKey
	}
	return key
}

func compactWireAliasMap(value map[string]any) map[string]any {
	encoded := make(map[string]any, len(value))
	for key, item := range value {
		compactKey := compactWireReadableKey(key)
		if compactKey == "" {
			compactKey = key
		}
		encoded[compactKey] = item
	}
	return encoded
}
