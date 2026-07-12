package networking

import (
	"sync"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestSessionContextConcurrentSnapshotsAreCoherent(t *testing.T) {
	a := rooms.NewRoom("a", rooms.RoomStateLobby, nil)
	b := rooms.NewRoom("b", rooms.RoomStateLobby, nil)
	s := &webSocketSession{}
	contexts := []SessionContext{{Room: a, RoomID: a.ID, GamePlayerID: "player-a"}, {Room: b, RoomID: b.ID, GamePlayerID: "player-b"}}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			s.mu.Lock()
			s.context = contexts[i%2]
			s.mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			got := s.sessionContext()
			if got != (SessionContext{}) && got != contexts[0] && got != contexts[1] {
				t.Errorf("mixed context: %#v", got)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			got := s.sessionContext()
			if got != (SessionContext{}) && got != contexts[0] && got != contexts[1] {
				t.Errorf("mixed context: %#v", got)
				return
			}
		}
	}()
	wg.Wait()
}

func TestSessionContextExpectedRoomMutations(t *testing.T) {
	a := rooms.NewRoom("a", rooms.RoomStateLobby, nil)
	b := rooms.NewRoom("b", rooms.RoomStateLobby, nil)
	s := &webSocketSession{}
	s.bindRoom(a)
	if s.setGamePlayerIDForRoom(b, "player") {
		t.Fatal("stale set succeeded")
	}
	if s.clearGamePlayerIDForRoom(b) {
		t.Fatal("stale clear succeeded")
	}
	if !s.setGamePlayerIDForRoom(a, "player") {
		t.Fatal("expected set")
	}
	if s.clearGamePlayerIDForRoom(b) {
		t.Fatal("stale clear succeeded")
	}
	if !s.clearGamePlayerIDForRoom(a) {
		t.Fatal("expected clear")
	}
	s.bindRoom(b)
	if s.clearRoomContextIfMatch(SessionContext{Room: a, RoomID: a.ID}) {
		t.Fatal("conditional clear matched newer context")
	}
	s.bindRoom(a)
	if !s.clearRoomContextIfMatch(SessionContext{Room: a, RoomID: a.ID}) {
		t.Fatal("expected conditional clear")
	}
	s.clearRoomContext()
	s.clearRoomContext()
	if s.sessionContext() != (SessionContext{}) {
		t.Fatal("unconditional clear not idempotent")
	}
}

func TestSessionIdentityConcurrentReadWrite(t *testing.T) {
	s := &webSocketSession{identity: NewGuestSessionIdentity()}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			s.SetAuthenticatedAccountIdentity(int64(i), "account", "display")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			got := s.SessionIdentity()
			if got.AccountID != "" && got.AccountID != "account" {
				t.Errorf("invalid identity: %#v", got)
				return
			}
		}
	}()
	wg.Wait()
}
