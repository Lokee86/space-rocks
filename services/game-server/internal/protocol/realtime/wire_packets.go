package realtime

import (
	"reflect"
	"strings"
)

func WireLanePacket(candidate RealtimeLaneCandidate) map[string]any {
	switch packet := candidate.Full.(type) {
	case WorldFullPacket:
		return wireWorldFullPacket(packet)
	case WorldWireFullPacket:
		return wireWorldWireFullPacket(packet)
	case OverlayFullPacket:
		return wireOverlayFullPacket(packet)
	case OverlayWireFullPacket:
		return wireOverlayWireFullPacket(packet)
	case SessionFullPacket:
		return wireSessionFullPacket(packet)
	case SessionWireFullPacket:
		return wireSessionWireFullPacket(packet)
	case EventBatchPacket:
		return wireEventBatchPacket(packet)
	}

	if candidate.Delta != nil {
		switch packet := candidate.Delta.(type) {
		case WorldDeltaPacket:
			return wireWorldDeltaPacket(packet)
		case WorldWireDeltaPacket:
			return wireWorldWireDeltaPacket(packet)
		case OverlayLaneDelta:
			return wireOverlayDeltaPacket(packet)
		case OverlayWireLaneDelta:
			return wireOverlayWireDeltaPacket(packet)
		case SessionLaneDelta:
			return wireSessionDeltaPacket(packet)
		case SessionWireLaneDelta:
			return wireSessionWireDeltaPacket(packet)
		default:
			return wireLaneDelta(candidate.Delta)
		}
	}

	return map[string]any{}
}

func wireWorldFullPacket(packet WorldFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	wire["ships"] = wireRecords(packet.Ships)
	wire["bullets"] = wireRecords(packet.Bullets)
	wire["asteroids"] = wireRecords(packet.Asteroids)
	wire["pickups"] = wireRecords(packet.Pickups)
	return wire
}

func wireWorldWireFullPacket(packet WorldWireFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	wire["ships"] = wireRecords(packet.Ships)
	wire["bullets"] = wireRecords(packet.Bullets)
	wire["asteroids"] = wireRecords(packet.Asteroids)
	wire["pickups"] = wireRecords(packet.Pickups)
	return wire
}
func wireOverlayFullPacket(packet OverlayFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	for key, value := range wireStructToMap(packet.Receiver) {
		wire[key] = value
	}
	return wire
}

func wireSessionFullPacket(packet SessionFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	wire["players"] = wireRecords(packet.Players)
	wire["player_lifecycle"] = wireRecords(packet.PlayerLifecycle)
	wire["total_asteroids"] = packet.TotalAsteroids
	return wire
}

func wireOverlayWireFullPacket(packet OverlayWireFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	wire["self_id"] = packet.Receiver.SelfID
	wire["lives"] = packet.Receiver.Lives
	wire["score"] = packet.Receiver.Score
	wire["respawn_cooldown"] = packet.Receiver.RespawnCooldown
	wire["primary_weapon_id"] = packet.Receiver.PrimaryWeaponID
	wire["primary_ammo_policy"] = packet.Receiver.PrimaryAmmoPolicy
	wire["primary_cooldown_remaining"] = packet.Receiver.PrimaryCooldownRemaining
	wire["primary_ammo_remaining"] = packet.Receiver.PrimaryAmmoRemaining
	wire["secondary_weapon_id"] = packet.Receiver.SecondaryWeaponID
	wire["secondary_ammo_policy"] = packet.Receiver.SecondaryAmmoPolicy
	wire["secondary_cooldown_remaining"] = packet.Receiver.SecondaryCooldownRemaining
	wire["secondary_ammo_remaining"] = packet.Receiver.SecondaryAmmoRemaining
	return wire
}

func wireSessionWireFullPacket(packet SessionWireFullPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	players := make([]any, 0, len(packet.Players))
	for _, player := range packet.Players {
		players = append(players, map[string]any{
			"id":                  player.ID,
			"ship_type":          player.ShipType,
			"score":              player.Score,
			"lives":              player.Lives,
			"respawn_cooldown":   player.RespawnCooldown,
			"primary_weapon_id":   player.PrimaryWeaponID,
			"primary_ammo_policy": player.PrimaryAmmoPolicy,
			"secondary_weapon_id": player.SecondaryWeaponID,
			"secondary_ammo_policy": player.SecondaryAmmoPolicy,
			"spawn_x":            player.SpawnX,
			"spawn_y":            player.SpawnY,
		})
	}
	wire["players"] = players
	wire["player_lifecycle"] = wireRecords(packet.PlayerLifecycle)
	wire["total_asteroids"] = packet.TotalAsteroids
	return wire
}

