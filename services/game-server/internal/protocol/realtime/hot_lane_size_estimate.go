package realtime

import (
	"fmt"
	"strconv"
)

func estimateCompactJSONTupleBytes(items []any) int {
	if len(items) == 0 {
		return 2
	}

	size := 2 + len(items) - 1
	for _, item := range items {
		size += estimateJSONValueBytes(item)
	}
	return size
}

func estimateJSONValueBytes(value any) int {
	switch typed := value.(type) {
	case nil:
		return 4
	case string:
		return estimateJSONStringBytes(typed)
	case int:
		return estimateJSONIntBytes(int64(typed))
	case int8:
		return estimateJSONIntBytes(int64(typed))
	case int16:
		return estimateJSONIntBytes(int64(typed))
	case int32:
		return estimateJSONIntBytes(int64(typed))
	case int64:
		return estimateJSONIntBytes(typed)
	case uint:
		return estimateJSONUintBytes(uint64(typed))
	case uint8:
		return estimateJSONUintBytes(uint64(typed))
	case uint16:
		return estimateJSONUintBytes(uint64(typed))
	case uint32:
		return estimateJSONUintBytes(uint64(typed))
	case uint64:
		return estimateJSONUintBytes(typed)
	case bool:
		if typed {
			return 4
		}
		return 5
	default:
		return estimateJSONStringBytes(fmt.Sprint(value))
	}
}

func estimateJSONIntBytes(value int64) int {
	if value == 0 {
		return 1
	}

	size := 0
	if value < 0 {
		size++
	}

	if value == -9223372036854775808 {
		return size + len("9223372036854775808")
	}
	if value < 0 {
		value = -value
	}
	for value > 0 {
		size++
		value /= 10
	}
	return size
}

func estimateJSONUintBytes(value uint64) int {
	if value == 0 {
		return 1
	}
	return len(strconv.FormatUint(value, 10))
}

func estimateJSONStringBytes(value string) int {
	size := 2 + len(value)
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '"', '\\':
			size++
		case '\b', '\f', '\n', '\r', '\t':
			size++
		default:
			if value[i] < 0x20 {
				size += 5
			}
		}
	}
	return size
}

func estimateCompactBulletMovementUpdateBytes(update map[string]any) int {
	if update == nil {
		return estimateCompactJSONTupleBytes([]any{compactWirePackBulletID(nil)})
	}

	id := update["id"]
	if id == nil {
		id = update["i"]
	}

	items := []any{compactWirePackBulletID(id)}
	x, hasX := update["x"]
	y, hasY := update["y"]
	rotation, hasRotation := update["rotation"]
	if !hasRotation {
		rotation, hasRotation = update["r"]
	}

	switch {
	case hasX && hasY && hasRotation:
		items = append(items, x, y, rotation)
	case hasX && hasY:
		items = append(items, x, y)
	case hasX && hasRotation:
		items = append(items, x, nil, rotation)
	case hasX:
		items = append(items, x)
	case hasY:
		items = append(items, nil, y)
	case hasRotation:
		items = append(items, nil, nil, rotation)
	}

	return estimateCompactJSONTupleBytes(items)
}

func estimateCompactAsteroidMovementUpdateBytes(update map[string]any) int {
	if update == nil {
		return estimateCompactJSONTupleBytes([]any{compactWirePackAsteroidID(nil)})
	}

	id := update["id"]
	if id == nil {
		id = update["i"]
	}

	items := []any{compactWirePackAsteroidID(id)}
	x, hasX := update["x"]
	y, hasY := update["y"]
	if hasX {
		items = append(items, x)
		if hasY {
			items = append(items, y)
		}
		return estimateCompactJSONTupleBytes(items)
	}
	if hasY {
		items = append(items, nil, y)
	}

	return estimateCompactJSONTupleBytes(items)
}

func estimateBulletDeltaPacketBytes(packet BulletWireDeltaPacket, updates []map[string]any) int {
	return estimateCompactLaneDeltaPacketBytes(packet.Type, packet.Metadata, "bu", updates, estimateCompactBulletMovementUpdateBytes)
}

func estimateAsteroidDeltaPacketBytes(packet AsteroidWireDeltaPacket, updates []map[string]any) int {
	return estimateCompactLaneDeltaPacketBytes(packet.Type, packet.Metadata, "au", updates, estimateCompactAsteroidMovementUpdateBytes)
}

func estimateCompactLaneDeltaPacketBytes(packetType string, metadata Metadata, updateField string, updates []map[string]any, updateEstimator func(map[string]any) int) int {
	size := 2
	fields := 0

	addField := func(key string, valueBytes int) {
		if fields > 0 {
			size++
		}
		size += estimateJSONStringBytes(key)
		size++
		size += valueBytes
		fields++
	}

	addField("t", estimateJSONValueBytes(compactWireValue(packetType, "type")))
	addField("q", estimateJSONIntBytes(int64(metadata.Sequence)))
	addField("ms", estimateJSONIntBytes(int64(metadata.ServerSentMsec)))

	if metadata.SnapshotKind == SnapshotKind("delta") {
		if baselineSequence, ok := runtimeBaselineSequence(metadata.Lane, metadata.BaselineID); ok {
			addField("bq", estimateJSONIntBytes(int64(baselineSequence)))
		} else if metadata.BaselineID != "" {
			addField("b", estimateJSONStringBytes(metadata.BaselineID))
		}
	} else if metadata.BaselineID != "" && !isRuntimeGeneratedFullBaseline(metadata) {
		addField("b", estimateJSONStringBytes(metadata.BaselineID))
	}

	if metadata.ChunkCount > 1 {
		addField("ci", estimateJSONIntBytes(int64(metadata.ChunkIndex)))
		addField("cc", estimateJSONIntBytes(int64(metadata.ChunkCount)))
		if metadata.IsFinalChunk {
			addField("fc", 4)
		}
	}

	arraySize := 2
	if len(updates) > 0 {
		arraySize += len(updates) - 1
		for _, update := range updates {
			arraySize += updateEstimator(update)
		}
	}
	addField(updateField, arraySize)
	return size
}
