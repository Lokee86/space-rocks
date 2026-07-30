package main

import "testing"

func TestBuildWebRTCTransportConfigFromEnvIncludesAdvertisedPortRange(t *testing.T) {
	t.Setenv(webRTCAdvertisedIPsEnv, "147.185.221.230")
	t.Setenv(webRTCUDPPortMinEnv, "50000")
	t.Setenv(webRTCUDPPortMaxEnv, "50003")
	t.Setenv(webRTCAdvertisedUDPPortMinEnv, "21212")
	t.Setenv(webRTCAdvertisedUDPPortMaxEnv, "21215")

	config := buildWebRTCTransportConfigFromEnv()
	if len(config.AdvertisedIPs) != 1 || config.AdvertisedIPs[0] != "147.185.221.230" {
		t.Fatalf("unexpected advertised IPs: %#v", config.AdvertisedIPs)
	}
	if config.UDPPortMin != 50000 || config.UDPPortMax != 50003 {
		t.Fatalf("unexpected local UDP range: %#v", config)
	}
	if config.AdvertisedUDPPortMin != 21212 || config.AdvertisedUDPPortMax != 21215 {
		t.Fatalf("unexpected advertised UDP range: %#v", config)
	}
}
