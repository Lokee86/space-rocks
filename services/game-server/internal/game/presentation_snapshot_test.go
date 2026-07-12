package game

import (
	"testing"
	"time"
)

func TestGameplayPresentationSnapshotSetsLiveServerTimestamp(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	before := time.Now().UnixMilli()
	snapshot := game.GameplayPresentationSnapshot(playerID)
	after := time.Now().UnixMilli()

	if snapshot.ServerSentMsec <= 0 {
		t.Fatalf("expected positive server timestamp, got %d", snapshot.ServerSentMsec)
	}
	if snapshot.ServerSentMsec < int(before) || snapshot.ServerSentMsec > int(after) {
		t.Fatalf("expected server timestamp %d to be within [%d, %d]", snapshot.ServerSentMsec, before, after)
	}
}
