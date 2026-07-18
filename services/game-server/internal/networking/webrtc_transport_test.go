package networking

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pion/webrtc/v4"
)

type fakeWebRTCDataChannel struct {
	readyState webrtc.DataChannelState
	sentTexts  []string
	sendErr    error
	onOpen     func()
	onClose    func()
	onMessage  func(webrtc.DataChannelMessage)
	closed     bool
}

func (c *fakeWebRTCDataChannel) OnOpen(f func())                             { c.onOpen = f }
func (c *fakeWebRTCDataChannel) OnClose(f func())                            { c.onClose = f }
func (c *fakeWebRTCDataChannel) OnMessage(f func(webrtc.DataChannelMessage)) { c.onMessage = f }
func (c *fakeWebRTCDataChannel) ReadyState() webrtc.DataChannelState         { return c.readyState }
func (c *fakeWebRTCDataChannel) SendText(s string) error {
	if c.sendErr != nil {
		return c.sendErr
	}
	c.sentTexts = append(c.sentTexts, s)
	return nil
}
func (c *fakeWebRTCDataChannel) Close() error { c.closed = true; return nil }

type fakeWebRTCPeer struct {
	remoteDesc   webrtc.SessionDescription
	localDesc    webrtc.SessionDescription
	answer       webrtc.SessionDescription
	answerErr    error
	setRemoteErr error
	setLocalErr  error
	addICEErr    error
	channels     map[string]*fakeWebRTCDataChannel
	created      []fakeWebRTCDataChannelCreateSpec
	onICE        func(*webrtc.ICECandidate)
	closed       bool
}

func (p *fakeWebRTCPeer) SetRemoteDescription(desc webrtc.SessionDescription) error {
	p.remoteDesc = desc
	return p.setRemoteErr
}
func (p *fakeWebRTCPeer) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) {
	return p.answer, p.answerErr
}
func (p *fakeWebRTCPeer) SetLocalDescription(desc webrtc.SessionDescription) error {
	p.localDesc = desc
	return p.setLocalErr
}
func (p *fakeWebRTCPeer) AddICECandidate(candidate webrtc.ICECandidateInit) error { return p.addICEErr }
func (p *fakeWebRTCPeer) OnICECandidate(f func(*webrtc.ICECandidate))             { p.onICE = f }
func (p *fakeWebRTCPeer) CreateDataChannel(label string, init *webrtc.DataChannelInit) (webRTCDataChannel, error) {
	if p.channels == nil {
		p.channels = make(map[string]*fakeWebRTCDataChannel)
	}
	channel := p.channels[label]
	if channel == nil {
		channel = &fakeWebRTCDataChannel{}
		p.channels[label] = channel
	}
	p.created = append(p.created, fakeWebRTCDataChannelCreateSpec{
		Label:          label,
		Ordered:        init != nil && init.Ordered != nil && *init.Ordered,
		Negotiated:     init != nil && init.Negotiated != nil && *init.Negotiated,
		ID:             initIDValue(init),
		MaxRetransmits: initMaxRetransmitsValue(init),
	})
	return channel, nil
}
func (p *fakeWebRTCPeer) Close() error { p.closed = true; return nil }

type fakeWebRTCDataChannelCreateSpec struct {
	Label          string
	Ordered        bool
	Negotiated     bool
	ID             uint16
	MaxRetransmits *uint16
}

func initIDValue(init *webrtc.DataChannelInit) uint16 {
	if init == nil || init.ID == nil {
		return 0
	}
	return *init.ID
}

func initMaxRetransmitsValue(init *webrtc.DataChannelInit) *uint16 {
	if init == nil || init.MaxRetransmits == nil {
		return nil
	}
	value := *init.MaxRetransmits
	return &value
}

func uint16Ptr(value uint16) *uint16 {
	return &value
}

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

