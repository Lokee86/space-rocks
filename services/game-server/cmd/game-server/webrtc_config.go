package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
)

const (
	webRTCAdvertisedIPsEnv = "SPACE_ROCKS_WEBRTC_ADVERTISED_IPS"
	webRTCUDPPortMinEnv    = "SPACE_ROCKS_WEBRTC_UDP_PORT_MIN"
	webRTCUDPPortMaxEnv    = "SPACE_ROCKS_WEBRTC_UDP_PORT_MAX"
)

func buildWebRTCTransportConfigFromEnv() networking.WebRTCTransportConfig {
	config := networking.WebRTCTransportConfig{}
	config.AdvertisedIPs = parseCommaSeparatedEnvList(os.Getenv(webRTCAdvertisedIPsEnv))
	config.UDPPortMin = parseOptionalUint16Env(webRTCUDPPortMinEnv)
	config.UDPPortMax = parseOptionalUint16Env(webRTCUDPPortMaxEnv)
	return config
}

func parseCommaSeparatedEnvList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func parseOptionalUint16Env(name string) uint16 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0
	}
	return uint16(parsed)
}
