# Devtools Packet Protocol

Parent index: [Design](./!INDEX.md)

## Purpose

This document describes the current devtools packet protocol for Space Rocks.

It covers debug command requests, server-emitted debug readouts, source-of-truth packet files, generated outputs, routing order, authority boundaries, and the relationship between devtools packets and normal gameplay protocol paths.

## Overview

See also: [Game Control Devtools Adapter](../server/game-control-devtools-adapter.md)

The devtools packet protocol currently spans two transports: runtime measurement uses the dedicated `sr.tooling` WebRTC DataChannel, while legacy runtime debug commands and readouts still use the normal WebSocket path pending migration. It is not a developer console protocol and not a parallel gameplay authority layer.

Current legacy debug-command flow:

```text
client devtools input or window control
-> client devtools command/placement/readout context
-> generated packet helper or devtools packet builder
-> ClientConnectionService.send_packet(...)
-> ClientPacketSender
-> NetworkClient
-> PacketCodec
-> WebSocket text message
-> game-server inbound packet classification
-> packet decode into devtools.DebugCommand
-> game.NewControl(room.GameInstance())
-> devtools.NewController(...)
-> Controller.HandleCommand
-> capability-specific handler
-> game.Control
-> normal authoritative game state mutation
-> normal lane-native gameplay/debug output packets
-> client inbound packet routing
-> devtools readmodels, overlays, or window presentation
```

Current measurement flow:

```text
client measurement coordinator
-> generated tooling packet with request_id
-> ClientConnectionService.send_tooling_packet(...)
-> RealtimeTransportSession
-> sr.tooling
-> server tooling router
-> measurement controller
-> correlated measurement response or tooling_error
-> client tooling packet router
-> measurement coordinator and devtools presentation
```

The client may request debug behavior. The server owns whether the request is valid and what gameplay state changes. Client-side devtools packet construction does not confirm success by itself.

Devtools packets currently fall into four broad groups:

```text
client -> server runtime measurement and telemetry packets on sr.tooling
server -> client runtime measurement and telemetry packets on sr.tooling
client -> server legacy debug command packets on WebSocket
server -> client legacy debug status and shape catalog packets on WebSocket
```

Measurement and telemetry packet schemas are defined in `shared/packets/tooling.toml`. Runtime debug command and readout schemas remain in `shared/packets/debug.toml` until their migration slice moves physical routing and schema ownership. The authoritative migration inventory is [Tooling Channel Migration Contract](./tooling-channel-migration-contract.md).

## Debug-only scope

Devtools packets are for local development and diagnostics.

They may request or report:

```text
debug invincibility
debug infinite lives
debug freeze controls
debug kill requests
debug spawn requests
debug pickup spawn requests
debug continuous bullet streams
debug respawn requests
debug score/lives changes
debug clear bullets
debug clear asteroids
debug status output
debug shape catalog output
```

They must not:

```text
replace normal gameplay packets
grant gameplay authority to the client
duplicate gameplay mutation rules client-side
become player-facing product protocol
make devtools readouts the authoritative state model
bypass game-owned mutation seams
treat generated packet schemas as behavior ownership
```

The packet protocol defines how debug requests and readouts cross the client/server boundary. The semantic result of a request belongs to server devtools handlers and the owning gameplay seams.

## Participating systems

The current protocol participants are:

```text
client devtools
client networking
shared packet schema pipeline
game-server networking
game-server devtools
game-server Controller/Target/Control boundary
client devtools presentation
```

Client devtools collect local intent, build packet dictionaries, and display server-fed state.

Client networking owns encode/send/decode/dispatch behavior.

The shared packet schema pipeline owns packet type strings, generated constants, generated Go structs, and generated GDScript packet builders.

Game-server networking owns the `sr.tooling` router and channel send boundary, plus the temporary WebSocket envelope classification and outbound debug writing used by legacy runtime devtools.

Game-server devtools own command dispatch, debug status projection, shape catalog projection, and devtools-specific runtime state such as continuous bullet stream runtime.