func TestWebRTCTransportPeerConnectionUsesDefaultConfigPath(t *testing.T) {
	oldBuilder := newWebRTCPeerConnectionAPI
	defer func() { newWebRTCPeerConnectionAPI = oldBuilder }()

	var captured WebRTCTransportConfig
	newWebRTCPeerConnectionAPI = func(config WebRTCTransportConfig) (*webrtc.API, error) {
		captured = config
		return webrtc.NewAPI(), nil
	}

	restore := SetWebRTCTransportConfigForTests(WebRTCTransportConfig{})
	defer restore()

	peer, err := newWebRTCPeerConnection()
	if err != nil {
		t.Fatalf("newWebRTCPeerConnection returned error: %v", err)
	}
	if peer == nil {
		t.Fatal("expected peer to be created")
	}
	if !reflect.DeepEqual(captured, WebRTCTransportConfig{}) {
		t.Fatalf("expected empty config, got %#v", captured)
	}
}

func TestWebRTCTransportPeerConnectionAcceptsAdvertisedIPsAndUDPPortRange(t *testing.T) {
	oldBuilder := newWebRTCPeerConnectionAPI
	defer func() { newWebRTCPeerConnectionAPI = oldBuilder }()

	var captured WebRTCTransportConfig
	newWebRTCPeerConnectionAPI = func(config WebRTCTransportConfig) (*webrtc.API, error) {
		captured = config
		return webrtc.NewAPI(), nil
	}

	restore := SetWebRTCTransportConfigForTests(WebRTCTransportConfig{
		AdvertisedIPs: []string{"198.51.100.10", "203.0.113.25"},
		UDPPortMin:    40000,
		UDPPortMax:    40010,
	})
	defer restore()

	peer, err := newWebRTCPeerConnection()
	if err != nil {
		t.Fatalf("newWebRTCPeerConnection returned error: %v", err)
	}
	if peer == nil {
		t.Fatal("expected peer to be created")
	}
	if !reflect.DeepEqual(captured.AdvertisedIPs, []string{"198.51.100.10", "203.0.113.25"}) {
		t.Fatalf("expected advertised IPs to be forwarded, got %#v", captured.AdvertisedIPs)
	}
	if captured.UDPPortMin != 40000 || captured.UDPPortMax != 40010 {
		t.Fatalf("expected UDP port range to be forwarded, got %#v", captured)
	}
}

