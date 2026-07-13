package health

import (
	"sync"
	"testing"
	"time"
)

func TestStateSnapshotAndTransitions(t *testing.T) {
	started := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	state := NewState("instance-1", "v1", "test", started)
	state.MarkReady()
	state.IncBatchesReceived()
	state.IncEventsAccepted()
	state.IncEventsRejected()
	state.IncEventsRedacted()
	state.IncDuplicatesSuppressed()
	state.IncStorageFailures()
	state.IncRotations()
	state.IncRetentionDeletions()
	state.IncQueryFailures()
	state.IncDiagnosticBundlesCreated()
	snapshot := state.Snapshot()
	if snapshot.ServiceInstanceID != "instance-1" || !snapshot.Ready || snapshot.Stopping || !snapshot.StartedAt.Equal(started) {
		t.Fatalf("unexpected identity: %+v", snapshot)
	}
	if snapshot.BatchesReceived != 1 || snapshot.EventsAccepted != 1 || snapshot.EventsRejected != 1 || snapshot.EventsRedacted != 1 || snapshot.DuplicatesSuppressed != 1 || snapshot.StorageFailures != 1 || snapshot.Rotations != 1 || snapshot.RetentionDeletions != 1 || snapshot.QueryFailures != 1 || snapshot.DiagnosticBundlesCreated != 1 {
		t.Fatalf("unexpected counters: %+v", snapshot)
	}
	state.MarkStopping()
	if snapshot := state.Snapshot(); snapshot.Ready || !snapshot.Stopping {
		t.Fatalf("unexpected stopping state: %+v", snapshot)
	}
}

func TestStateCountByAmount(t *testing.T) {
	state := NewState("i", "v", "e", time.Time{})
	state.AddBatchesReceived(2)
	state.AddEventsAccepted(4)
	state.AddEventsRejected(3)
	state.AddEventsRedacted(2)
	state.AddQueryFailures(5)
	snapshot := state.Snapshot()
	if snapshot.BatchesReceived != 2 || snapshot.EventsAccepted != 4 || snapshot.EventsRejected != 3 || snapshot.EventsRedacted != 2 || snapshot.QueryFailures != 5 {
		t.Fatalf("unexpected counters: %+v", snapshot)
	}
}

func TestStateConcurrentCounters(t *testing.T) {
	state := NewState("i", "v", "e", time.Time{})
	const workers = 20
	const increments = 100
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			for range increments {
				state.IncEventsAccepted()
			}
		}()
	}
	wait.Wait()
	if got := state.Snapshot().EventsAccepted; got != workers*increments {
		t.Fatalf("accepted = %d", got)
	}
}
