package packetmetrics

import (
	"testing"
	"time"
)

func TestNewGameplayPresentationPacketMetrics(t *testing.T) {
	contributors := GameplayPacketContributors{
		RoomState:       "in_game",
		Players:         3,
		PlayerSessions:  4,
		PlayerLifecycle: 2,
		Asteroids:       5,
		Bullets:         6,
		Pickups:         7,
		Enemies:         8,
		Events:          9,
		TotalAsteroids:  10,
	}
	buildDuration := 13 * time.Millisecond
	encodeDuration := 17 * time.Millisecond

	metrics := NewGameplayPresentationPacketMetrics(1200, contributors, buildDuration, encodeDuration)

	if metrics.PacketSize != 1200 {
		t.Fatalf("expected packet size %d, got %d", 1200, metrics.PacketSize)
	}

	if metrics.Contributors != contributors {
		t.Fatalf("expected contributors to be preserved, got %#v", metrics.Contributors)
	}

	if metrics.BuildDuration != buildDuration {
		t.Fatalf("expected build duration %v, got %v", buildDuration, metrics.BuildDuration)
	}

	if metrics.EncodeDuration != encodeDuration {
		t.Fatalf("expected encode duration %v, got %v", encodeDuration, metrics.EncodeDuration)
	}
}
