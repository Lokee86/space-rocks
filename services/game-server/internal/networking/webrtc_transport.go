package networking

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
)

type webRTCChannelSpec struct {
	Lane           string
	Label          string
	ID             uint16
	Ordered        bool
	MaxRetransmits *uint16
}

const webRTCChannelLaneWorld = "world"

func webRTCChannelSpecs() []webRTCChannelSpec {
	zeroRetransmits := uint16(0)
	return []webRTCChannelSpec{
		{Lane: "world", Label: "sr.world", ID: 1, Ordered: true},
		{Lane: "overlay", Label: "sr.overlay", ID: 2, Ordered: true},
		{Lane: "session", Label: "sr.session", ID: 3, Ordered: true},
		{Lane: "event", Label: "sr.event", ID: 4, Ordered: true},
		{Lane: "asteroids", Label: "sr.asteroids", ID: 5, Ordered: false, MaxRetransmits: &zeroRetransmits},
		{Lane: "bullets", Label: "sr.bullets", ID: 6, Ordered: false, MaxRetransmits: &zeroRetransmits},
		{Lane: "asteroids.lifecycle", Label: "sr.asteroids.lifecycle", ID: 7, Ordered: true},
		{Lane: "bullets.lifecycle", Label: "sr.bullets.lifecycle", ID: 8, Ordered: true},
		{Lane: "tooling", Label: "sr.tooling", ID: 9, Ordered: true},
		{Lane: "ships", Label: "sr.ships", ID: 10, Ordered: false, MaxRetransmits: &zeroRetransmits},
		{Lane: "ships.lifecycle", Label: "sr.ships.lifecycle", ID: 11, Ordered: true},
	}
}

func webRTCChannelLabelForLane(lane string) (string, bool) {
	for _, spec := range webRTCChannelSpecs() {
		if spec.Lane == lane {
			return spec.Label, true
		}
	}
	return "", false
}

const webRTCSmokeOriginServer = "server"

type WebRTCAnswerPayload struct {
	DescriptionType string
	SDP             string
}

type WebRTCSignalHooks struct {
	OnLocalICECandidate func(media string, index int, name string)
	OnReady             func()
	OnPacketReceived    func(packet map[string]any, lane string)
	OnSmokeReceived     func(packet map[string]any)
	OnChannelClosed     func(lane string)
}

type WebRTCTransportConfig struct {
	AdvertisedIPs []string
	UDPPortMin    uint16
	UDPPortMax    uint16
}

var webRTCTransportConfig WebRTCTransportConfig

var newWebRTCPeerConnectionAPI = func(config WebRTCTransportConfig) (*webrtc.API, error) {
	settingEngine := webrtc.SettingEngine{}
	if len(config.AdvertisedIPs) > 0 {
		settingEngine.SetNAT1To1IPs(config.AdvertisedIPs, webrtc.ICECandidateTypeHost)
	}
	if config.UDPPortMin != 0 && config.UDPPortMax != 0 {
		settingEngine.SetEphemeralUDPPortRange(uint16(config.UDPPortMin), uint16(config.UDPPortMax))
	}
	return webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine)), nil
}

func SetWebRTCTransportConfig(config WebRTCTransportConfig) {
	webRTCTransportConfig = config
}

func SetWebRTCTransportConfigForTests(config WebRTCTransportConfig) func() {
	previous := webRTCTransportConfig
	SetWebRTCTransportConfig(config)
	return func() {
		SetWebRTCTransportConfig(previous)
	}
}

type webRTCPeer interface {
	SetRemoteDescription(desc webrtc.SessionDescription) error
	CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error)
	SetLocalDescription(desc webrtc.SessionDescription) error
	AddICECandidate(candidate webrtc.ICECandidateInit) error
	OnICECandidate(f func(*webrtc.ICECandidate))
	CreateDataChannel(label string, init *webrtc.DataChannelInit) (webRTCDataChannel, error)
	Close() error
}

type webRTCDataChannel interface {
	OnOpen(f func())
	OnClose(f func())
	OnMessage(f func(webrtc.DataChannelMessage))
	ReadyState() webrtc.DataChannelState
	SendText(s string) error
	Close() error
}

type pionPeerAdapter struct {
	peer *webrtc.PeerConnection
}

func (adapter *pionPeerAdapter) SetRemoteDescription(desc webrtc.SessionDescription) error {
	return adapter.peer.SetRemoteDescription(desc)
}

func (adapter *pionPeerAdapter) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) {
	return adapter.peer.CreateAnswer(options)
}

func (adapter *pionPeerAdapter) SetLocalDescription(desc webrtc.SessionDescription) error {
	return adapter.peer.SetLocalDescription(desc)
}

