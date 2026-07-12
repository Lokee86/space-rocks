package networking

import "testing"

func TestWebSocketSessionFirstPacketLoggingIsIndependentPerMatchAndSession(t *testing.T) {
	first := &webSocketSession{}
	if !first.shouldLogFirstInputPacket("match-1") || first.shouldLogFirstInputPacket("match-1") {
		t.Fatal("input should log once per match")
	}
	if !first.shouldLogFirstRespawnPacket("match-1") || first.shouldLogFirstRespawnPacket("match-1") {
		t.Fatal("respawn should log once per match")
	}
	if !first.shouldLogFirstInputPacket("match-2") || !first.shouldLogFirstRespawnPacket("match-2") {
		t.Fatal("changing matches should reset both flags")
	}
	if first.shouldLogFirstInputPacket("match-2") || first.shouldLogFirstRespawnPacket("match-2") {
		t.Fatal("both packet types should remain logged after their first packet")
	}

	second := &webSocketSession{}
	if !second.shouldLogFirstInputPacket("match-1") || !second.shouldLogFirstRespawnPacket("match-1") {
		t.Fatal("sessions should track independently")
	}
}