func TestWebRTCTransportHandleOfferBuildsAnswerAndChannels(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
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
	if len(fakePeer.created) != 9 {
		t.Fatalf("expected 9 data channels to be created, got %d", len(fakePeer.created))
	}
	expected := map[string]fakeWebRTCDataChannelCreateSpec{
		"sr.world":               {Label: "sr.world", Ordered: true, Negotiated: true, ID: 1, MaxRetransmits: nil},
		"sr.overlay":             {Label: "sr.overlay", Ordered: true, Negotiated: true, ID: 2, MaxRetransmits: nil},
		"sr.session":             {Label: "sr.session", Ordered: true, Negotiated: true, ID: 3, MaxRetransmits: nil},
		"sr.event":               {Label: "sr.event", Ordered: true, Negotiated: true, ID: 4, MaxRetransmits: nil},
		"sr.asteroids":           {Label: "sr.asteroids", Ordered: false, Negotiated: true, ID: 5, MaxRetransmits: uint16Ptr(0)},
		"sr.bullets":             {Label: "sr.bullets", Ordered: false, Negotiated: true, ID: 6, MaxRetransmits: uint16Ptr(0)},
		"sr.asteroids.lifecycle": {Label: "sr.asteroids.lifecycle", Ordered: true, Negotiated: true, ID: 7, MaxRetransmits: nil},
		"sr.bullets.lifecycle":   {Label: "sr.bullets.lifecycle", Ordered: true, Negotiated: true, ID: 8, MaxRetransmits: nil},
		"sr.tooling":             {Label: "sr.tooling", Ordered: true, Negotiated: true, ID: 9, MaxRetransmits: nil},
	}
	for _, created := range fakePeer.created {
		want, ok := expected[created.Label]
		if !ok {
			t.Fatalf("unexpected data channel created: %#v", created)
		}
		if created.Label != want.Label {
			t.Fatalf("unexpected data channel label for %s: got %q want %q", created.Label, created.Label, want.Label)
		}
		if created.Ordered != want.Ordered {
			t.Fatalf("unexpected data channel ordered flag for %s: got %t want %t", created.Label, created.Ordered, want.Ordered)
		}
		if created.Negotiated != want.Negotiated {
			t.Fatalf("unexpected data channel negotiated flag for %s: got %t want %t", created.Label, created.Negotiated, want.Negotiated)
		}
		if created.ID != want.ID {
			t.Fatalf("unexpected data channel id for %s: got %d want %d", created.Label, created.ID, want.ID)
		}
		if want.MaxRetransmits == nil {
			if created.MaxRetransmits != nil {
				t.Fatalf("unexpected max retransmits for %s: got %d want nil", created.Label, *created.MaxRetransmits)
			}
		} else {
			if created.MaxRetransmits == nil {
				t.Fatalf("unexpected max retransmits for %s: got nil want %d", created.Label, *want.MaxRetransmits)
			}
			if *created.MaxRetransmits != *want.MaxRetransmits {
				t.Fatalf("unexpected max retransmits for %s: got %d want %d", created.Label, *created.MaxRetransmits, *want.MaxRetransmits)
			}
		}
		delete(expected, created.Label)
	}
	if len(expected) != 0 {
		t.Fatalf("missing created channels: %#v", expected)
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
	peer.channels = map[string]webRTCDataChannel{
		webRTCChannelLaneWorld: channel,
	}
	if err := peer.SendJSON(map[string]any{"type": "custom", "value": "hello"}); err != nil {
		t.Fatalf("SendJSON returned error: %v", err)
	}
	if len(channel.sentTexts) != 1 {
		t.Fatalf("expected 1 sent text, got %d", len(channel.sentTexts))
	}
	assertSentJSONField(t, channel.sentTexts[0], "type", "custom")
	assertSentJSONField(t, channel.sentTexts[0], "value", "hello")
}
func TestWebRTCTransportSendEncodedJSONRequiresOpenChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if err := peer.SendEncodedJSON([]byte(`{"type":"custom"}`)); err == nil {
		t.Fatal("expected error without channel")
	}

	channel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateClosed}
	peer.channels = map[string]webRTCDataChannel{
		webRTCChannelLaneWorld: channel,
	}
	if err := peer.SendEncodedJSON([]byte(`{"type":"custom"}`)); err == nil {
		t.Fatal("expected error for closed channel")
	}

	channel.readyState = webrtc.DataChannelStateOpen
	payload := []byte(`{"type":"custom","value":"hello"}`)
	if err := peer.SendEncodedJSON(payload); err != nil {
		t.Fatalf("SendEncodedJSON returned error: %v", err)
	}
	if len(channel.sentTexts) != 1 {
		t.Fatalf("expected 1 sent text, got %d", len(channel.sentTexts))
	}
	if channel.sentTexts[0] != string(payload) {
		t.Fatalf("expected exact payload %q, got %q", string(payload), channel.sentTexts[0])
	}
}

func TestWebRTCTransportSendEncodedLaneJSONRoutesToMatchingChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	world := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	overlay := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	sessionChannel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	event := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	peer.channels = map[string]webRTCDataChannel{
		webRTCChannelLaneWorld: world,
		"overlay":              overlay,
		"session":              sessionChannel,
		"event":                event,
	}

	if err := peer.SendEncodedLaneJSON(webRTCChannelLaneWorld, []byte("world")); err != nil {
		t.Fatalf("SendEncodedLaneJSON(world) returned error: %v", err)
	}
	if err := peer.SendEncodedLaneJSON("overlay", []byte("overlay")); err != nil {
		t.Fatalf("SendEncodedLaneJSON(overlay) returned error: %v", err)
	}
	if err := peer.SendEncodedLaneJSON("session", []byte("session")); err != nil {
		t.Fatalf("SendEncodedLaneJSON(session) returned error: %v", err)
	}
	if err := peer.SendEncodedLaneJSON("event", []byte("event")); err != nil {
		t.Fatalf("SendEncodedLaneJSON(event) returned error: %v", err)
	}
	if err := peer.SendEncodedLaneJSON("control", []byte("control")); err == nil {
		t.Fatal("expected error for control lane")
	}

	if got := world.sentTexts; len(got) != 1 || got[0] != "world" {
		t.Fatalf("expected world channel only to receive world payload, got %#v", got)
	}
	if got := overlay.sentTexts; len(got) != 1 || got[0] != "overlay" {
		t.Fatalf("expected overlay channel only to receive overlay payload, got %#v", got)
	}
	if got := sessionChannel.sentTexts; len(got) != 1 || got[0] != "session" {
		t.Fatalf("expected session channel only to receive session payload, got %#v", got)
	}
	if got := event.sentTexts; len(got) != 1 || got[0] != "event" {
		t.Fatalf("expected event channel only to receive event payload, got %#v", got)
	}
}

func TestWebRTCTransportSendToolingJSONUsesDedicatedChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	tooling := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	peer.channels = map[string]webRTCDataChannel{
		"tooling": tooling,
	}

	if err := peer.SendToolingJSON(map[string]any{"type": "tooling_packet"}); err != nil {
		t.Fatalf("SendToolingJSON returned error: %v", err)
	}
	if err := peer.SendEncodedToolingJSON([]byte(`{"type":"tooling_encoded"}`)); err != nil {
		t.Fatalf("SendEncodedToolingJSON returned error: %v", err)
	}
	if len(tooling.sentTexts) != 2 {
		t.Fatalf("expected 2 tooling payloads, got %d", len(tooling.sentTexts))
	}
	assertSentJSONField(t, tooling.sentTexts[0], "type", "tooling_packet")
	assertSentJSONField(t, tooling.sentTexts[1], "type", "tooling_encoded")

	if err := peer.SendEncodedLaneJSON("tooling", []byte("gameplay-api")); err == nil {
		t.Fatal("expected gameplay lane API to reject tooling lane")
	}
}

func TestWebRTCTransportReadyTracksAllRequiredChannels(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if _, err := peer.HandleOffer("offer", "offer-sdp"); err != nil {
		t.Fatalf("HandleOffer returned error: %v", err)
	}
	if peer.Ready() {
		t.Fatal("expected Ready to be false before channels open")
	}

	fakePeer.channels["sr.world"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.world"].onOpen()
	if peer.Ready() {
		t.Fatal("expected Ready to stay false until all channels are open")
	}

	fakePeer.channels["sr.overlay"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.overlay"].onOpen()
	fakePeer.channels["sr.session"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.session"].onOpen()
	fakePeer.channels["sr.event"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.event"].onOpen()
	fakePeer.channels["sr.asteroids"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.asteroids"].onOpen()
	fakePeer.channels["sr.bullets"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.bullets"].onOpen()
	fakePeer.channels["sr.asteroids.lifecycle"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.asteroids.lifecycle"].onOpen()
	fakePeer.channels["sr.bullets.lifecycle"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.bullets.lifecycle"].onOpen()
	fakePeer.channels["sr.tooling"].readyState = webrtc.DataChannelStateOpen
	fakePeer.channels["sr.tooling"].onOpen()
	if !peer.Ready() {
		t.Fatal("expected Ready to become true after all channels are open")
	}
}

func TestWebRTCTransportUnexpectedChannelCloseCallsHookOnceWithLane(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }

	var closedLanes []string
	peer := NewWebRTCTransport(WebRTCSignalHooks{
		OnChannelClosed: func(lane string) { closedLanes = append(closedLanes, lane) },
	})
	if _, err := peer.HandleOffer("offer", "offer-sdp"); err != nil {
		t.Fatalf("HandleOffer returned error: %v", err)
	}

	tooling := fakePeer.channels["sr.tooling"]
	tooling.readyState = webrtc.DataChannelStateOpen
	tooling.onOpen()
	tooling.onClose()
	tooling.onClose()

	if !reflect.DeepEqual(closedLanes, []string{"tooling"}) {
		t.Fatalf("expected one tooling close callback, got %#v", closedLanes)
	}
	if peer.Ready() {
		t.Fatal("expected transport to become not ready after channel close")
	}
}

func TestWebRTCTransportSendSmokeRequiresOpenChannel(t *testing.T) {
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	if err := peer.SendSmoke("smoke-1", "hello"); err == nil {
		t.Fatal("expected error without channel")
	}

	channel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	peer.channels = map[string]webRTCDataChannel{
		webRTCChannelLaneWorld: channel,
	}
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
	var packetLanes []string
	var smokePackets []map[string]any
	peer := NewWebRTCTransport(WebRTCSignalHooks{
		OnPacketReceived: func(packet map[string]any, lane string) {
			packets = append(packets, packet)
			packetLanes = append(packetLanes, lane)
		},
		OnSmokeReceived: func(packet map[string]any) { smokePackets = append(smokePackets, packet) },
	})

	if err := peer.handleChannelMessage("overlay", []byte(`{"type":"custom_packet","value":42}`)); err != nil {
		t.Fatalf("handleChannelMessage returned error: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("expected generic packet callback, got %d", len(packets))
	}
	if packetLanes[0] != "overlay" {
		t.Fatalf("expected overlay packet lane, got %q", packetLanes[0])
	}
	if len(smokePackets) != 0 {
		t.Fatalf("expected no smoke callback, got %d", len(smokePackets))
	}

	if err := peer.handleChannelMessage("tooling", []byte(`{"type":"webrtc_smoke","smoke_id":"smoke-1","message":"hello"}`)); err != nil {
		t.Fatalf("handleChannelMessage returned error: %v", err)
	}
	if len(packets) != 2 {
		t.Fatalf("expected generic packet callback for smoke packet too, got %d", len(packets))
	}
	if packetLanes[1] != "tooling" {
		t.Fatalf("expected tooling packet lane, got %q", packetLanes[1])
	}
	if len(smokePackets) != 1 {
		t.Fatalf("expected smoke packet callback, got %d", len(smokePackets))
	}

	if err := peer.handleChannelMessage("overlay", []byte("{invalid json")); err == nil {
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
	fakePeer := &fakeWebRTCPeer{channels: make(map[string]*fakeWebRTCDataChannel)}
	peer := NewWebRTCTransport(WebRTCSignalHooks{})
	peer.peer = fakePeer
	peer.channels = map[string]webRTCDataChannel{
		"sr.world":               &fakeWebRTCDataChannel{},
		"sr.overlay":             &fakeWebRTCDataChannel{},
		"sr.session":             &fakeWebRTCDataChannel{},
		"sr.event":               &fakeWebRTCDataChannel{},
		"sr.asteroids":           &fakeWebRTCDataChannel{},
		"sr.bullets":             &fakeWebRTCDataChannel{},
		"sr.asteroids.lifecycle": &fakeWebRTCDataChannel{},
		"sr.bullets.lifecycle":   &fakeWebRTCDataChannel{},
		"sr.tooling":             &fakeWebRTCDataChannel{},
	}
	if err := peer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	for lane, channel := range peer.channels {
		if !channel.(*fakeWebRTCDataChannel).closed {
			t.Fatalf("expected channel %s to be closed", lane)
		}
	}
	if !fakePeer.closed {
		t.Fatal("expected peer to be closed")
	}
}