func wireEventBatchPacket(packet EventBatchPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	wire["batch_id"] = packet.Batch.BatchID
	wire["events"] = wireEventRecords(packet.Batch.Events)
	return wire
}

func wireRecordArray(records any) any {
	items := wireRecords(records)
	if items == nil {
		return []any{}
	}
	return items
}
func wireFilteredRecordArray(records any, allowedKeys []string) any {
	items := wireRecordArray(records)
	list, ok := items.([]any)
	if !ok {
		return items
	}
	filtered := make([]any, 0, len(list))
	for _, item := range list {
		filtered = append(filtered, wireFilteredRecord(item, allowedKeys))
	}
	return filtered
}

func wireFilteredRecord(record any, allowedKeys []string) map[string]any {
	wire := wireValue(record)
	items, ok := wire.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	allowed := make(map[string]struct{}, len(allowedKeys))
	for _, key := range allowedKeys {
		allowed[key] = struct{}{}
	}
	filtered := make(map[string]any, len(allowedKeys))
	for _, key := range allowedKeys {
		if value, ok := items[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

func wireStringArray(records []string) any {
	if records == nil {
		return []string{}
	}
	return records
}

func putRecordArrayIfNonEmpty(wire map[string]any, key string, records any) {
	result := wireRecordArray(records)
	if !hasWireItems(result) {
		return
	}
	wire[key] = result
}

func putFilteredRecordArrayIfNonEmpty(wire map[string]any, key string, records any, allowedKeys []string) {
	result := wireFilteredRecordArray(records, allowedKeys)
	if !hasWireItems(result) {
		return
	}
	wire[key] = result
}

func putStringArrayIfNonEmpty(wire map[string]any, key string, records []string) {
	if len(records) == 0 {
		return
	}
	wire[key] = records
}

func putValueIfNotNil(wire map[string]any, key string, value any) {
	if value == nil {
		return
	}
	wire[key] = value
}

func hasWireItems(value any) bool {
	if value == nil {
		return false
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len() > 0
	default:
		return true
	}
}

func wireWorldDeltaPacket(packet WorldDeltaPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "ship_creates", packet.Ships.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "ship_updates", packet.Ships.Updates, []string{"id", "x", "y", "rotation", "thrusting"})
	putStringArrayIfNonEmpty(wire, "ship_deletes", packet.Ships.Deletes)
	putRecordArrayIfNonEmpty(wire, "bullet_creates", packet.Bullets.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "bullet_updates", packet.Bullets.Updates, []string{"id", "x", "y", "rotation"})
	putStringArrayIfNonEmpty(wire, "bullet_deletes", packet.Bullets.Deletes)
	putRecordArrayIfNonEmpty(wire, "asteroid_creates", packet.Asteroids.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "asteroid_updates", packet.Asteroids.Updates, []string{"id", "x", "y"})
	putStringArrayIfNonEmpty(wire, "asteroid_deletes", packet.Asteroids.Deletes)
	putRecordArrayIfNonEmpty(wire, "pickup_creates", packet.Pickups.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "pickup_updates", packet.Pickups.Updates, []string{"id", "x", "y", "age_seconds"})
	putStringArrayIfNonEmpty(wire, "pickup_deletes", packet.Pickups.Deletes)
	return wire
}

func wireWorldWireDeltaPacket(packet WorldWireDeltaPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "ship_creates", packet.Ships.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "ship_updates", packet.Ships.Updates, []string{"id", "x", "y", "rotation", "thrusting"})
	putStringArrayIfNonEmpty(wire, "ship_deletes", packet.Ships.Deletes)
	putRecordArrayIfNonEmpty(wire, "bullet_creates", packet.Bullets.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "bullet_updates", packet.Bullets.Updates, []string{"id", "x", "y", "rotation"})
	putStringArrayIfNonEmpty(wire, "bullet_deletes", packet.Bullets.Deletes)
	putRecordArrayIfNonEmpty(wire, "asteroid_creates", packet.Asteroids.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "asteroid_updates", packet.Asteroids.Updates, []string{"id", "x", "y"})
	putStringArrayIfNonEmpty(wire, "asteroid_deletes", packet.Asteroids.Deletes)
	putRecordArrayIfNonEmpty(wire, "pickup_creates", packet.Pickups.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "pickup_updates", packet.Pickups.Updates, []string{"id", "x", "y", "age_seconds"})
	putStringArrayIfNonEmpty(wire, "pickup_deletes", packet.Pickups.Deletes)
	return wire
}


func wireOverlayDeltaPacket(packet OverlayLaneDelta) map[string]any {
	wire := wireMetadataPacket(PacketTypeOverlayDelta, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "receiver_creates", packet.Receiver.Creates)
	putRecordArrayIfNonEmpty(wire, "receiver_updates", packet.Receiver.Updates)
	putStringArrayIfNonEmpty(wire, "receiver_deletes", packet.Receiver.Deletes)
	return wire
}

func wireOverlayWireDeltaPacket(packet OverlayWireLaneDelta) map[string]any {
	wire := wireMetadataPacket(PacketTypeOverlayDelta, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "receiver_creates", packet.Receiver.Creates)
	putRecordArrayIfNonEmpty(wire, "receiver_updates", packet.Receiver.Updates)
	putStringArrayIfNonEmpty(wire, "receiver_deletes", packet.Receiver.Deletes)
	return wire
}

func wireSessionDeltaPacket(packet SessionLaneDelta) map[string]any {
	wire := wireMetadataPacket(PacketTypeSessionDelta, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "players", packet.Players.Creates)
	putRecordArrayIfNonEmpty(wire, "player_session_updates", packet.Players.Updates)
	putStringArrayIfNonEmpty(wire, "player_session_deletes", packet.Players.Deletes)
	putRecordArrayIfNonEmpty(wire, "player_lifecycle", packet.PlayerLifecycle.Creates)
	putRecordArrayIfNonEmpty(wire, "player_lifecycle_updates", packet.PlayerLifecycle.Updates)
	putStringArrayIfNonEmpty(wire, "player_lifecycle_deletes", packet.PlayerLifecycle.Deletes)
	putValueIfNotNil(wire, "total_asteroids", firstSessionTotalAsteroids(packet.TotalAsteroids))
	return wire
}

func wireSessionWireDeltaPacket(packet SessionWireLaneDelta) map[string]any {
	wire := wireMetadataPacket(PacketTypeSessionDelta, packet.Metadata)
	putRecordArrayIfNonEmpty(wire, "players", packet.Players.Creates)
	putRecordArrayIfNonEmpty(wire, "player_session_updates", packet.Players.Updates)
	putStringArrayIfNonEmpty(wire, "player_session_deletes", packet.Players.Deletes)
	putRecordArrayIfNonEmpty(wire, "player_lifecycle", packet.PlayerLifecycle.Creates)
	putRecordArrayIfNonEmpty(wire, "player_lifecycle_updates", packet.PlayerLifecycle.Updates)
	putStringArrayIfNonEmpty(wire, "player_lifecycle_deletes", packet.PlayerLifecycle.Deletes)
	putValueIfNotNil(wire, "total_asteroids", firstSessionTotalAsteroids(packet.TotalAsteroids))
	return wire
}

func wireLaneDelta(delta any) map[string]any {
	switch packet := delta.(type) {
	case WorldLaneDelta:
		return map[string]any{
			"ship_creates":     wireRecords(packet.Ships.Creates),
			"ship_updates":     wireRecords(packet.Ships.Updates),
			"ship_deletes":     packet.Ships.Deletes,
			"bullet_creates":   wireRecords(packet.Bullets.Creates),
			"bullet_updates":   wireRecords(packet.Bullets.Updates),
			"bullet_deletes":   packet.Bullets.Deletes,
			"asteroid_creates": wireRecords(packet.Asteroids.Creates),
			"asteroid_updates": wireRecords(packet.Asteroids.Updates),
			"asteroid_deletes": packet.Asteroids.Deletes,
			"pickup_creates":   wireRecords(packet.Pickups.Creates),
			"pickup_updates":   wireRecords(packet.Pickups.Updates),
			"pickup_deletes":   packet.Pickups.Deletes,
		}
	case OverlayLaneDelta:
		return map[string]any{
			"receiver_creates": wireRecords(packet.Receiver.Creates),
			"receiver_updates": wireRecords(packet.Receiver.Updates),
			"receiver_deletes": packet.Receiver.Deletes,
		}
	case SessionLaneDelta:
		return map[string]any{
			"players":                 wireRecordArray(packet.Players.Creates),
			"player_session_updates":   wireRecordArray(packet.Players.Updates),
			"player_session_deletes":   wireStringArray(packet.Players.Deletes),
			"player_lifecycle":         wireRecordArray(packet.PlayerLifecycle.Creates),
			"player_lifecycle_updates": wireRecordArray(packet.PlayerLifecycle.Updates),
			"player_lifecycle_deletes": wireStringArray(packet.PlayerLifecycle.Deletes),
			"total_asteroids":          firstSessionTotalAsteroids(packet.TotalAsteroids),
		}
	default:
		return map[string]any{}
	}
}

func wireEventRecords(records []EventRecord) []any {
	items := make([]any, 0, len(records))
	for _, record := range records {
		items = append(items, wireEventRecord(record))
	}
	return items
}

func wireEventRecord(record EventRecord) map[string]any {
	wire := map[string]any{
		"event_id": record.EventID,
	}
	for key, value := range wireStructToMap(record.Event) {
		wire[key] = value
	}
	return wire
}

func firstSessionTotalAsteroids(delta RecordDelta[SessionTotalAsteroidsRecord]) any {
	if len(delta.Updates) > 0 {
		return delta.Updates[0].Count
	}
	if len(delta.Creates) > 0 {
		return delta.Creates[0].Count
	}
	return nil
}

func wireMetadataPacket(packetType string, metadata Metadata) map[string]any {
	return map[string]any{
		"type":             packetType,
		"lane":             string(metadata.Lane),
		"sequence":         metadata.Sequence,
		"baseline_id":      metadata.BaselineID,
		"snapshot_id":      metadata.SnapshotID,
		"server_sent_msec": metadata.ServerSentMsec,
		"snapshot_kind":    string(metadata.SnapshotKind),
		"chunk_index":      metadata.ChunkIndex,
		"chunk_count":      metadata.ChunkCount,
		"is_final_chunk":   metadata.IsFinalChunk,
	}
}

func wireRecords(records any) any {
	if records == nil {
		return nil
	}

	rv := reflect.ValueOf(records)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items = append(items, wireValue(rv.Index(i).Interface()))
		}
		return items
	case reflect.Map:
		items := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				continue
			}
			items[key.String()] = wireValue(iter.Value().Interface())
		}
		return items
	default:
		return wireValue(records)
	}
}

func wireStructToMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return map[string]any{}
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return map[string]any{}
	}

	wire := make(map[string]any, rv.NumField())
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)
		if !fieldValue.CanInterface() {
			continue
		}
		wire[toSnakeCase(field.Name)] = wireValue(fieldValue.Interface())
	}
	return wire
}

func wireValue(value any) any {
	if value == nil {
		return nil
	}

	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		value = rv.Interface()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return wireStructToMap(value)
	case reflect.Slice, reflect.Array:
		items := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items = append(items, wireValue(rv.Index(i).Interface()))
		}
		return items
	case reflect.Map:
		items := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				continue
			}
			items[key.String()] = wireValue(iter.Value().Interface())
		}
		return items
	default:
		return value
	}
}

func toSnakeCase(value string) string {
	if value == "" {
		return value
	}

	runes := []rune(value)
	var builder strings.Builder
	builder.Grow(len(runes) + 4)
	for i, r := range runes {
		if i > 0 && isUpper(r) && (isLower(runes[i-1]) || isDigit(runes[i-1]) || (isUpper(runes[i-1]) && i+1 < len(runes) && isLower(runes[i+1]))) {
			builder.WriteByte('_')
		}
		if isUpper(r) {
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }






