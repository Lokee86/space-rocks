package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestInboundSessionAdapterEnqueueResyncRequestUsesSuppliedContext(t *testing.T) {
	oldRoom := rooms.NewRoom("old-room", rooms.RoomStateLobby, nil)
	newRoom := rooms.NewRoom("new-room", rooms.RoomStateLobby, nil)
	session := &webSocketSession{
		context:        SessionContext{Room: newRoom, RoomID: newRoom.ID, GamePlayerID: "new-player"},
		resyncRequests: make(chan queuedResyncRequest, 1),
	}
	adapter := newInboundSessionAdapter(session)
	supplied := inbound.SessionContext{Room: oldRoom, RoomID: oldRoom.ID, GamePlayerID: "old-player"}
	request := realtime.ResyncRequest{MatchID: "supplied-match", Lane: realtime.LaneWorld}

	if !adapter.EnqueueResyncRequest(supplied, request) {
		t.Fatal("expected resync request to be queued")
	}
	queued := <-session.resyncRequests
	if queued.RoomID != supplied.RoomID || queued.ReceiverID != supplied.GamePlayerID || queued.MatchID != request.MatchID {
		t.Fatalf("expected supplied context/request, got %#v", queued)
	}
}