func (adapter *pionPeerAdapter) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return adapter.peer.AddICECandidate(candidate)
}

func (adapter *pionPeerAdapter) OnICECandidate(f func(*webrtc.ICECandidate)) {
	adapter.peer.OnICECandidate(f)
}

func (adapter *pionPeerAdapter) CreateDataChannel(label string, init *webrtc.DataChannelInit) (webRTCDataChannel, error) {
	channel, err := adapter.peer.CreateDataChannel(label, init)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (adapter *pionPeerAdapter) Close() error {
	return adapter.peer.Close()
}

var newWebRTCPeerConnection = func() (webRTCPeer, error) {
	api, err := newWebRTCPeerConnectionAPI(webRTCTransportConfig)
	if err != nil {
		return nil, err
	}
	peer, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}
	return &pionPeerAdapter{peer: peer}, nil
}

type WebRTCTransport struct {
	peer                 webRTCPeer
	mu                   sync.RWMutex
	channels             map[string]webRTCDataChannel
	readyChannels        map[string]bool
	channelCloseReported map[string]bool
	hooks                WebRTCSignalHooks
	ready                bool
	closing              bool
}

func NewWebRTCTransport(hooks WebRTCSignalHooks) *WebRTCTransport {
	return &WebRTCTransport{
		hooks:                hooks,
		channels:             make(map[string]webRTCDataChannel),
		readyChannels:        make(map[string]bool),
		channelCloseReported: make(map[string]bool),
	}
}

func (p *WebRTCTransport) Ready() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ready
}

func (p *WebRTCTransport) HandleOffer(descriptionType string, sdp string) (*WebRTCAnswerPayload, error) {
	if descriptionType == "" || sdp == "" {
		return nil, errors.New("offer description_type and sdp are required")
	}

	if p.peer == nil {
		peer, err := newWebRTCPeerConnection()
		if err != nil {
			return nil, err
		}
		p.peer = peer
		p.attachPeerHandlers()
		if err := p.createNegotiatedChannels(); err != nil {
			return nil, err
		}
	}

	if err := p.peer.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}); err != nil {
		return nil, err
	}
	answer, err := p.peer.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	if err := p.peer.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	return &WebRTCAnswerPayload{DescriptionType: answer.Type.String(), SDP: answer.SDP}, nil
}

func (p *WebRTCTransport) AddRemoteCandidate(media string, index int, name string) error {
	if p.peer == nil {
		return errors.New("webrtc peer is not initialized")
	}
	idx := uint16(index)
	return p.peer.AddICECandidate(webrtc.ICECandidateInit{SDPMid: &media, SDPMLineIndex: &idx, Candidate: name})
}

func (p *WebRTCTransport) SendEncodedJSON(encoded []byte) error {
	return p.SendEncodedLaneJSON(webRTCChannelLaneWorld, encoded)
}

func (p *WebRTCTransport) SendJSON(packet map[string]any) error {
	return p.SendLaneJSON(webRTCChannelLaneWorld, packet)
}

func (p *WebRTCTransport) SendEncodedLaneJSON(lane string, encoded []byte) error {
	channel, err := p.gameplayChannelForLane(lane)
	if err != nil {
		return err
	}
	return channel.SendText(string(encoded))
}

func (p *WebRTCTransport) SendLaneJSON(lane string, packet map[string]any) error {
	payload, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	return p.SendEncodedLaneJSON(lane, payload)
}

func (p *WebRTCTransport) SendEncodedToolingJSON(encoded []byte) error {
	channel, err := p.toolingChannel()
	if err != nil {
		return err
	}
	return channel.SendText(string(encoded))
}

func (p *WebRTCTransport) SendToolingJSON(packet map[string]any) error {
	payload, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	return p.SendEncodedToolingJSON(payload)
}

func (p *WebRTCTransport) SendSmoke(smokeID string, message string) error {
	return p.SendJSON(map[string]any{
		"type":     "webrtc_smoke",
		"smoke_id": smokeID,
		"origin":   webRTCSmokeOriginServer,
		"message":  message,
	})
}

func (p *WebRTCTransport) Close() error {
	p.mu.Lock()
	p.closing = true
	channels := make([]webRTCDataChannel, 0)
	for lane, channel := range p.channels {
		if channel != nil {
			channels = append(channels, channel)
		}
		delete(p.channels, lane)
	}
	for lane := range p.readyChannels {
		delete(p.readyChannels, lane)
	}
	p.ready = false
	peer := p.peer
	p.mu.Unlock()

	for _, channel := range channels {
		_ = channel.Close()
	}
	if peer != nil {
		return peer.Close()
	}
	return nil
}

