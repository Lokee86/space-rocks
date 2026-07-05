package networking

import (
	"encoding/json"
	"testing"

	"github.com/pion/webrtc/v4"
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
	if session.webrtcTransport != nil {
		t.Fatal("expected no peer to be created by ICE before offer")
	}
}

func TestWebSocketSessionHandleWebRTCPacketRepliesToSmokePacket(t *testing.T) {
	session := &webSocketSession{sessionID: "session-test"}
	channel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	session.webrtcTransport = &WebRTCTransport{
		channels: map[string]webRTCDataChannel{
			webRTCGameplayChannelLaneWorld: channel,
		},
	}

	session.handleWebRTCPacket(map[string]any{
		"type":     "webrtc_smoke",
		"smoke_id": "smoke-1",
		"origin":   "client",
		"message":  "hello",
	})

	if len(channel.sentTexts) != 1 {
		t.Fatalf("expected reply packet to be sent, got %d", len(channel.sentTexts))
	}
	assertSentJSONField(t, channel.sentTexts[0], "type", "webrtc_smoke")
	assertSentJSONField(t, channel.sentTexts[0], "smoke_id", "smoke-1")
	assertSentJSONField(t, channel.sentTexts[0], "origin", "server")
	assertSentJSONField(t, channel.sentTexts[0], "message", "server reply")
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
