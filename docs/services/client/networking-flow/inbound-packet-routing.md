# Inbound Packet Routing

Parent index: [Networking Flow](./!README.md)

## Purpose

This document describes the current client inbound packet routing path.

It covers how decoded server packet dictionaries move from the client WebSocket transport into packet classification, dispatcher signals, connection-service signals, and downstream session, room, gameplay, auth, devtools, and telemetry consumers.

## Overview

Inbound packet routing begins after the WebSocket transport has already decoded raw text into a packet dictionary.

`NetworkClient` owns raw WebSocket polling, text receive, JSON decode, envelope validation, and `packet_received` emission. After that signal fires, inbound routing is owned by `ClientConnectionService`, `ServerPacketDispatcher`, and `ServerPacketRouter`.

Current flow:

```text
NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter packet-type checks
-> ServerPacketDispatcher emits a typed signal
-> ClientConnectionService re-emits the typed service signal
-> SessionNetworkController or direct consumer handles the packet
```

The routing path is signal-based. It does not mutate server authority, does not parse payload-specific gameplay data, and does not apply presentation state directly. Its job is to classify packet family by generated packet type constants and forward the dictionary to the owning client subsystem.

## Code root

* `client/scripts/networking/`
* `client/scripts/session/`

## Responsibilities

* Receive decoded packet dictionaries from `NetworkClient`.
* Dispatch inbound packets by generated packet type constants.
* Re-emit typed packet signals from the connection service.
* Track websocket auth result state from `authenticate_result` packets.
* Route room packets to room session handling.
* Route gameplay packets to gameplay session handling.
* Route debug shape catalog and debug status packets to gameplay/devtools handling.
* Route player pause state packets to gameplay session handling.
* Route telemetry pong packets to telemetry consumers.
* Emit an unknown-packet signal for recognized envelopes with unhandled packet types.
* Keep packet-family routing separate from raw WebSocket transport and payload-specific packet readers.

## Does not own

* Raw WebSocket connection lifecycle.
* WebSocket polling.
* Packet JSON parsing.
* Packet encode/decode result types.
* Packet schema source-of-truth files.
* Generated packet constant ownership.
* Server packet production.
* Server room authority.
* Server gameplay authority.
* Gameplay state application.
* Payload-specific packet reader behavior.
* World sync or entity rendering.
* HUD or menu presentation.
* Outbound packet construction.
* Outbound send timing.
* Auth token verification.
* Rails account identity.
* Persistent player data.
* Devtools command authority.

## Domain roles

### Decoded packet handoff

`NetworkClient` emits `packet_received(packet)` only after a raw WebSocket text message has passed through client packet decode and envelope validation.

`ClientConnectionService._on_packet_received(packet)` receives that dictionary and delegates to `ServerPacketDispatcher.dispatch(packet)`.

Inbound routing therefore assumes:

```text
packet is a Dictionary
packet has a non-empty String type field
payload envelope validation has already happened
```

Payload-specific validation still belongs to later packet readers or consumers.

### Packet type classification

`ServerPacketRouter` is the pure classification helper.

It reads the packet type through:

```gdscript
packet.get(Packets.FIELD_TYPE, "")
```

and compares the value against generated packet type constants from:

```text
client/scripts/generated/networking/packets/packets.gd
```

Current classified inbound packet types are:

```text
room_snapshot
authenticate_result
room_state_changed
room_error
state
debug_shape_catalog
debug_status
player_pause_state
telemetry_pong
```

The router does not emit signals, mutate state, or inspect packet payload contents beyond the packet type.

### Dispatcher signal fanout

`ServerPacketDispatcher` owns the ordered classification chain and typed signal emission.

Current dispatcher outputs are:

```text
room_snapshot_received(packet)
authenticate_result_received(packet)
room_state_changed(packet)
room_error_received(packet)
gameplay_state_received(packet)
debug_shape_catalog_received(packet)
debug_status_received(packet)
player_pause_state_received(packet)
telemetry_pong_received(packet)
unknown_packet_received(packet)
```

The dispatcher does not know which application subsystem will consume each signal. It only converts packet-type classification into a signal.

### Connection-service signal bridge

`ClientConnectionService` creates and owns the dispatcher instance.

