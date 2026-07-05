package networking

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
)

const webRTCChannelLabel = "sr.reliable"
const webRTCChannelID = 1
const webRTCSmokeOriginServer = "server"

type WebRTCAnswerPayload struct {
	DescriptionType string
	SDP             string
}

type WebRTCSignalHooks struct {
	OnLocalICECandidate func(media string, index int, name string)
	OnReady             func()
	OnPacketReceived    func(packet map[string]any)
	OnSmokeReceived     func(packet map[string]any)
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
	api := webrtc.NewAPI()
	peer, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}
	return &pionPeerAdapter{peer: peer}, nil
}

type WebRTCTransport struct {
	peer    webRTCPeer
	channel webRTCDataChannel
	hooks   WebRTCSignalHooks
	ready   bool
}

func NewWebRTCTransport(hooks WebRTCSignalHooks) *WebRTCTransport {
	return &WebRTCTransport{hooks: hooks}
}

func (p *WebRTCTransport) Ready() bool {
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
		if err := p.createNegotiatedChannel(); err != nil {
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

func (p *WebRTCTransport) SendJSON(packet map[string]any) error {
	if p.channel == nil {
		return errors.New("webrtc data channel is not ready")
	}
	if p.channel.ReadyState() != webrtc.DataChannelStateOpen {
		return errors.New("webrtc data channel is not open")
	}
	payload, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	return p.channel.SendText(string(payload))
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
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.peer != nil {
		return p.peer.Close()
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

func (p *WebRTCTransport) createNegotiatedChannel() error {
	ordered := true
	negotiated := true
	channelID := uint16(webRTCChannelID)
	channel, err := p.peer.CreateDataChannel(webRTCChannelLabel, &webrtc.DataChannelInit{
		Ordered:    &ordered,
		Negotiated: &negotiated,
		ID:         &channelID,
	})
	if err != nil {
		return err
	}
	p.channel = channel
	p.channel.OnOpen(func() {
		p.ready = true
		if p.hooks.OnReady != nil {
			p.hooks.OnReady()
		}
	})
	p.channel.OnMessage(func(msg webrtc.DataChannelMessage) {
		_ = p.handleChannelMessage(msg.Data)
	})
	return nil
}

func (p *WebRTCTransport) handleChannelMessage(data []byte) error {
	var packet map[string]any
	if err := json.Unmarshal(data, &packet); err != nil {
		return fmt.Errorf("invalid smoke packet json: %w", err)
	}
	if p.hooks.OnPacketReceived != nil {
		p.hooks.OnPacketReceived(packet)
	}
	if fmt.Sprint(packet["type"]) == "webrtc_smoke" && p.hooks.OnSmokeReceived != nil {
		p.hooks.OnSmokeReceived(packet)
	}
	return nil
}

