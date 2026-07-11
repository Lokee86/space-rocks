package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func mustWireLanePacket(t *testing.T, candidate RealtimeLaneCandidate) map[string]any {
	t.Helper()
	wire, err := WireLanePacket(candidate)
	if err != nil {
		t.Fatalf("wire lane packet: %v", err)
	}
	return wire
}

func mustEncodeLanePacket(t *testing.T, candidate RealtimeLaneCandidate) ([]byte, int) {
	t.Helper()
	encoded, bytes, err := encodeLanePacket(candidate)
	if err != nil {
		t.Fatalf("encode lane packet: %v", err)
	}
	return encoded, bytes
}

func mustEncodeLanePacketUnchecked(t *testing.T, candidate RealtimeLaneCandidate) ([]byte, int) {
	t.Helper()
	encoded, bytes, err := encodeLanePacketUnchecked(candidate)
	if err != nil {
		t.Fatalf("encode lane packet unchecked: %v", err)
	}
	return encoded, bytes
}

func mustBuildActiveRealtimeResult(t *testing.T, snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ActiveRealtimeResult {
	t.Helper()
	result, err := BuildActiveRealtimeResult(snapshot, state)
	if err != nil {
		t.Fatalf("build active realtime result: %v", err)
	}
	return result
}

func mustBuildShadowRealtimeResult(t *testing.T, snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) ShadowRealtimeResult {
	t.Helper()
	result, err := BuildShadowRealtimeResult(snapshot, state)
	if err != nil {
		t.Fatalf("build shadow realtime result: %v", err)
	}
	return result
}