It connects dispatcher signals to local handlers, then re-emits service-level signals with the same packet dictionary.

This keeps callers attached to one public networking facade instead of directly depending on `NetworkClient` or `ServerPacketDispatcher`.

### Websocket auth result cache

`ClientConnectionService` handles `authenticate_result` specially because websocket auth state is connection-level state.

On `authenticate_result`, the connection service updates:

```text
websocket_auth_authenticated
websocket_auth_user_id
websocket_auth_display_name
```

and emits:

```text
websocket_auth_result_received(packet)
```

The connection service does not verify the token. It only records the result returned by the game server.

### Session network handoff

`SessionNetworkController` connects to connection-service signals in three groups:

```text
connect_connection_signals()
connect_room_signals()
connect_gameplay_signals()
```

Connection-level signals handle:

```text
connected
closed
packet_parse_failed
unknown_packet_received
websocket_auth_result_received
```

Room signals handle:

```text
room_snapshot_received
room_state_changed
room_error_received
```

Gameplay signals handle:

```text
gameplay_state_received
debug_shape_catalog_received
debug_status_received
player_pause_state_received
```

`SessionNetworkController` is the bridge from networking events into room and gameplay session controllers. It does not classify packet types itself.

### Room packet handoff

Room packets route through `SessionNetworkController` into `RoomSessionController`.

Current handoff:

```text
room_snapshot_received
-> SessionNetworkController._on_room_snapshot_received
-> RoomSessionController.handle_room_snapshot

room_state_changed
-> SessionNetworkController._on_room_state_changed
-> RoomSessionController.handle_room_state_changed

room_error_received
-> SessionNetworkController._on_room_error_received
-> RoomSessionController.handle_room_error
```

`RoomSessionController` owns lobby/session consequences such as applying room snapshots, tracking latest room state, caching match result data from snapshots, showing room errors, and sending client config after room entry when needed.

Inbound packet routing does not own those consequences.

### Gameplay packet handoff

Gameplay state packets route through `SessionNetworkController` into `GameplaySessionController`.

Current handoff:

```text
gameplay_state_received
-> SessionNetworkController._on_gameplay_state_received
-> GameplaySessionController.handle_gameplay_state
```

`GameplaySessionController` gates gameplay state with `accepts_gameplay_packets`.

Gameplay packet application continues through gameplay runtime documentation after this point. Inbound routing only delivers the packet.

### Debug packet handoff

Debug shape catalog and debug status packets route through the same gameplay session controller because current devtools presentation is composed inside gameplay session context.

Current handoff:

```text
debug_shape_catalog_received
-> SessionNetworkController._on_debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

debug_status_received
-> SessionNetworkController._on_debug_status_received
-> GameplaySessionController.handle_debug_status_packet
```

Debug packet routing does not grant authority to the client. Server-side devtools authority remains server-owned.

### Player pause packet handoff

Player pause state packets route through gameplay session handling:

```text
player_pause_state_received
-> SessionNetworkController._on_player_pause_state_received
-> GameplaySessionController.handle_player_pause_state
```

`GameplaySessionController` applies the same gameplay-packet acceptance gate used for normal gameplay state before forwarding pause state to gameplay composition.

### Telemetry packet handoff

`telemetry_pong` packets are routed by the dispatcher and re-emitted by the connection service.

`WorldTelemetryContext` connects directly to:

```text
ClientConnectionService.telemetry_pong_received
```

and applies the pong packet to network telemetry metrics.

Telemetry pong handling is diagnostic. It does not require room membership, does not mutate gameplay state, and does not route through normal gameplay state application.

### Unknown packet fallback

If the packet envelope is valid but no current router predicate matches the packet type, the dispatcher emits:

```text
unknown_packet_received(packet)
```

`SessionNetworkController` currently logs the unknown-packet event through its configured logger.

Unknown packets are not applied to gameplay, room, auth, or telemetry state.

## Protocols and APIs

### Inbound routing surface

The inbound routing surface is the client-side handling path for decoded server packets.

The surface is consumed by client session controllers and direct consumers such as telemetry. The game server owns authority behind the packets. Data crossing this boundary is a decoded packet dictionary whose `type` field has already passed envelope validation.

