package realtime

import (
	"reflect"
	"sort"
)

type RecordDelta[T any] struct {
	Creates []T
	Updates []T
	Deletes []string
}

type FieldRecordDelta[T any] struct {
	Creates []T
	Updates []map[string]any
	Deletes []string
}

type WorldLaneDelta struct {
	Ships     FieldRecordDelta[WorldShipRecord]
	Bullets   FieldRecordDelta[WorldBulletRecord]
	Asteroids FieldRecordDelta[WorldAsteroidRecord]
	Pickups   FieldRecordDelta[WorldPickupRecord]
}

type OverlayLaneDelta struct {
	Metadata Metadata
	Receiver FieldRecordDelta[OverlayReceiverRecord]
}

type SessionLaneDelta struct {
	Metadata        Metadata
	Players         FieldRecordDelta[SessionPlayerRecord]
	PlayerLifecycle FieldRecordDelta[SessionLifecycleRecord]
	TotalAsteroids  RecordDelta[SessionTotalAsteroidsRecord]
}

type SessionTotalAsteroidsRecord struct {
	ID    string
	Count int
}

type WorldDeltaPacket struct {
	Type      string
	Metadata  Metadata
	Ships     FieldRecordDelta[WorldShipRecord]
	Bullets   FieldRecordDelta[WorldBulletRecord]
	Asteroids FieldRecordDelta[WorldAsteroidRecord]
	Pickups   FieldRecordDelta[WorldPickupRecord]
}

func CompareLaneRecordFields[T any](previous []T, current []T, recordID func(T) string, identityWireKey string) FieldRecordDelta[T] {
	previousByID := make(map[string]T, len(previous))
	for _, record := range previous {
		previousByID[recordID(record)] = record
	}

	currentByID := make(map[string]T, len(current))
	currentIDs := make([]string, 0, len(current))
	for _, record := range current {
		id := recordID(record)
		currentByID[id] = record
		currentIDs = append(currentIDs, id)
	}
	sort.Strings(currentIDs)

	previousIDs := make([]string, 0, len(previous))
	for _, record := range previous {
		previousIDs = append(previousIDs, recordID(record))
	}
	sort.Strings(previousIDs)

	delta := FieldRecordDelta[T]{}

	for _, id := range currentIDs {
		currentRecord := currentByID[id]
		previousRecord, ok := previousByID[id]
		if !ok {
			delta.Creates = append(delta.Creates, currentRecord)
			continue
		}

		previousWire := wireStructToMap(previousRecord)
		currentWire := wireStructToMap(currentRecord)
		update := map[string]any{identityWireKey: currentWire[identityWireKey]}
		if update[identityWireKey] == nil {
			update[identityWireKey] = recordID(currentRecord)
		}

		for key, currentValue := range currentWire {
			if key == identityWireKey {
				continue
			}
			if !reflect.DeepEqual(previousWire[key], currentValue) {
				update[key] = currentValue
			}
		}

		if len(update) > 1 {
			delta.Updates = append(delta.Updates, update)
		}
	}

	for _, id := range previousIDs {
		if _, ok := currentByID[id]; !ok {
			delta.Deletes = append(delta.Deletes, id)
		}
	}

	return delta
}

func CompareLaneRecords[T any](previous []T, current []T, recordID func(T) string, equal func(T, T) bool) RecordDelta[T] {
	previousByID := make(map[string]T, len(previous))
	for _, record := range previous {
		previousByID[recordID(record)] = record
	}

	currentByID := make(map[string]T, len(current))
	currentIDs := make([]string, 0, len(current))
	for _, record := range current {
		id := recordID(record)
		currentByID[id] = record
		currentIDs = append(currentIDs, id)
	}
	sort.Strings(currentIDs)

	previousIDs := make([]string, 0, len(previous))
	for _, record := range previous {
		previousIDs = append(previousIDs, recordID(record))
	}
	sort.Strings(previousIDs)

	delta := RecordDelta[T]{}

	for _, id := range currentIDs {
		currentRecord := currentByID[id]
		previousRecord, ok := previousByID[id]
		if !ok {
			delta.Creates = append(delta.Creates, currentRecord)
			continue
		}
		if !equal(previousRecord, currentRecord) {
			delta.Updates = append(delta.Updates, currentRecord)
		}
	}

	for _, id := range previousIDs {
		if _, ok := currentByID[id]; !ok {
			delta.Deletes = append(delta.Deletes, id)
		}
	}

	return delta
}