Controller owns command policy, the Control adapter exposes authoritative capabilities, and gameplay state remains server-owned. Devtools should call the current Controller/Target/Control boundary instead of duplicating gameplay rules in normal game systems.

Client devtools presentation owns the devtools window, debug readmodels, hitbox overlays, telemetry overlays, and dev labels.

## Protocol authority

Server authority is the core rule.

The client can send:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_kill_player
debug_spawn_entity
debug_spawn_pickup
debug_begin_continuous_bullet_stream
debug_respawn_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
```

The server decides whether those commands have an effect.

The current legacy WebSocket command route requires a current room and a current game player ID before decoding and applying a devtools command. If either is missing, the packet is consumed and no gameplay command is applied. This is a migration limitation, not the final authorization contract: the `sr.tooling` route will authorize by session capability and room attachment without requiring gameplay participation.

The server emits:

```text
debug_status
debug_shape_catalog
```

The client consumes those outputs as diagnostic presentation data. They do not replace lane-native gameplay output.

## Source-of-truth files

Current packet schemas are split by transport ownership:

```text
shared/packets/tooling.toml
-> active sr.tooling measurement and telemetry packets

shared/packets/debug.toml
-> legacy runtime debug commands and readouts pending migration
```

Packet output routing is defined in:

```text
shared/packets/outputs.toml
```

Current generated devtools-related outputs are:

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/protocol/tooling/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

`debug.toml` owns these generated struct shapes:

```text
DebugCommand
DebugStatus
DebugShapePoint
DebugShapeDefinition
DebugShapeCatalogPacket
DebugStatusPacket
```

`outputs.toml` selects which debug packet types, structs, and builders are emitted into the client and server generated outputs.

The packet schema owns wire shape and generated constants. It does not own command behavior, gameplay mutation, room eligibility, collision effects, scoring rules, respawn rules, or presentation layout.

## Command request surface

The current `DebugCommand` shape supports a broad command envelope:

```text
type
target_player_id
target_scope
entity_type
pickup_type
x
y
has_direction
direction_x
direction_y
freeze_target
score
amount
lives
```

Not every field is valid for every command.

Current field usage:

| Field                        | Role                                                                |
| ---------------------------- | ------------------------------------------------------------------- |
| `type`                       | Command packet type string.                                         |
| `target_player_id`           | Player target for player-only commands.                             |
| `target_scope`               | Scope selector, currently including `all_players`.                  |
| `entity_type`                | Spawn target for debug entity placement.                            |
| `pickup_type`                | Pickup type for debug pickup placement.                             |
| `x`, `y`                     | Server-space placement coordinates.                                 |
| `has_direction`              | Whether placement included a drag direction.                        |
| `direction_x`, `direction_y` | Server-space direction vector for directional placement or streams. |
| `freeze_target`              | World-freeze sub-target selector for targeted freeze controls.      |
| `score`                      | Absolute score value for set-score commands.                        |
| `amount`                     | Delta value for add-score or add-lives commands.                    |
| `lives`                      | Absolute lives value for set-lives commands.                        |

`target_player_id` is still present for devtools player-only compatibility. Normal gameplay targeting uses canonical `target_kind` and `target_id` on gameplay target request/state paths. Devtools readouts should prefer canonical target kind/id when displaying target state.

## Targeting and scope

Player-targeted commands resolve through explicit target data, not client authority.

Current server target resolution is:

```text
if target_scope == "all_players":
    use current game-owned devtools target player ids
else:
    use target_player_id
    if target_player_id is empty:
        fall back to requesting player id