Inbound routing explicitly does not own the packet schema, the raw transport, or the domain consequences of applying a packet.

### Routing sequence

Normal decoded packet sequence:

```text
NetworkClient.poll()
-> raw WebSocket text received
-> PacketCodec.decode(text)
-> NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter checks packet type
-> typed dispatcher signal emitted
-> ClientConnectionService typed signal emitted
-> owning session/controller handles the packet
```

Packet parse failures do not enter this routing path. They are emitted separately as:

```text
packet_parse_failed(text)
```

### Current inbound packet routes

```text
authenticate_result
-> websocket_auth_result_received
-> SessionNetworkController websocket auth gate

room_snapshot
-> room_snapshot_received
-> RoomSessionController.handle_room_snapshot

room_state_changed
-> room_state_changed
-> RoomSessionController.handle_room_state_changed

room_error
-> room_error_received
-> RoomSessionController.handle_room_error

state
-> gameplay_state_received
-> GameplaySessionController.handle_gameplay_state

debug_shape_catalog
-> debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

debug_status
-> debug_status_received
-> GameplaySessionController.handle_debug_status_packet

player_pause_state
-> player_pause_state_received
-> GameplaySessionController.handle_player_pause_state

telemetry_pong
-> telemetry_pong_received
-> WorldTelemetryContext._on_telemetry_pong_received

unmatched packet type
-> unknown_packet_received
-> SessionNetworkController logs unknown packet
```

### Auth gate interaction

Inbound routing participates in multiplayer boot gating only by delivering connection and auth signals.

Current multiplayer boot behavior is owned by `SessionNetworkController` and `ShellBootFlow`:

```text
connected + pending multiplayer request + websocket auth already authenticated
-> send pending request

connected + pending multiplayer request + websocket auth not authenticated
-> wait for authenticate_result

authenticate_result authenticated=true
-> send pending request

authenticate_result authenticated=false, error_code=token_verification_unavailable
-> send pending request so server-side admission can fail explicitly

authenticate_result authenticated=false, other error
-> keep pending multiplayer request unsent
```

Inbound routing does not decide account validity. It only delivers `authenticate_result` and stores connection-level auth result state.

### Gameplay packet acceptance

Inbound routing can deliver gameplay packets before the gameplay session is ready to consume them.

`GameplaySessionController` owns the current acceptance gate:

```text
accepts_gameplay_packets
```

Room snapshot and room state handling call `begin_accepting_gameplay_packets()` when room state reaches `InGame`.

This means inbound routing is intentionally simple: it delivers typed packets, while gameplay session ownership decides whether to ignore or apply them.

## Data ownership

Inbound packet routing owns only transient routing state and connection-level websocket auth result cache.

Owned state:

```text
websocket_auth_authenticated
websocket_auth_user_id
websocket_auth_display_name
```

Transient routed data:

```text
packet Dictionary
packet type String
typed signal payload
```

Inbound routing does not persist packet data.

Packet types and field-name constants come from generated client packet helpers:

```text
client/scripts/generated/networking/packets/packets.gd
```

Those helpers are generated from shared packet source files under:

```text
shared/packets/
```

Generated output and packet source-of-truth ownership belong to protocol/data documentation, not this client service doc.

## Code map

### Primary inbound routing files

* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/networking/inbound/server_packet_dispatcher.gd`
* `client/scripts/networking/inbound/server_packet_router.gd`

### Transport boundary

* `client/scripts/networking/network_client.gd`
* `client/scripts/networking/packets/packet_codec.gd`
* `client/scripts/networking/packets/packet_decode_result.gd`
* `client/scripts/networking/packets/packet_encode_result.gd`

### Session consumers

* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/room_session_controller.gd`
* `client/scripts/session/gameplay_session_controller.gd`

### Downstream room and lobby consumers

* `client/scripts/lobby/lobby_flow.gd`
* `client/scripts/lobby/lobby_shell_flow.gd`
* `client/scripts/lobby/multiplayer_lobby_presenter.gd`
* `client/scripts/lobby/multiplayer_dialog_status_presenter.gd`

### Downstream gameplay consumers

