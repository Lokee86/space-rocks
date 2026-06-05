package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

type placementDevtoolsSession struct{}

func (placementDevtoolsSession) CurrentRoom() *rooms.Room { return nil }
func (placementDevtoolsSession) CurrentRoomID() string { return "" }
func (placementDevtoolsSession) CurrentGamePlayerID() string { return "" }
func (placementDevtoolsSession) SessionID() string { return "" }

func TestHandlePlacementDevtoolsPacketAcceptsPickupSpawn(t *testing.T) {
	session := placementDevtoolsSession{}

	cases := []string{
		devtools.PacketTypeDebugSpawnEntity,
		devtools.PacketTypeDebugSpawnPickup,
	}

	for _, packetType := range cases {
		if !inbound.HandlePlacementDevtoolsPacket(session, "", nil, inbound.ClientPacketEnvelope{Type: packetType}) {
			t.Fatalf("expected %s to be treated as a placement devtools packet", packetType)
		}
	}
}