```

`all_players` is a scope, not a fake player ID.

Client target resolution must preserve these boundaries:

```text
explicit selected player wins
Game Target resolves only when canonical target_kind is player
All Players serializes as target_scope = "all_players"
non-player canonical targets do not become target_player_id
local player fallback is allowed only where the command path supports fallback
```

Score, lives, kill, respawn, invincibility, infinite lives, and player freeze are player-targeted controls.

World freeze and clear-entity commands are room/global controls.

Spawn placement commands use placement coordinates and entity/pickup fields rather than normal gameplay target selection.

## Placement command flow

Placement commands are built from server-space positions collected by client devtools input.

Current client placement packet builders can emit:

```text
debug_spawn_entity
debug_spawn_pickup
debug_begin_continuous_bullet_stream
```

Entity placement uses:

```text
type = "debug_spawn_entity"
entity_type = "player" | "asteroid" | "bullet"
x
y
has_direction
direction_x
direction_y
target_player_id, when relevant
```

Pickup placement uses:

```text
type = "debug_spawn_pickup"
pickup_type
x
y
```

Continuous bullet stream placement uses:

```text
type = "debug_begin_continuous_bullet_stream"
x
y
has_direction = true
direction_x
direction_y
```

Continuous bullet stream packets require a non-zero direction. If client placement does not produce direction data, the client builder returns an empty packet and nothing is sent.

The server owns the actual spawn behavior. Client placement only supplies requested coordinates and optional direction.

## Server command routing

Game-server WebSocket reads first decode a lightweight packet envelope to inspect `type`.

Current routing order is:

```text
simple devtools command classification
placement devtools command classification
remaining devtools command classification
normal game.ClientPacket decode
auth packet handling
telemetry packet handling
lobby packet handling
gameplay packet handling
```

Devtools command packets are intentionally classified before normal gameplay packet decode. This keeps devtools command structs in `internal/devtools` and prevents debug command fields from becoming part of normal `game.ClientPacket` ownership.

Current devtools command groups in inbound routing are:

```text
simple:
  toggle_debug_invincible
  toggle_debug_infinite_lives
  toggle_debug_freeze_world
  toggle_debug_freeze_player
  debug_kill_player
  debug_set_score
  debug_add_score
  debug_set_lives
  debug_add_lives
  debug_clear_bullets
  debug_clear_asteroids

placement:
  debug_spawn_entity
  debug_spawn_pickup

remaining:
  debug_begin_continuous_bullet_stream
  debug_respawn_player
```

Once a packet is classified as devtools, the active routed path decodes it into `devtools.DebugCommand`, constructs `game.NewControl(room.GameInstance())`, constructs `devtools.NewController(...)`, and dispatches through `Controller.HandleCommand`.

`devtools.HandleCommand(...)` remains only as a direct-call compatibility wrapper.

`Controller.HandleCommand` dispatches by command type to the owning server devtools handler. Those handlers then use the game-owned Control adapter for authoritative gameplay mutation.

Devtools commands do not route through normal `Game.HandlePacket` gameplay packet handling.

## Server output packets

### Debug status

`debug_status` reports current debug-control state. It is a WebSocket devtools readout packet when its write-loop delivery path is active.

Current packet shape:

```text
type = "debug_status"
debug_status = status for the receiving/current player
debug_statuses = map of every match player id to that player's debug status
```

Current debug status fields are:

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

The server can send debug status when:

```text
room exists
room has a game instance
devtools are enabled
room state is InGame or GameOver
session has a current game player id
```

The session write loop triggers normal gameplay presentation output every server write tick; active gameplay lane bytes are delivered over lane-specific WebRTC gameplay DataChannels using the current protocol channel policy. `debug_status` is a WebSocket devtools readout packet when its write-loop delivery path is active. Current docs must not claim periodic `debug_status` delivery unless the active write loop calls the debug status builder.

The client uses `debug_status` for receiver/global status labels and `debug_statuses` for per-player target/status rows.

### Debug shape catalog

`debug_shape_catalog` reports shape definitions used by server hitbox overlay presentation.

Current packet shape:

```text
type = "debug_shape_catalog"
shapes = map of shape id to debug shape definition
```

Each shape definition contains:

```text
id
kind
shape_type
points
```

The server builds this output from the collision shape catalog and converts it through the devtools shape catalog builder.

The shape catalog is sent from the server write loop when eligible and when the current room ID differs from the last room ID that received a catalog on that connection. It is diagnostic shape metadata, not gameplay authority.

## Client inbound routing

Client inbound routing uses the normal packet receive path:

```text
NetworkClient.poll()
-> PacketCodec.decode(text)
-> NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter
-> typed signal
-> devtools consumer
```

Devtools-related inbound routes currently include:

```text
debug_status
-> NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter classifies debug_status
-> ClientInboundCoordinator.debug_status_received(packet)
-> ClientConnectionService._on_debug_status_received(packet)
-> ClientConnectionService.debug_status_received(packet)
-> SessionNetworkController
-> GameplaySessionController.handle_debug_status_packet
-> GameplayComposition.apply_devtools_debug_status_packet
-> GameplayDevtoolsContext.apply_debug_status_packet
-> devtools readmodel refresh built from lane-applied state

