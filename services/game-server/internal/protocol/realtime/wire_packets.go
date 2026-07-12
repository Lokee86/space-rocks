package realtime

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime/quantize"
)

type AsteroidWireDeltaPacket struct {
	Type            string `json:"type"`
	Metadata        Metadata
	AsteroidCreates []WorldAsteroidWireRecord
	AsteroidUpdates []map[string]any `json:"asteroid_updates"`
	AsteroidDeletes []string
}

type BulletWireDeltaPacket struct {
	Type          string `json:"type"`
	Metadata      Metadata
	BulletCreates []WorldBulletWireRecord
	BulletUpdates []map[string]any `json:"bullet_updates"`
	BulletDeletes []string
}

func WireLanePacket(candidate RealtimeLaneCandidate) (map[string]any, error) {
	if err := ValidateRealtimeLanePayload(candidate.Payload); err != nil {
		return nil, fmt.Errorf("wire lane packet lane=%q family=%q: %w", candidate.Lane(), candidate.PacketFamily(), err)
	}
	wire := candidate.Payload.WirePacket()
	if candidate.MatchID != "" {
		wire["match_id"] = candidate.MatchID
	}
	if err := validateWireLaneMap(candidate, wire); err != nil {
		return nil, err
	}
	return wire, nil
}

func validateWireLaneMap(candidate RealtimeLaneCandidate, wire map[string]any) error {
	if len(wire) == 0 {
		return fmt.Errorf("wire lane packet lane=%q family=%q: serializer returned empty map", candidate.Lane(), candidate.PacketFamily())
	}
	rawType, ok := wire["type"]
	if !ok {
		return fmt.Errorf("wire lane packet lane=%q family=%q: missing type", candidate.Lane(), candidate.PacketFamily())
	}
	wireType, ok := rawType.(string)
	if !ok || strings.TrimSpace(wireType) == "" {
		return fmt.Errorf("wire lane packet lane=%q family=%q: invalid type %v", candidate.Lane(), candidate.PacketFamily(), rawType)
	}
	if wireType != candidate.PacketFamily() {
		return fmt.Errorf("wire lane packet lane=%q family=%q: wire type %q does not match payload family", candidate.Lane(), candidate.PacketFamily(), wireType)
	}
	return nil
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
			"id":                    player.ID,
			"ship_type":             player.ShipType,
			"score":                 player.Score,
			"lives":                 player.Lives,
			"respawn_cooldown":      player.RespawnCooldown,
			"primary_weapon_id":     player.PrimaryWeaponID,
			"primary_ammo_policy":   player.PrimaryAmmoPolicy,
			"secondary_weapon_id":   player.SecondaryWeaponID,
			"secondary_ammo_policy": player.SecondaryAmmoPolicy,
			"spawn_x":               player.SpawnX,
			"spawn_y":               player.SpawnY,
		})
	}
	wire["players"] = players
	wire["player_lifecycle"] = wireRecords(packet.PlayerLifecycle)
	wire["total_asteroids"] = packet.TotalAsteroids
	return wire
}

func wireEventBatchPacket(packet EventBatchPacket) map[string]any {
	return map[string]any{
		"type":             packet.Type,
		"sequence":         packet.Metadata.Sequence,
		"server_sent_msec": packet.Metadata.ServerSentMsec,
		"batch_id":         packet.Batch.BatchID,
		"events":           wireEventRecords(packet.Batch.Events),
	}
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
	putFilteredRecordArrayIfNonEmpty(wire, "asteroid_updates", packet.Asteroids.Updates, []string{"id", "x", "y"})
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
	putFilteredRecordArrayIfNonEmpty(wire, "asteroid_updates", packet.Asteroids.Updates, []string{"id", "x", "y"})
	putRecordArrayIfNonEmpty(wire, "pickup_creates", packet.Pickups.Creates)
	putFilteredRecordArrayIfNonEmpty(wire, "pickup_updates", packet.Pickups.Updates, []string{"id", "x", "y", "age_seconds"})
	putStringArrayIfNonEmpty(wire, "pickup_deletes", packet.Pickups.Deletes)
	return wire
}

