package networking

import (
	"strings"
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestAdvertisedWebRTCCandidateMapsSequentialUDPPortRange(t *testing.T) {
	candidate := &webrtc.ICECandidate{
		Foundation: "1",
		Priority:   2130706431,
		Address:    "147.185.221.230",
		Protocol:   webrtc.ICEProtocolUDP,
		Port:       50002,
		Typ:        webrtc.ICECandidateTypeHost,
		Component:  1,
	}
	config := WebRTCTransportConfig{
		UDPPortMin:           50000,
		UDPPortMax:           50003,
		AdvertisedUDPPortMin: 21212,
		AdvertisedUDPPortMax: 21215,
	}

	advertised := advertisedWebRTCCandidate(candidate, config)
	if advertised.Port != 21214 {
		t.Fatalf("advertised port = %d, want 21214", advertised.Port)
	}
	if candidate.Port != 50002 {
		t.Fatalf("source candidate was mutated: %d", candidate.Port)
	}
	if got := advertised.ToJSON().Candidate; !strings.Contains(got, " 21214 typ host") {
		t.Fatalf("candidate did not contain translated port: %q", got)
	}
}

func TestWebRTCTransportPublishesTranslatedCandidate(t *testing.T) {
	oldFactory := newWebRTCPeerConnection
	defer func() { newWebRTCPeerConnection = oldFactory }()

	fakePeer := &fakeWebRTCPeer{
		answer:   webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: "answer-sdp"},
		channels: make(map[string]*fakeWebRTCDataChannel),
	}
	newWebRTCPeerConnection = func() (webRTCPeer, error) { return fakePeer, nil }
	restore := SetWebRTCTransportConfigForTests(WebRTCTransportConfig{
		UDPPortMin:           50000,
		UDPPortMax:           50003,
		AdvertisedUDPPortMin: 21212,
		AdvertisedUDPPortMax: 21215,
	})
	defer restore()

	var published string
	transport := NewWebRTCTransport(WebRTCSignalHooks{
		OnLocalICECandidate: func(_ string, _ int, candidate string) {
			published = candidate
		},
	})
	if _, err := transport.HandleOffer("offer", "offer-sdp"); err != nil {
		t.Fatalf("HandleOffer returned error: %v", err)
	}
	fakePeer.onICE(&webrtc.ICECandidate{
		Foundation: "1",
		Priority:   2130706431,
		Address:    "147.185.221.230",
		Protocol:   webrtc.ICEProtocolUDP,
		Port:       50001,
		Typ:        webrtc.ICECandidateTypeHost,
		Component:  1,
	})

	if !strings.Contains(published, " 21213 typ host") {
		t.Fatalf("published candidate did not contain translated port: %q", published)
	}
}

func TestAdvertisedWebRTCCandidateLeavesInvalidMappingsUnchanged(t *testing.T) {
	candidate := &webrtc.ICECandidate{
		Protocol: webrtc.ICEProtocolUDP,
		Port:     50001,
		Typ:      webrtc.ICECandidateTypeHost,
	}
	config := WebRTCTransportConfig{
		UDPPortMin:           50000,
		UDPPortMax:           50003,
		AdvertisedUDPPortMin: 21212,
		AdvertisedUDPPortMax: 21214,
	}

	advertised := advertisedWebRTCCandidate(candidate, config)
	if advertised.Port != candidate.Port {
		t.Fatalf("invalid unequal ranges translated port to %d", advertised.Port)
	}
}