debug_shape_catalog
-> NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter classifies debug_shape_catalog
-> ClientInboundCoordinator.debug_shape_catalog_received(packet)
-> ClientConnectionService._on_debug_shape_catalog_received(packet)
-> ClientConnectionService.debug_shape_catalog_received(packet)
-> SessionNetworkController
-> GameplaySessionController.handle_debug_shape_catalog_packet
-> GameplayComposition.apply_devtools_debug_shape_catalog_packet
-> server hitbox overlay catalog state

telemetry_pong
-> ClientConnectionService.telemetry_pong_received
-> WorldTelemetryContext
-> NetworkTelemetryMetrics
```

`telemetry_pong` is included here because devtools telemetry consumes it, but it remains a normal gameplay telemetry packet.

## Client outbound routing

Most client devtools commands converge on the normal outbound packet path:

```text
devtools hotkey or window action
-> devtools command, placement, or overlay context
-> generated packet builder or devtools packet builder
-> ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
-> PacketCodec.encode(packet)
-> WebSocketPeer.send_text(wire_message)
```

Some commands use generated helpers through `DevtoolsClientPackets`.

Other commands build dictionaries directly in focused devtools builders, especially placement, respawn, and continuous bullet stream commands.

Both paths must converge on the normal packet sender and packet codec. Devtools should not create a parallel socket path.

## Build and runtime gates

Client-side gates are presentation/input gates only.

Current client gates include:

```text
public_build removes DevToggle0 through DevToggle9 input events
command sends require a configured connection path
placement sends require a non-empty placement result
continuous bullet stream sends require direction data
packet send requires an open WebSocket and successful encoding
```

Server-side outbound debug output builders check `devtools.Enabled()` before building debug status or debug shape catalog packets.

The server source also defines `Enabled()` and `ShouldHandleCommand(...)` for default and `nodevtools` builds. The current inbound command routing classifies devtools command packet types directly in `services/game-server/internal/networking/inbound/devtools.go`; documentation for server build gates should be verified against that route when changing or relying on disabled-command behavior.

No client-side gate should be treated as the authority boundary for gameplay-affecting devtools behavior.

## Relationship to real gameplay seams

Devtools packets route into real gameplay seams; they do not create alternate gameplay systems.

Required boundary rules:

```text
devtools command packets stay out of generated game packet structs
normal gameplay packets stay out of devtools command ownership
server devtools handlers own debug command dispatch
game-owned Control adapters expose narrow mutation capabilities
internal/game must not import internal/devtools
score and lives commands use the shared player counter seam
damage-related debug behavior uses damage/capability seams
freeze behavior uses gameplay suspension/simulation gates
clear commands mutate authoritative server state only
client observes results through state/debug packets
```

The packet protocol should remain a request/readout layer over real gameplay systems.

## Validation and verification

Run packet pipeline checks when changing `shared/packets/debug.toml`, `shared/packets/outputs.toml`, or generated devtools packet outputs:

```bash
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

