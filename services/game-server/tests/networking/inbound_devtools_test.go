package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
)

type placementDevtoolsSession struct{}

func (placementDevtoolsSession) CurrentSessionContext() inbound.SessionContext {
	return inbound.SessionContext{}
}

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