func wireAsteroidWireDeltaPacket(packet AsteroidWireDeltaPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	if packet.Metadata.Lane == LaneAsteroidsLifecycle {
		putRecordArrayIfNonEmpty(wire, "asteroid_creates", packet.AsteroidCreates)
		putStringArrayIfNonEmpty(wire, "asteroid_deletes", packet.AsteroidDeletes)
		return wire
	}
	putFilteredRecordArrayIfNonEmpty(wire, "asteroid_updates", packet.AsteroidUpdates, []string{"id", "x", "y"})
	return wire
}

func wireBulletWireDeltaPacket(packet BulletWireDeltaPacket) map[string]any {
	wire := wireMetadataPacket(packet.Type, packet.Metadata)
	if packet.Metadata.Lane == LaneBulletsLifecycle {
		putRecordArrayIfNonEmpty(wire, "bullet_creates", packet.BulletCreates)
		putStringArrayIfNonEmpty(wire, "bullet_deletes", packet.BulletDeletes)
		return wire
	}
	putFilteredRecordArrayIfNonEmpty(wire, "bullet_updates", packet.BulletUpdates, []string{"id", "x", "y", "rotation"})
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

func wireEventRecords(records []EventRecord) []any {
	items := make([]any, 0, len(records))
	for _, record := range records {
		items = append(items, wireEventRecord(record))
	}
	return items
}

func wireEventRecord(record EventRecord) map[string]any {
	switch record.Event.Type {
	case "bullet_blast":
		return wireBulletBlastEventRecord(record)
	case "ship_death":
		return wireShipDeathEventRecord(record)
	case "damage_applied":
		return wireDamageAppliedEventRecord(record)
	case "damage_over_time_started":
		return wireDamageOverTimeStartedEventRecord(record)
	case "damage_over_time_tick":
		return wireDamageOverTimeTickEventRecord(record)
	case "radial_effect_started":
		return wireRadialEffectStartedEventRecord(record)
	case "pickup_collected":
		return wirePickupCollectedEventRecord(record)
	case "pickup_effect_applied":
		return wirePickupEffectAppliedEventRecord(record)
	case "pickup_expired":
		return wirePickupExpiredEventRecord(record)
	case "pickup_dropped":
		return wirePickupDroppedEventRecord(record)
	}

	wire := map[string]any{
		"event_id": record.EventID,
	}
	for key, value := range wireStructToMap(record.Event) {
		wire[key] = value
	}
	return wire
}

func wireBulletBlastEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id": record.EventID,
		"type":     record.Event.Type,
		"x":        wireQuantizedEventFloat("event.bullet_blast.x", record.Event.X),
		"y":        wireQuantizedEventFloat("event.bullet_blast.y", record.Event.Y),
	}
}

func wireShipDeathEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":      record.EventID,
		"type":          record.Event.Type,
		"player_id":     record.Event.PlayerID,
		"lives":         record.Event.Lives,
		"respawn_delay": wireQuantizedEventFloat("event.ship_death.respawn_delay", record.Event.RespawnDelay),
		"x":             wireQuantizedEventFloat("event.ship_death.x", record.Event.X),
		"y":             wireQuantizedEventFloat("event.ship_death.y", record.Event.Y),
	}
}

func wireDamageAppliedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"source_type": record.Event.SourceType,
		"source_id":   record.Event.SourceID,
		"effect_type": record.Event.EffectType,
		"amount":      record.Event.Amount,
		"x":           wireQuantizedEventFloat("event.damage_applied.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.damage_applied.y", record.Event.Y),
	}
}

func wireDamageOverTimeStartedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"source_type": record.Event.SourceType,
		"source_id":   record.Event.SourceID,
		"effect_type": record.Event.EffectType,
		"amount":      record.Event.Amount,
	}
}

func wireDamageOverTimeTickEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"source_type": record.Event.SourceType,
		"source_id":   record.Event.SourceID,
		"effect_type": record.Event.EffectType,
		"amount":      record.Event.Amount,
		"x":           wireQuantizedEventFloat("event.damage_over_time_tick.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.damage_over_time_tick.y", record.Event.Y),
	}
}

func wireRadialEffectStartedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"source_type": record.Event.SourceType,
		"source_id":   record.Event.SourceID,
		"effect_type": record.Event.EffectType,
		"x":           wireQuantizedEventFloat("event.radial_effect_started.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.radial_effect_started.y", record.Event.Y),
	}
}

func wirePickupCollectedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"player_id":   record.Event.PlayerID,
		"pickup_id":   record.Event.PickupID,
		"pickup_type": record.Event.PickupType,
		"x":           wireQuantizedEventFloat("event.pickup_collected.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.pickup_collected.y", record.Event.Y),
	}
}

func wirePickupEffectAppliedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"player_id":   record.Event.PlayerID,
		"pickup_id":   record.Event.PickupID,
		"pickup_type": record.Event.PickupType,
		"effect_type": record.Event.EffectType,
		"amount":      record.Event.Amount,
		"lives_after": record.Event.LivesAfter,
	}
}

func wirePickupExpiredEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"pickup_id":   record.Event.PickupID,
		"pickup_type": record.Event.PickupType,
		"x":           wireQuantizedEventFloat("event.pickup_expired.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.pickup_expired.y", record.Event.Y),
	}
}

func wirePickupDroppedEventRecord(record EventRecord) map[string]any {
	return map[string]any{
		"event_id":    record.EventID,
		"type":        record.Event.Type,
		"pickup_id":   record.Event.PickupID,
		"pickup_type": record.Event.PickupType,
		"source_type": record.Event.SourceType,
		"source_id":   record.Event.SourceID,
		"table_id":    record.Event.TableID,
		"x":           wireQuantizedEventFloat("event.pickup_dropped.x", record.Event.X),
		"y":           wireQuantizedEventFloat("event.pickup_dropped.y", record.Event.Y),
	}
}

func wireQuantizedEventFloat(fieldPath string, value float64) any {
	policy, ok := quantize.LookupPolicy(fieldPath)
	if !ok {
		return value
	}
	encoded, err := quantize.EncodeFloat(policy, value)
	if err != nil {
		return value
	}
	return encoded
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
	wire := map[string]any{
		"type":             packetType,
		"sequence":         metadata.Sequence,
		"server_sent_msec": metadata.ServerSentMsec,
	}
	if !isRuntimePacketType(packetType) {
		wire["lane"] = string(metadata.Lane)
		wire["baseline_id"] = metadata.BaselineID
		wire["snapshot_id"] = metadata.SnapshotID
		wire["snapshot_kind"] = string(metadata.SnapshotKind)
		wire["chunk_index"] = metadata.ChunkIndex
		wire["chunk_count"] = metadata.ChunkCount
		wire["is_final_chunk"] = metadata.IsFinalChunk
		return wire
	}

	if metadata.SnapshotKind == SnapshotKind("delta") {
		if baselineSequence, ok := runtimeBaselineSequence(metadata.Lane, metadata.BaselineID); ok {
			wire["baseline_sequence"] = baselineSequence
		} else if metadata.BaselineID != "" {
			wire["baseline_id"] = metadata.BaselineID
		}
	} else if metadata.BaselineID != "" && !isRuntimeGeneratedFullBaseline(metadata) {
		wire["baseline_id"] = metadata.BaselineID
	}

	if metadata.ChunkCount > 1 {
		wire["chunk_index"] = metadata.ChunkIndex
		wire["chunk_count"] = metadata.ChunkCount
		if metadata.IsFinalChunk {
			wire["is_final_chunk"] = true
		}
	}
	return wire
}

func isRuntimePacketType(packetType string) bool {
	switch packetType {
	case PacketFamilyWorldFull, PacketTypeWorldDelta, PacketFamilyOverlayFull, PacketTypeOverlayDelta,
		PacketFamilySessionFull, PacketTypeSessionDelta, PacketFamilyAsteroidDelta, PacketFamilyBulletDelta:
		return true
	default:
		return false
	}
}

func runtimeBaselineSequence(lane Lane, baselineID string) (int, bool) {
	prefix := string(lane) + "-baseline-"
	if !strings.HasPrefix(baselineID, prefix) {
		return 0, false
	}
	sequence, err := strconv.Atoi(strings.TrimPrefix(baselineID, prefix))
	if err != nil {
		return 0, false
	}
	return sequence, true
}

func isRuntimeGeneratedFullBaseline(metadata Metadata) bool {
	if metadata.BaselineID == "" {
		return true
	}
	sequence, ok := runtimeBaselineSequence(metadata.Lane, metadata.BaselineID)
	if !ok {
		return false
	}
	return sequence == metadata.Sequence
}