func ProjectionChanged(previous any, current any) bool {
	if previous == nil || current == nil {
		return true
	}
	return !reflect.DeepEqual(previous, current)
}

func BuildWorldDeltaPacket(previous WorldFullPacket, current WorldFullPacket) WorldDeltaPacket {
	metadata := current.Metadata
	metadata.SnapshotKind = SnapshotKind("delta")
	return WorldDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: metadata,
		Ships: CompareLaneRecordFields(previous.Ships, current.Ships,
			func(record WorldShipRecord) string { return record.ID },
			"id",
		),
		Bullets: CompareLaneRecordFields(previous.Bullets, current.Bullets,
			func(record WorldBulletRecord) string { return record.ID },
			"id",
		),
		Asteroids: CompareLaneRecordFields(previous.Asteroids, current.Asteroids,
			func(record WorldAsteroidRecord) string { return record.ID },
			"id",
		),
		Pickups: CompareLaneRecordFields(previous.Pickups, current.Pickups,
			func(record WorldPickupRecord) string { return record.ID },
			"id",
		),
	}
}

func WorldDeltaHasChanges(delta WorldDeltaPacket) bool {
	return len(delta.Ships.Creates) > 0 || len(delta.Ships.Updates) > 0 || len(delta.Ships.Deletes) > 0 ||
		len(delta.Bullets.Creates) > 0 || len(delta.Bullets.Updates) > 0 || len(delta.Bullets.Deletes) > 0 ||
		len(delta.Asteroids.Creates) > 0 || len(delta.Asteroids.Updates) > 0 || len(delta.Asteroids.Deletes) > 0 ||
		len(delta.Pickups.Creates) > 0 || len(delta.Pickups.Updates) > 0 || len(delta.Pickups.Deletes) > 0
}

func BuildOverlayDeltaPacket(previous OverlayFullPacket, current OverlayFullPacket) OverlayLaneDelta {
	previousRecords := []OverlayReceiverRecord{previous.Receiver}
	currentRecords := []OverlayReceiverRecord{current.Receiver}
	metadata := current.Metadata
	metadata.SnapshotKind = SnapshotKind("delta")
	return OverlayLaneDelta{
		Metadata: metadata,
		Receiver: CompareLaneRecordFields(previousRecords, currentRecords,
			func(record OverlayReceiverRecord) string { return record.SelfID },
			"self_id",
		),
	}
}

func OverlayDeltaHasChanges(delta OverlayLaneDelta) bool {
	return len(delta.Receiver.Creates) > 0 || len(delta.Receiver.Updates) > 0 || len(delta.Receiver.Deletes) > 0
}

func BuildSessionDeltaPacket(previous SessionFullPacket, current SessionFullPacket) SessionLaneDelta {
	previousTotal := []SessionTotalAsteroidsRecord{{ID: previous.Metadata.SnapshotID, Count: previous.TotalAsteroids}}
	currentTotal := []SessionTotalAsteroidsRecord{{ID: current.Metadata.SnapshotID, Count: current.TotalAsteroids}}
	metadata := current.Metadata
	metadata.SnapshotKind = SnapshotKind("delta")
	return SessionLaneDelta{
		Metadata: metadata,
		Players: CompareLaneRecordFields(previous.Players, current.Players,
			func(record SessionPlayerRecord) string { return record.ID },
			"id",
		),
		PlayerLifecycle: CompareLaneRecordFields(previous.PlayerLifecycle, current.PlayerLifecycle,
			func(record SessionLifecycleRecord) string { return record.PlayerID },
			"player_id",
		),
		TotalAsteroids: CompareLaneRecords(previousTotal, currentTotal,
			func(record SessionTotalAsteroidsRecord) string { return record.ID },
			func(left, right SessionTotalAsteroidsRecord) bool { return left == right },
		),
	}
}

func SessionDeltaHasChanges(delta SessionLaneDelta) bool {
	return len(delta.Players.Creates) > 0 || len(delta.Players.Updates) > 0 || len(delta.Players.Deletes) > 0 ||
		len(delta.PlayerLifecycle.Creates) > 0 || len(delta.PlayerLifecycle.Updates) > 0 || len(delta.PlayerLifecycle.Deletes) > 0 ||
		len(delta.TotalAsteroids.Creates) > 0 || len(delta.TotalAsteroids.Updates) > 0 || len(delta.TotalAsteroids.Deletes) > 0
}



