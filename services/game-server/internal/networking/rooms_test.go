package networking

import (
	"testing"
	"time"
)

func TestRoomManagerUsesDefaultRoomForBlankID(t *testing.T) {
	manager := NewRoomManager()
	defer manager.StopAll()

	defaultRoom := manager.DefaultRoom()
	blankRoom := manager.GetOrCreate("")
	spaceRoom := manager.GetOrCreate("   ")

	if blankRoom != defaultRoom {
		t.Fatal("expected blank room id to use default room")
	}
	if spaceRoom != defaultRoom {
		t.Fatal("expected whitespace room id to use default room")
	}
}

func TestRoomManagerCreatesSeparateRoomsByID(t *testing.T) {
	manager := NewRoomManager()
	defer manager.StopAll()

	first := manager.GetOrCreate("abc")
	again := manager.GetOrCreate("abc")
	second := manager.GetOrCreate("xyz")

	if first != again {
		t.Fatal("expected same room id to return same room")
	}
	if first == second {
		t.Fatal("expected different room ids to return different rooms")
	}
	if first.Game == second.Game {
		t.Fatal("expected different rooms to own different games")
	}
}

func TestRoomManagerCleansUpEmptyRoomAfterGracePeriod(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	if room == nil {
		t.Fatal("expected room")
	}
	leave()

	if !waitUntil(200*time.Millisecond, func() bool {
		return manager.GetOrCreate("abc") != room
	}) {
		t.Fatal("expected empty room to be cleaned up after grace period")
	}
}

func TestRoomManagerDoesNotCleanUpActiveRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	defer leave()

	time.Sleep(30 * time.Millisecond)
	if manager.GetOrCreate("abc") != room {
		t.Fatal("expected active room to stay alive")
	}
}

func TestRoomManagerCancelsCleanupWhenRoomIsRejoined(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(30 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	leave()

	rejoined, leaveRejoined := manager.Join("abc")
	defer leaveRejoined()
	if rejoined != room {
		t.Fatal("expected reconnect during grace period to reuse room")
	}

	time.Sleep(60 * time.Millisecond)
	if manager.GetOrCreate("abc") != room {
		t.Fatal("expected rejoined room to survive canceled cleanup")
	}
}

func waitUntil(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(time.Millisecond)
	}

	return condition()
}
