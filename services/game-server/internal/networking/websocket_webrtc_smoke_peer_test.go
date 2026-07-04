package networking

import (
	"encoding/json"
	"testing"
)

func TestWebSocketSessionEnqueueWebRTCMessages(t *testing.T) {
	session := &webSocketSession{outbound: make(chan []byte, 4), sessionID: "session-test"}

	session.enqueueWebRTCAnswer("answer", "sdp")
	session.enqueueWebRTCICECandidate("audio", 2, "candidate")
	session.enqueueWebRTCReady()
	session.enqueueWebRTCFailed("error", "message")

	assertQueuedPacket(t, <-session.outbound, "webrtc_answer")
	assertQueuedPacket(t, <-session.outbound, "webrtc_ice_candidate")
	assertQueuedPacket(t, <-session.outbound, "webrtc_ready")
	assertQueuedPacket(t, <-session.outbound, "webrtc_failed")
}

func TestWebSocketSessionIgnoresICEBeforeOffer(t *testing.T) {
	session := &webSocketSession{sessionID: "session-test"}
	session.HandleWebRTCIceCandidate("audio", 1, "candidate")
	if session.webrtcSmokePeer != nil {
		t.Fatal("expected no peer to be created by ICE before offer")
	}
}

func assertQueuedPacket(t *testing.T, raw []byte, expectedType string) {
	t.Helper()
	var packet map[string]any
	if err := json.Unmarshal(raw, &packet); err != nil {
		t.Fatalf("packet did not unmarshal: %v", err)
	}
	if packet["type"] != expectedType {
		t.Fatalf("expected packet type %q, got %#v", expectedType, packet["type"])
	}
}
