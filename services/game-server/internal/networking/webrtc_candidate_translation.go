package networking

import "github.com/pion/webrtc/v4"

func advertisedWebRTCCandidate(candidate *webrtc.ICECandidate, config WebRTCTransportConfig) webrtc.ICECandidate {
	advertised := *candidate
	if candidate.Protocol != webrtc.ICEProtocolUDP || candidate.Typ != webrtc.ICECandidateTypeHost {
		return advertised
	}
	if config.UDPPortMin == 0 || config.UDPPortMax == 0 || config.AdvertisedUDPPortMin == 0 || config.AdvertisedUDPPortMax == 0 {
		return advertised
	}
	if config.UDPPortMin > config.UDPPortMax || config.AdvertisedUDPPortMin > config.AdvertisedUDPPortMax {
		return advertised
	}
	if config.UDPPortMax-config.UDPPortMin != config.AdvertisedUDPPortMax-config.AdvertisedUDPPortMin {
		return advertised
	}
	if candidate.Port < config.UDPPortMin || candidate.Port > config.UDPPortMax {
		return advertised
	}
	advertised.Port = config.AdvertisedUDPPortMin + (candidate.Port - config.UDPPortMin)
	return advertised
}