func (p *WebRTCTransport) attachPeerHandlers() {
	p.peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil || p.hooks.OnLocalICECandidate == nil {
			return
		}
		init := candidate.ToJSON()
		index := 0
		if init.SDPMLineIndex != nil {
			index = int(*init.SDPMLineIndex)
		}
		media := ""
		if init.SDPMid != nil {
			media = *init.SDPMid
		}
		p.hooks.OnLocalICECandidate(media, index, init.Candidate)
	})
}

func (p *WebRTCTransport) createNegotiatedChannels() error {
	negotiated := true
	for _, spec := range webRTCChannelSpecs() {
		ordered := spec.Ordered
		channelID := spec.ID
		channel, err := p.peer.CreateDataChannel(spec.Label, &webrtc.DataChannelInit{
			Ordered:        &ordered,
			Negotiated:     &negotiated,
			ID:             &channelID,
			MaxRetransmits: spec.MaxRetransmits,
		})
		if err != nil {
			return err
		}
		lane := spec.Lane
		p.mu.Lock()
		p.channels[lane] = channel
		p.readyChannels[lane] = false
		p.mu.Unlock()
		channel.OnOpen(func() {
			p.mu.Lock()
			p.readyChannels[lane] = true
			wasReady := p.ready
			p.ready = p.allRequiredChannelsReadyLocked()
			readyChanged := !wasReady && p.ready
			onReady := p.hooks.OnReady
			p.mu.Unlock()
			if readyChanged && onReady != nil {
				onReady()
			}
		})
		channel.OnClose(func() {
			p.handleChannelClosed(lane)
		})
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			_ = p.handleChannelMessage(lane, msg.Data)
		})
	}
	return nil
}

func (p *WebRTCTransport) gameplayChannelForLane(lane string) (webRTCDataChannel, error) {
	if lane == "" {
		return nil, errors.New("webrtc gameplay lane is required")
	}
	if lane == "control" {
		return nil, errors.New("webrtc control lane is websocket-owned")
	}
	if lane == "tooling" {
		return nil, errors.New("webrtc tooling lane uses dedicated transport API")
	}
	if _, ok := webRTCChannelLabelForLane(lane); !ok {
		return nil, errors.New("webrtc gameplay lane channel is not configured")
	}
	p.mu.RLock()
	channel, ok := p.channels[lane]
	p.mu.RUnlock()
	if !ok || channel == nil {
		return nil, errors.New("webrtc gameplay lane channel is not configured")
	}
	if channel.ReadyState() != webrtc.DataChannelStateOpen {
		return nil, errors.New("webrtc gameplay lane channel is not open")
	}
	return channel, nil
}

func (p *WebRTCTransport) toolingChannel() (webRTCDataChannel, error) {
	p.mu.RLock()
	channel, ok := p.channels["tooling"]
	p.mu.RUnlock()
	if !ok || channel == nil {
		return nil, errors.New("webrtc tooling lane channel is not configured")
	}
	if channel.ReadyState() != webrtc.DataChannelStateOpen {
		return nil, errors.New("webrtc tooling lane channel is not open")
	}
	return channel, nil
}

func (p *WebRTCTransport) allRequiredChannelsReadyLocked() bool {
	for _, spec := range webRTCChannelSpecs() {
		if !p.readyChannels[spec.Lane] {
			return false
		}
	}
	return true
}

func (p *WebRTCTransport) handleChannelClosed(lane string) {
	p.mu.Lock()
	if p.channelCloseReported == nil {
		p.channelCloseReported = make(map[string]bool)
	}
	wasReady := p.readyChannels[lane]
	if !wasReady || p.channelCloseReported[lane] || p.closing {
		p.readyChannels[lane] = false
		p.ready = false
		p.mu.Unlock()
		return
	}
	p.readyChannels[lane] = false
	p.ready = false
	p.channelCloseReported[lane] = true
	onChannelClosed := p.hooks.OnChannelClosed
	p.mu.Unlock()
	if onChannelClosed != nil {
		onChannelClosed(lane)
	}
}

func (p *WebRTCTransport) handleChannelMessage(lane string, data []byte) error {
	var packet map[string]any
	if err := json.Unmarshal(data, &packet); err != nil {
		return fmt.Errorf("invalid smoke packet json: %w", err)
	}
	if p.hooks.OnPacketReceived != nil {
		p.hooks.OnPacketReceived(packet, lane)
	}
	if fmt.Sprint(packet["type"]) == "webrtc_smoke" && p.hooks.OnSmokeReceived != nil {
		p.hooks.OnSmokeReceived(packet)
	}
	return nil
}