Relevant server tests include:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
```

Relevant client tests include:

```text
client/tests/unit/devtools/gameplay_debug_flow_test.gd
client/tests/unit/devtools/context/test_devtools_command_context.gd
client/tests/unit/devtools/context/test_devtools_placement_context.gd
client/tests/unit/test_gameplay_devtools_context.gd
client/tests/unit/test_devtools_target_resolver.gd
client/tests/unit/test_devtools_player_target_model.gd
client/tests/unit/devtools/debug_status_packet_reader_test.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd
client/tests/unit/test_packet_codec.gd
```

Run server tests after changing command classification, command dispatch, command effects, debug status projection, shape catalog output, build tags, Controller dispatch, or Control adapter behavior.

Run client tests after changing packet builders, devtools command routing, inbound packet routing, debug status readers, target readmodels, placement builders, overlays, or packet codec behavior.

## Code map

Packet source files:

```text
shared/packets/debug.toml
shared/packets/gameplay.toml
shared/packets/outputs.toml
```

Generated packet outputs:

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/game/packets.go
```

Client outbound devtools packet paths:

```text
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/packets/packet_codec.gd
```

Client inbound devtools packet paths:

```text
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/client_connection_service.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/gameplay/gameplay_composition.gd
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/debug_status_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
```

Client devtools presentation consumers and readmodel builders:

```text
client/scripts/protocol/realtime/devtools_lane_state_adapter.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_display_refresh_flow.gd
client/scripts/devtools/devtools_player_target_model.gd
client/scripts/devtools/devtools_target_resolver.gd
client/scripts/devtools/telemetry/world_telemetry_context.gd
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
client/scripts/devtools/player_labels/player_dev_labels_context.gd
```

Server inbound routing:

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/protocol/packetcodec/
```

Server devtools command handling:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/clear_entities.go
```

Server debug output:

```text
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

Server build gates:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Game-owned devtools seams:

```text
services/game-server/internal/game/control*.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/packets.go does not own devtools command packets.
services/game-server/internal/game/ owns simulation authority, not devtools protocol policy.
client/scripts/world/ owns world presentation, not debug command authority.
client/scripts/ui/ owns player-facing UI, not devtools packet authority.
shared/packets/ owns packet shape, not runtime command semantics.
```

## Related docs

* [Devtools](../!INDEX.md)
* [Client Devtools](../client/!INDEX.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Telemetry Overlays](../client/telemetry-overlays.md)
* [Hitbox Overlays](../client/hitbox-overlays.md)
* [Server Devtools](../server/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Client](../../services/client/!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
* [Outbound Packet Sending](../../services/client/networking-flow/outbound-packet-sending.md)
* [Devtools And Telemetry](../../planning/devtools/devtools-and-telemetry.md)

## Notes

The `sr.tooling` transport foundation is implemented as the mandatory negotiated id 9 channel: reliable, ordered, bidirectional, and ready alongside the eight gameplay channels. Tooling receive routing is separated before normal gameplay dispatch. Runtime measurement is active end to end on this channel. Telemetry packet routing exists, but the production telemetry provider and the legacy overlay ping migration remain incomplete. Runtime debug commands, `debug_status`, and `debug_shape_catalog` still use WebSocket pending the migration defined in [Tooling Channel Migration Contract](./tooling-channel-migration-contract.md).

Unexpected required-channel closure preserves the WebSocket/session/room/game context while replacing only the WebRTC peer. Recovery uses a 10-second deadline; success preserves the active match and requests fresh world, overlay, and session baselines, while failure disables only single-player replay.

`debug_status` and `debug_shape_catalog` are devtools readout packets. They help the client render debug controls and overlays, but they do not replace normal lane-native realtime packets.

Measurement and telemetry packet schemas belong to `shared/packets/tooling.toml`. Legacy runtime debug commands and readouts remain in `shared/packets/debug.toml` until their route and schema ownership migrate together.

`target_player_id` remains a devtools compatibility field for player-only debug commands. New gameplay targeting should continue to use canonical target kind/id fields instead of extending `target_player_id` further.
