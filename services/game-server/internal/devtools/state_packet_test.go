package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestWrapStatePacketPreservesServerSentMsec(t *testing.T) {
	state := game.StatePacket{
		Type:           game.PacketTypeState,
		ServerSentMsec: 123456789,
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if wrapped.ServerSentMsec != state.ServerSentMsec {
		t.Fatalf("ServerSentMsec = %d, want %d", wrapped.ServerSentMsec, state.ServerSentMsec)
	}
}
