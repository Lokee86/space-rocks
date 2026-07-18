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
			webRTCChannelLaneWorld: channel,
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

func TestWebSocketSessionUnexpectedWebRTCChannelCloseClearsOnlyTransport(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	t.Cleanup(func() { newWebRTCPeerConnection = oldFactory })
	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	session := &webSocketSession{sessionID: "session-test"}
	session.HandleWebRTCOffer("offer", "offer-sdp")
	if session.webRTCTransportSnapshot() == nil {
		t.Fatal("expected WebRTC transport after offer")
	}

	channel := fakePeer.channels["sr.world"]
	channel.readyState = webrtc.DataChannelStateOpen
	channel.onOpen()
	channel.onClose()

	if session.webRTCTransportSnapshot() != nil {
		t.Fatal("expected channel close to clear only the WebRTC transport")
	}
	if fakePeer.closed != true {
		t.Fatal("expected the closed WebRTC peer to be closed")
	}
}

func TestWebSocketSessionLaterWebRTCOfferReplacesExistingTransport(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	t.Cleanup(func() { newWebRTCPeerConnection = oldFactory })
	firstPeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "first-answer"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	secondPeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "second-answer"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	peers := []*fakeWebRTCPeer{firstPeer, secondPeer}
	newWebRTCPeerConnection = func() (webRTCPeer, error) {
		peer := peers[0]
		peers = peers[1:]
		return peer, nil
	}

	session := &webSocketSession{sessionID: "session-test"}
	session.HandleWebRTCOffer("offer", "first-offer")
	firstTransport := session.webRTCTransportSnapshot()
	session.HandleWebRTCOffer("offer", "second-offer")
	secondTransport := session.webRTCTransportSnapshot()

	if firstTransport == nil || secondTransport == nil || firstTransport == secondTransport {
		t.Fatal("expected replacement to install a fresh WebRTC transport")
	}
	if !firstPeer.closed {
		t.Fatal("expected the existing WebRTC peer to close before replacement")
	}
	if secondPeer.remoteDesc.SDP != "second-offer" {
		t.Fatalf("expected replacement offer to reach fresh peer, got %q", secondPeer.remoteDesc.SDP)
	}
}

func TestWebSocketSessionIgnoresToolingPacketsAtSessionBoundary(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	t.Cleanup(func() { newWebRTCPeerConnection = oldFactory })
	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	session := &webSocketSession{sessionID: "session-test"}
	session.HandleWebRTCOffer("offer", "offer-sdp")
	tooling := fakePeer.channels["sr.tooling"]
	tooling.readyState = webrtc.DataChannelStateOpen
	tooling.onOpen()
	tooling.onMessage(webrtc.DataChannelMessage{Data: []byte(`{"type":"webrtc_smoke","smoke_id":"tooling-smoke"}`)})

	if len(tooling.sentTexts) != 0 {
		t.Fatal("expected tooling packet to stay out of session smoke handling")
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
