package realtime

import "sort"

type RecordDelta[T any] struct {
	Creates []T
	Updates []T
	Deletes []string
}

type WorldLaneDelta struct {
	Ships     RecordDelta[WorldShipRecord]
	Bullets   RecordDelta[WorldBulletRecord]
	Asteroids RecordDelta[WorldAsteroidRecord]
	Pickups   RecordDelta[WorldPickupRecord]
}

type OverlayLaneDelta struct {
	Receiver RecordDelta[OverlayReceiverRecord]
}

type SessionLaneDelta struct {
	Players         RecordDelta[SessionPlayerRecord]
	PlayerLifecycle RecordDelta[SessionLifecycleRecord]
	TotalAsteroids  RecordDelta[SessionTotalAsteroidsRecord]
}

type SessionTotalAsteroidsRecord struct {
	ID    string
	Count int
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