* `client/scripts/gameplay/state/gameplay_state_flow.gd`
* `client/scripts/gameplay/state/gameplay_state_packet_reader.gd`
* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/gameplay/runtime/`
* `client/scripts/world/world_sync.gd`

### Downstream devtools and telemetry consumers

* `client/scripts/devtools/telemetry/world_telemetry_context.gd`
* `client/scripts/devtools/telemetry/network_telemetry_metrics.gd`
* `client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd`
* `client/scripts/devtools/context/`
* `client/scripts/devtools/hitboxes/`

### Generated and source boundaries

* `client/scripts/generated/networking/packets/packets.gd`
* `shared/packets/lobby.toml`
* `shared/packets/gameplay.toml`
* `shared/packets/debug.toml`
* `shared/packets/outputs.toml`

### Non-ownership boundaries

* `services/game-server/internal/networking/` owns server websocket session routing.
* `services/game-server/internal/rooms/` owns room membership and room lifecycle authority.
* `services/game-server/internal/game/` owns authoritative gameplay state production.
* `services/game-server/internal/devtools/` owns server-authoritative devtools command handling.
* `services/game-server/internal/protocol/packetcodec/` owns server packet wire encode/decode behavior.
* `docs/protocol/` owns protocol-level packet behavior.
* `docs/data/` owns packet source-of-truth and generation pipeline documentation.

## Tests

Relevant tests include:

* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/test_room_session_controller.gd`
* `client/tests/unit/test_packet_codec.gd`
* `client/tests/unit/test_gameplay_state_packet_reader.gd`
* `client/tests/unit/test_gameplay_state_apply_flow.gd`
* `client/tests/unit/devtools/telemetry/test_network_telemetry_metrics.gd`
* `client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd`
* `client/tests/unit/devtools/hitboxes/test_debug_shape_catalog_packet_reader.gd`
* `client/tests/unit/devtools/debug_status_packet_reader_test.gd`

`test_session_network_controller.gd` covers the most important current routing-adjacent behavior: connection handling, websocket auth gating, pending boot dispatch, and auth-failure behavior.

`test_packet_codec.gd` verifies the decode/envelope behavior that happens before inbound routing begins.

Current direct coverage for `ServerPacketRouter` and `ServerPacketDispatcher` is thin. Their behavior is simple packet-type classification and signal fanout, but adding focused tests would be reasonable if more inbound packet types are added.

## Related docs

* [Networking Flow](./!README.md)
* [WebSocket Connection Lifecycle](websocket-connection-lifecycle.md) - Raw client WebSocket lifecycle and packet decode/encode boundary.
* [Outbound Packet Sending](outbound-packet-sending.md) - Client outbound packet construction and send handoff.
* [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
* [Room Session State](../app-shell-and-session/room-session-state.md)
* [Auth Session Flow](../auth-session-flow.md)
* [Lobby Flow](../lobby-flow/!README.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Gameplay State Application](../gameplay-runtime/gameplay-state-application.md)
* [Gameplay Event Presentation](../gameplay-event-presentation/!README.md)
* [World Sync](../world-sync/!README.md)
* [Realtime Websocket Protocol](../../../protocol/stubs/realtime-websocket-protocol.md) - Stub: realtime websocket protocol documentation.
* [Gameplay Packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay packet documentation.
* [Lobby Packets](../../../protocol/stubs/lobby-packets.md) - Stub: lobby packet documentation.
* [Devtools Packets](../../../protocol/stubs/devtools-packets.md) - Stub: devtools packet documentation.
* [Packet Schema Pipeline](../../../data/stubs/packet-schema-pipeline.md) - Stub: shared packet schema and generated output documentation.

## Notes

Legacy architecture notes correctly identified `client/scripts/networking/inbound/` as the client server-packet classification and dispatch boundary. This document rewrites that fact against the current client implementation.

`ClientConnectionService` currently acts as the public networking facade for both inbound and outbound flow. That is current implementation, not a reason to merge inbound and outbound docs. Inbound routing and outbound sending have different call directions, packet ownership, and downstream consequences.

Telemetry pong is routed through the same inbound dispatcher but consumed directly by telemetry context rather than through `SessionNetworkController`.

Gameplay packet acceptance is intentionally not handled by the router. The router classifies packets; `GameplaySessionController` decides whether gameplay packets are currently accepted.
