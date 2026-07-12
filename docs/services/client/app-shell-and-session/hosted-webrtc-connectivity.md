# Hosted WebRTC Connectivity: ICE, STUN, And TURN

Parent index: [App Shell And Session](./!INDEX.md)

## Purpose

This document records the hosted-connectivity policy for the Godot client's WebRTC DataChannel connection to the dedicated game server.

It exists to prevent future work from treating an empty external ICE-server list as proof that hosted multiplayer requires TURN, or as proof that ICE is disabled.

> [!IMPORTANT]
> ICE is always part of WebRTC connection establishment. An empty `WEBRTC_ICE_SERVERS` value means that no external STUN or TURN service is configured; it does not disable ICE.

## Current topology

Space Rocks uses a client-to-dedicated-server topology rather than peer-to-peer player hosting.

The normal hosted path should be:

```text
player client
-> direct WebRTC UDP connection
-> publicly reachable game server
```

The hosted game server must advertise a reachable public ICE candidate and expose the required UDP path through its host firewall, network firewall, and deployment platform. The WebSocket signaling URL is separate from this UDP path. A Cloudflare-proxied HTTP or WebSocket route must not be assumed to relay WebRTC UDP traffic.

A correctly configured public dedicated server can accept direct connections from many normal home networks without TURN and may not require an external STUN service for the server candidate itself.

## STUN and TURN roles

STUN helps a WebRTC endpoint discover the public address and port assigned by NAT. It is candidate-discovery infrastructure; it does not normally carry gameplay traffic.

TURN relays gameplay traffic when no usable direct candidate pair can be established:

```text
player client
-> TURN relay
-> game server
```

TURN may be needed for players behind restrictive NAT, networks that block UDP, institutional or corporate firewalls, proxies, or other environments where direct connectivity fails. TURN adds relay bandwidth cost, operational ownership, credentials and abuse protection, latency, and another production dependency.

## Deployment policy

Do not make TURN the default path or add it solely because players are on separate networks.

Use this sequence for hosted deployment:

1. Deploy the game server with a correct public advertised address and open UDP path.
2. Verify direct WebRTC connectivity from multiple external residential networks.
3. Test mobile hotspots, VPNs, institutional Wi-Fi, restrictive firewalls, and networks that block UDP.
4. Record connection-success and failure reasons through the existing signaling and transport diagnostics.
5. Add STUN or TURN only when hosted test evidence shows that the direct candidate path is insufficient.
6. Prefer TURN as a fallback candidate rather than the normal gameplay route.

## Revisit triggers

Revisit external STUN or TURN deployment when any of the following occurs:

- hosted external clients cannot establish a direct candidate pair despite correct public server advertisement and firewall configuration;
- connection failures cluster on restrictive NAT, UDP-blocked, institutional, corporate, VPN, or mobile networks;
- the product requires an explicit supported-network compatibility target that direct UDP cannot meet;
- signaling diagnostics show a meaningful connection-failure rate that a relay could address.

Until those triggers are observed, the empty external ICE-server configuration is an untested hosted-connectivity concern, not a confirmed multiplayer limitation.

## Ownership

- `shared/constants/client/shell.toml` owns the client external ICE-server configuration seam.
- `client/scripts/networking/webrtc/` owns client WebRTC transport behavior.
- Server WebRTC configuration owns the advertised server candidate and UDP bind/public-address behavior.
- Deployment infrastructure owns public address routing and firewall exposure.
- Hosted connectivity tests determine whether STUN or TURN becomes a production requirement.

## Related docs

- [Session Boot And Network Target](session-boot-and-network-target.md)
- [Networking Flow](../networking-flow/!INDEX.md)
- [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)
- [Current System Limits](../../../limits/current-system-limits.md)
