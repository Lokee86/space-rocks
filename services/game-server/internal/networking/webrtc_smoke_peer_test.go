package networking

import (
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

func TestWebRTCSmokePeerHandleOfferBuildsAnswerAndChannel(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{answer: webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"}, channel: &fakeWebRTCDataChannel{}}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{})
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

func TestWebRTCSmokePeerRejectsInvalidOffer(t *testing.T) {
	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{})
	if _, err := peer.HandleOffer("", "sdp"); err == nil {
		t.Fatal("expected error for empty description type")
	}
	if _, err := peer.HandleOffer("offer", ""); err == nil {
		t.Fatal("expected error for empty sdp")
	}
}

func TestWebRTCSmokePeerSendSmokeRequiresOpenChannel(t *testing.T) {
	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{})
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
}

func TestWebRTCSmokePeerHandleChannelMessageRoutesSmokeAndInvalidJSON(t *testing.T) {
	var smokePackets []map[string]any
	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{OnSmokeReceived: func(packet map[string]any) { smokePackets = append(smokePackets, packet) }})

	if err := peer.handleChannelMessage([]byte(`{"type":"webrtc_smoke","smoke_id":"smoke-1","message":"hello"}`)); err != nil {
		t.Fatalf("handleChannelMessage returned error: %v", err)
	}
	if len(smokePackets) != 1 {
		t.Fatalf("expected smoke packet callback, got %d", len(smokePackets))
	}

	if err := peer.handleChannelMessage([]byte("{invalid json")); err == nil {
		t.Fatal("expected invalid json error")
	}
}

func TestWebRTCSmokePeerAddRemoteCandidateRequiresPeer(t *testing.T) {
	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{})
	if err := peer.AddRemoteCandidate("audio", 1, "candidate"); err == nil {
		t.Fatal("expected error without peer")
	}
}

func TestWebRTCSmokePeerClose(t *testing.T) {
	fakePeer := &fakeWebRTCPeer{channel: &fakeWebRTCDataChannel{}}
	peer := NewWebRTCSmokePeer(WebRTCSignalHooks{})
	peer.peer = fakePeer
	peer.channel = fakePeer.channel
	if err := peer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !fakePeer.closed || !fakePeer.channel.closed {
		t.Fatal("expected peer and channel to be closed")
	}
}
