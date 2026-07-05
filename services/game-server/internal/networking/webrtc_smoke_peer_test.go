package networking

import (
	"encoding/json"
	"testing"

	"github.com/pion/webrtc/v4"
)

type fakeWebRTCDataChannel struct {
	readyState webrtc.DataChannelState
	sentTexts  []string
	onOpen     func()
	onMessage  func(webrtc.DataChannelMessage)
	closed     bool
}

func (c *fakeWebRTCDataChannel) OnOpen(f func()) { c.onOpen = f }
func (c *fakeWebRTCDataChannel) OnMessage(f func(webrtc.DataChannelMessage)) { c.onMessage = f }
func (c *fakeWebRTCDataChannel) ReadyState() webrtc.DataChannelState { return c.readyState }
func (c *fakeWebRTCDataChannel) SendText(s string) error { c.sentTexts = append(c.sentTexts, s); return nil }
func (c *fakeWebRTCDataChannel) Close() error { c.closed = true; return nil }

type fakeWebRTCPeer struct {
	remoteDesc   webrtc.SessionDescription
	localDesc    webrtc.SessionDescription
	answer       webrtc.SessionDescription
	answerErr    error
	setRemoteErr error
	setLocalErr  error
	addICEErr    error
	channel      *fakeWebRTCDataChannel
	onICE        func(*webrtc.ICECandidate)
	closed       bool
}

func (p *fakeWebRTCPeer) SetRemoteDescription(desc webrtc.SessionDescription) error { p.remoteDesc = desc; return p.setRemoteErr }
func (p *fakeWebRTCPeer) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) { return p.answer, p.answerErr }
func (p *fakeWebRTCPeer) SetLocalDescription(desc webrtc.SessionDescription) error { p.localDesc = desc; return p.setLocalErr }
func (p *fakeWebRTCPeer) AddICECandidate(candidate webrtc.ICECandidateInit) error { return p.addICEErr }
func (p *fakeWebRTCPeer) OnICECandidate(f func(*webrtc.ICECandidate)) { p.onICE = f }
func (p *fakeWebRTCPeer) CreateDataChannel(label string, init *webrtc.DataChannelInit) (webRTCDataChannel, error) { return p.channel, nil }
func (p *fakeWebRTCPeer) Close() error { p.closed = true; return nil }

func assertSentJSONField(t *testing.T, raw string, key string, want any) {
	t.Helper()
	var packet map[string]any
	if err := json.Unmarshal([]byte(raw), &packet); err != nil {
		t.Fatalf("packet did not unmarshal: %v", err)
	}
	if packet[key] != want {
		t.Fatalf("expected %s=%#v, got %#v", key, want, packet[key])
	}
}

func TestWebRTCTransportHandleOfferBuildsAnswerAndChannel(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{answer: webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"}, channel: &fakeWebRTCDataChannel{}}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	answer, err := peer.HandleOffer("offer", "offer-sdp")
	if err != nil {
		t.Fatalf("HandleOffer returned error: %v", err)
	}
	if answer.DescriptionType != "answer" || answer.SDP != "answer-sdp" {
		t.Fatalf("unexpected answer payload: %#v", answer)
	}
	if fakePeer.remoteDesc.Type != webrtc.SDPTypeOffer || fakePeer.remoteDesc.SDP != "offer-sdp" {
		t.Fatalf("remote description not set: %#v", fakePeer.remoteDesc)
	}
	if fakePeer.localDesc.Type != webrtc.SDPTypeAnswer || fakePeer.localDesc.SDP != "answer-sdp" {
		t.Fatalf("local description not set: %#v", fakePeer.localDesc)
	}
	if fakePeer.channel == nil {
		t.Fatal("expected data channel to be created")
	}
	if fakePeer.onICE == nil {
		t.Fatal("expected ice callback to be installed")
	}
}

func TestWebRTCTransportRejectsInvalidOffer(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if _, err := peer.HandleOffer("", "sdp"); err == nil {
		t.Fatal("expected error for empty description type")
	}
	if _, err := peer.HandleOffer("offer", ""); err == nil {
		t.Fatal("expected error for empty sdp")
	}
}

func TestWebRTCTransportSendJSONRequiresOpenChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if err := peer.SendJSON(map[string]any{"type": "custom"}); err == nil {
		t.Fatal("expected error without channel")
	}

	channel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	peer.channel = channel
	if err := peer.SendJSON(map[string]any{"type": "custom", "value": "hello"}); err != nil {
		t.Fatalf("SendJSON returned error: %v", err)
	}
	if len(channel.sentTexts) != 1 {
		t.Fatalf("expected 1 sent text, got %d", len(channel.sentTexts))
	}
	assertSentJSONField(t, channel.sentTexts[0], "type", "custom")
	assertSentJSONField(t, channel.sentTexts[0], "value", "hello")
}

func TestWebRTCTransportSendSmokeRequiresOpenChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if err := peer.SendSmoke("smoke-1", "hello"); err == nil {
		t.Fatal("expected error without channel")
	}

	channel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	peer.channel = channel
	if err := peer.SendSmoke("smoke-1", "hello"); err != nil {
		t.Fatalf("SendSmoke returned error: %v", err)
	}
	if len(channel.sentTexts) != 1 {
		t.Fatalf("expected 1 sent text, got %d", len(channel.sentTexts))
	}
	assertSentJSONField(t, channel.sentTexts[0], "type", "webrtc_smoke")
	assertSentJSONField(t, channel.sentTexts[0], "smoke_id", "smoke-1")
	assertSentJSONField(t, channel.sentTexts[0], "origin", "server")
	assertSentJSONField(t, channel.sentTexts[0], "message", "hello")
}

func TestWebRTCTransportHandleChannelMessageRoutesPacketsAndInvalidJSON(t *testing.T) {
	var packets []map[string]any
	var smokePackets []map[string]any
	peer := NewWebRTCTransport(WebRTCSignalHooks{
		OnPacketReceived: func(packet map[string]any) { packets = append(packets, packet) },
		OnSmokeReceived:  func(packet map[string]any) { smokePackets = append(smokePackets, packet) },
	})

	if err := peer.handleChannelMessage([]byte(`{"type":"custom_packet","value":42}`)); err != nil {
		t.Fatalf("handleChannelMessage returned error: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("expected generic packet callback, got %d", len(packets))
	}
	if len(smokePackets) != 0 {
		t.Fatalf("expected no smoke callback, got %d", len(smokePackets))
	}

	if err := peer.handleChannelMessage([]byte(`{"type":"webrtc_smoke","smoke_id":"smoke-1","message":"hello"}`)); err != nil {
		t.Fatalf("handleChannelMessage returned error: %v", err)
	}
	if len(packets) != 2 {
		t.Fatalf("expected generic packet callback for smoke packet too, got %d", len(packets))
	}
	if len(smokePackets) != 1 {
		t.Fatalf("expected smoke packet callback, got %d", len(smokePackets))
	}

	if err := peer.handleChannelMessage([]byte("{invalid json")); err == nil {
		t.Fatal("expected invalid json error")
	}
}

func TestWebRTCTransportAddRemoteCandidateRequiresPeer(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if err := peer.AddRemoteCandidate("audio", 1, "candidate"); err == nil {
		t.Fatal("expected error without peer")
	}
}

func TestWebRTCTransportClose(t *testing.T) {
	fakePeer := &fakeWebRTCPeer{channel: &fakeWebRTCDataChannel{}}
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	peer.peer = fakePeer
	peer.channel = fakePeer.channel
	if err := peer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !fakePeer.closed || !fakePeer.channel.closed {
		t.Fatal("expected peer and channel to be closed")
	}
}
