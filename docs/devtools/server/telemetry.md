## Telemetry

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the game-server runtime telemetry and developer readout surfaces delivered through the dedicated reliable `sr.tooling` WebRTC DataChannel.

## Overview

Server telemetry is transient diagnostic visibility. It is not analytics, durable observability storage, or gameplay authority.

Current surfaces are:

```text
telemetry_subscribe / telemetry_unsubscribe
  room-scoped live runtime telemetry stream

telemetry_snapshot
  authoritative entity, process, and per-lane transport counters

telemetry_ping / telemetry_pong
  connection-scoped RTT and clock-offset diagnostic pair

debug_status_subscribe / debug_status_unsubscribe
  privileged room debug-status stream

debug_status
  room and per-player debug state

debug_shape_catalog_request / debug_shape_catalog
  privileged one-shot room collision-shape catalog

state.server_sent_msec
  authoritative timestamp on normal gameplay lane packets
```

All request-like tooling packets carry a non-empty `request_id`. Responses echo the originating request ID. None of these packet families use the normal WebSocket gameplay packet route.

## Transport and ownership

The high-level flow is:

```text
client tooling request
-> sr.tooling
-> per-connection tooling router
-> attachment and capability checks
-> telemetry provider or developer readout owner
-> correlated response or bounded server push
-> sr.tooling
```

Ownership remains split:

```text
networking/tooling router
  packet policy, request correlation, subscription lifecycle, cadence, and errors

tooling Controller
  live telemetry provider and measurement integration

devtools Controller
  debug-status and shape-catalog projection over authoritative game seams

game and rooms
  authoritative simulation, room, entity, and match facts

client
  overlay visibility, labels, formatting, and presentation
```

The transport layer does not duplicate simulation or mutation logic.

## Authorization

Public-safe telemetry packets require an eligible connected session and packet-specific room attachment, but not `tooling.read` or `tooling.control`:

```text
telemetry_ping
telemetry_subscribe
telemetry_unsubscribe
```

Privileged readouts require `tooling.read`:

```text
debug_status_subscribe
debug_status_unsubscribe
debug_shape_catalog_request
```

The temporary capability grant currently gives every connected session `tooling.read` and `tooling.control`. The packet router still enforces policy so the later account-backed grant source can replace the temporary source without changing packet handlers.

Readout authorization is attached-room based, not participant-slot based. An authorized observer can receive room-global status or catalog data without a `GamePlayerID`.

## Live telemetry snapshots

`services/game-server/internal/tooling/Controller` implements the production telemetry-provider seam independently from bounded measurement runs.

A `telemetry_snapshot` includes:

```text
server_room_count
server_match_id
server_players
server_player_sessions
server_enemies
server_asteroids
server_projectiles
server_pickups
server_radial_effects
server_total_asteroids_spawned
server_heap_allocated_bytes
server_heap_in_use_bytes
server_system_bytes
server_goroutines
server_gc_cycles
server_packets_out
server_bytes_out
server_max_packet_bytes
server_lane_metrics
```

`server_lane_metrics` is keyed by physical realtime lane. Each entry contains:

```text
packet_count
encoded_bytes_total
last_encoded_bytes
maximum_encoded_bytes
last_packet_family
```

Entity counts come from the authoritative game aggregate through the same concrete count seam used by runtime measurement. Process counters use the shared bounded process sampler. Packet counters are observed after successful encoded WebRTC lane writes.

Live counters are connection-local and room-scoped. Moving the same connection to another room resets its live telemetry sequence and packet totals rather than leaking the earlier room's traffic into the new room.

## Subscription lifecycle

Telemetry and debug-status subscriptions are independent.

```text
telemetry_subscribe
-> router records the current room attachment
-> first eligible tooling tick emits telemetry_snapshot
-> later snapshots emit at bounded cadence

telemetry_unsubscribe
-> router stops telemetry snapshots

room attachment changes
-> room-scoped subscription is invalidated

tooling channel closes
-> all subscriptions and live telemetry state are cleared

successful WebRTC recovery
-> client explicitly resubscribes after sr.tooling becomes ready
```

UI visibility may cause a client to subscribe or unsubscribe, but visibility is not server authority.

## Telemetry ping and pong

