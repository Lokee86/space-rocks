# Telemetry And Packet Routing

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes the game-server boundary for live telemetry packets carried by the dedicated reliable `sr.tooling` WebRTC DataChannel.

## Overview

Live telemetry is runtime diagnostic traffic, not durable observability storage and not gameplay authority.

The current route is:

```text
client telemetry UI
-> telemetry_subscribe or telemetry_ping
-> sr.tooling
-> per-connection tooling router
-> production telemetry provider or ping handler
-> telemetry_snapshot or telemetry_pong
-> sr.tooling
-> client tooling packet router
```

WebSocket remains responsible for authentication, WebRTC signaling, lobby/room control, and pre-readiness session control. It no longer carries telemetry subscriptions, snapshots, or ping/pong packets.

## Responsibilities

The tooling telemetry boundary owns:

* Validating generated telemetry packet contracts and non-empty request IDs.
* Enforcing required connection or room attachment before decoding payloads.
* Keeping subscriptions connection-local and bound to the attached room.
* Sending bounded-cadence `telemetry_snapshot` packets only while subscribed.
* Sending correlated `telemetry_pong` responses for `telemetry_ping` requests.
* Clearing telemetry state on unsubscribe, room change, tooling-channel close, or session close.
* Reading authoritative runtime facts through existing room, game, measurement, and networking seams.

It does not own gameplay mutation, room membership policy, client presentation, durable metrics storage, or external analytics.

## Packet surface

Client to server:

```text
telemetry_subscribe
telemetry_unsubscribe
telemetry_ping
```

Server to client:

```text
telemetry_snapshot
telemetry_pong
tooling_error
```

Every client-originated telemetry packet carries a non-empty `request_id`.

`telemetry_ping` is connection-scoped and does not require room membership. `telemetry_pong` echoes `request_id`, sequence, and client send time, then adds server receive and send timestamps.

`telemetry_subscribe` and `telemetry_unsubscribe` are room-scoped. The subscription is valid only while the connection remains attached to the same room.

## Production telemetry provider

`services/game-server/internal/tooling/Controller` implements both measurement and live telemetry provider contracts.

A live snapshot includes:

```text
authoritative entity counts
active room and match identity
process heap/system/goroutine/GC counters
aggregate server packet and byte totals for the connection
per-lane packet count, byte total, last/max packet size, and last packet family
```

The provider reuses the authoritative runtime entity-count seam used by measurement. Packet write counters are observed at the successful WebRTC lane-write boundary, so they reflect encoded packets actually handed to the transport.

Live telemetry state is independent from a bounded measurement run. A connection can receive telemetry snapshots without starting a measurement session.

## Subscription lifecycle

```text
visible eligible client panel
-> telemetry_subscribe
-> router stores current room id
-> first eligible tooling tick emits telemetry_snapshot
-> later snapshots emit at bounded cadence

panel hidden
-> telemetry_unsubscribe
-> router stops snapshots

room attachment changes
-> router invalidates subscription

WebRTC tooling recovery
-> old router closes and clears state
-> client explicitly resubscribes after readiness

session close
-> router clears telemetry state and finalizes any active measurement run
```

UI visibility decides whether the client requests a subscription. It is not server authority; the server still validates attachment and packet policy.

## Trust and validation

The server trusts generated packet structure and server-owned runtime state. It rejects malformed packets, missing request IDs, missing room attachment for subscription packets, client-submitted server packet types, and unavailable runtime ownership with `tooling_error`.

Telemetry packets never accept client claims about authoritative entity counts or server process state.

## Code map

```text
shared/packets/tooling.toml
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/preflight.go
services/game-server/internal/tooling/controller.go
services/game-server/internal/tooling/telemetry.go
services/game-server/internal/game/runtime_measurement.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/webrtc_transport.go
services/game-server/internal/protocol/tooling/
```

The historical file names `websocket_write.go` and `websocket_session.go` still own the connection/session write loop, but telemetry payload delivery itself uses `SendToolingJSON` on `sr.tooling`.

## Tests and verification

Relevant coverage includes:

```text
services/game-server/internal/networking/tooling/router_test.go
services/game-server/internal/tooling/controller_test.go
services/game-server/internal/networking/websocket_measurement_test.go
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
client/tests/unit/networking/test_webrtc_transport.gd
client/tests/unit/test_client_connection_service.gd
```

Tests cover request correlation, connection-local subscription ownership, room-change cleanup, authoritative snapshot content, independent live packet counters, client recovery resubscription, and WebSocket/WebRTC metric aggregation.

## Related docs

* [Tooling Channel Migration Contract](../../../devtools/design/tooling-channel-migration-contract.md)
* [Telemetry Overlays](../../../devtools/client/telemetry-overlays.md)
* [Game Server Networking](./!INDEX.md)
* [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)

## Notes

Live telemetry remains transient and shallow. Durable measurements are produced by the separate runtime measurement lifecycle; logs and traces remain owned by the observability system.