`telemetry_ping` is a connection-scoped tooling request with:

```text
type = telemetry_ping
request_id
sequence
client_sent_msec
```

The router responds on `sr.tooling` with:

```text
type = telemetry_pong
request_id
sequence
client_sent_msec
server_received_msec
server_sent_msec
```

The response preserves request correlation and client timing fields. It adds the server receive and send wall-clock timestamps used by the client to calculate RTT and estimate the offset between server Unix time and client monotonic time.

Ping/pong does not require room membership and does not mutate gameplay state.

## Debug status

`debug_status_subscribe` starts a connection-local bounded-cadence status stream for the attached room. `debug_status_unsubscribe` stops it.

`debug_status` may contain room-global flags and authorized player-status maps, including:

```text
invincible
infinite_lives
world_frozen
asteroids_frozen
bullets_frozen
spawning_frozen
collisions_frozen
player_frozen
```

The tooling router owns subscription and delivery. The devtools controller reads facts from authoritative game control seams. The readout does not grant mutation authority and does not require the receiving session to occupy a gameplay participant slot.

## Debug shape catalog

`debug_shape_catalog_request` is a one-shot correlated request for the current room's imported collision-shape definitions.

The response contains shape definitions keyed by stable IDs. Each definition contains:

```text
id
kind
shape_type
points
```

Current kinds include player, asteroid, bullet, and pickup. The catalog is static metadata for the room/runtime contract and is not pushed every write tick.

Live collision-body instances remain a separate authoritative observation seam; they are not hidden inside the catalog or normal gameplay packets.

## Gameplay timestamps

Normal gameplay lane packets carry `server_sent_msec` from the immutable authoritative presentation frame. All packets projected from the same frame share that timestamp.

The client combines `server_sent_msec` with the ping/pong-derived server clock offset to calculate packet age. This frame timestamp is separate from `telemetry_pong.server_sent_msec`, which is captured immediately before the pong is sent.

Compact runtime metadata inference may omit redundant envelope fields on the wire. It does not remove `server_sent_msec` from the gameplay timing contract.

## Build and runtime gates

The `nodevtools` build tag disables privileged developer command and readout implementations. Public-safe telemetry and runtime measurement remain networking/tooling capabilities and continue to compile and operate through their own seams.

Runtime requirements are:

```text
telemetry_ping
  connected tooling session

telemetry_subscribe
  attached room with an active game runtime

debug_status_subscribe
  attached room, tooling.read, and available devtools status owner

debug_shape_catalog_request
  attached room, tooling.read, and available catalog owner
```

Rejected requests return `tooling_error`; they do not silently mutate or fall back to WebSocket.

## Code map

```text
shared/packets/tooling.toml
shared/packets/debug.toml
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/preflight.go
services/game-server/internal/tooling/controller.go
services/game-server/internal/tooling/telemetry.go
services/game-server/internal/game/runtime_measurement.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/webrtc_transport.go
services/game-server/internal/protocol/tooling/
```

The historical connection-loop file name `websocket_write.go` still owns the session write loop. Diagnostic payload delivery itself uses `SendToolingJSON` on `sr.tooling`.

## Tests and verification

Focused coverage includes:

```text
services/game-server/internal/networking/tooling/router_test.go
services/game-server/internal/tooling/controller_test.go
services/game-server/internal/networking/websocket_measurement_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
client/tests/unit/networking/test_webrtc_transport.gd
client/tests/unit/test_client_connection_service.gd
```

Tests cover request correlation, room-scoped subscription cleanup, production snapshot fields, per-lane packet accounting, room-change counter reset, recovery resubscription, ping/pong delivery, and privileged readout routing.

## Related docs

* [Tooling Channel Migration Contract](../design/tooling-channel-migration-contract.md)
* [Client Telemetry Overlays](../client/telemetry-overlays.md)
* [Game Server Telemetry And Packet Routing](../../services/game-server/networking/telemetry-packet-routing.md)
* [Realtime WebRTC Gameplay Transport](../../protocol/realtime-webrtc-gameplay-transport.md)
* [Runtime Measurement](../../services/game-server/observability/!INDEX.md)

## Notes

Telemetry in this document means live runtime diagnostics. Durable measurements, structured logs, and trace events remain separate systems with separate ownership.
