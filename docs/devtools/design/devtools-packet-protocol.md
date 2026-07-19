# Devtools Packet Protocol

Parent index: [Design](./!INDEX.md)

## Purpose

This document describes the current devtools packet protocol for Space Rocks.

It covers debug command requests, server-emitted debug readouts, source-of-truth packet files, generated outputs, routing order, authority boundaries, and the relationship between devtools packets and normal gameplay protocol paths.

## Overview

See also: [Game Control Devtools Adapter](../server/game-control-devtools-adapter.md)

Runtime measurement, runtime debug commands, and developer readouts use the dedicated reliable ordered `sr.tooling` WebRTC DataChannel. Continuous telemetry/ping is the remaining legacy transport slice. The devtools protocol is not a parallel gameplay authority layer.

Current command flow:

```text
client devtools action
-> generated/request payload with request_id and trace_id
-> ClientConnectionService.send_tooling_packet(...)
-> sr.tooling
-> server tooling policy, room, and capability preflight
-> existing devtools Controller.HandleCommand
-> game-owned Control seam
-> tooling_command_result or tooling_error
-> ToolingPacketRouter
```

Current developer-readout flow:

```text
active room + ready sr.tooling
-> debug_status_subscribe(request_id)
-> bounded-cadence debug_status pushes

active room + ready sr.tooling
-> debug_shape_catalog_request(request_id)
-> correlated debug_shape_catalog response or tooling_error

server readout packet
-> ToolingPacketRouter
-> ClientConnectionService keeps the existing public readout signal
-> existing gameplay/devtools presentation owner
```

Measurement, telemetry, and readout request schemas are defined in `shared/packets/tooling.toml`. Runtime debug commands and readout response structs remain in `shared/packets/debug.toml`; their physical route is `sr.tooling`. The authoritative migration inventory is [Tooling Channel Migration Contract](./tooling-channel-migration-contract.md).

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

Game-server networking owns the `sr.tooling` router, capability/attachment preflight, readout subscription/request state, and channel send boundary. WebSocket remains responsible for auth, signaling, room/session setup, and the legacy continuous-telemetry path.

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

The `sr.tooling` command route capability-checks the session and requires room attachment before decoding and applying a devtools command. The current `GamePlayerID` is passed through when present but is not required merely to dispatch a room-global or explicit-target command.

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
-> sr.tooling measurement, telemetry, debug-status subscription, and shape-catalog request schemas

shared/packets/debug.toml
-> runtime debug commands and developer readout response structs; commands and readouts use sr.tooling
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
request_id
trace_id
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
| `request_id`                 | Correlation identifier copied to the terminal result or error.      |
| `trace_id`                   | Diagnostic trace identifier carried with the command request.       |
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

Runtime debug commands no longer enter through the game-server WebSocket read loop. The `sr.tooling` route applies command policy before decoding the debug command.

Current runtime command routing order is:

```text
client command payload with request_id and trace_id
sr.tooling
tooling packet policy and capability preflight
room attachment check
decode into devtools.DebugCommand
existing devtools Controller.HandleCommand
capability-specific handler
correlated tooling_command_result or tooling_error
```

Devtools command packets remain separate from normal gameplay packet decode. This keeps devtools command structs in `internal/devtools` and prevents debug command fields from becoming part of normal `game.ClientPacket` ownership.

Current devtools command groups are:

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

Once a packet passes tooling preflight, the active routed path decodes it into `devtools.DebugCommand` and dispatches through the existing `Controller.HandleCommand` seam.

`devtools.HandleCommand(...)` remains only as a direct-call compatibility wrapper.

`Controller.HandleCommand` dispatches by command type to the owning server devtools handler. Those handlers then use the game-owned Control adapter for authoritative gameplay mutation.

Devtools commands do not route through normal `Game.HandlePacket` gameplay packet handling.

## Server output packets

### Debug status

`debug_status` reports authoritative debug-control state through a connection-local `sr.tooling` subscription.

```text
debug_status_subscribe(request_id)
-> first eligible tooling-router tick emits immediately
-> later pushes occur every eight tooling-router ticks
-> debug_status.request_id echoes the subscription request ID

debug_status_unsubscribe(request_id)
-> clears the subscription
```

Eligibility requires an attached room, `tooling.read`, an active game instance, devtools enabled, and room state `InGame` or `GameOver`. A `GamePlayerID` is not required merely to receive room-global state and the per-player status map.

The packet contains:

```text
type = debug_status
request_id = active subscription correlation
debug_status = receiving gameplay identity status when available
debug_statuses = map of match player IDs to authoritative debug status
```

### Debug shape catalog

`debug_shape_catalog` is a one-shot correlated `sr.tooling` response.

```text
debug_shape_catalog_request(request_id)
-> tooling.read and room checks
-> BuildDebugShapeCatalogPacket
-> debug_shape_catalog with matching request_id
-> or tooling_error(debug_shape_catalog_unavailable)
```

The client requests it once per active room. The packet contains reusable shape definitions, not live entity state or collision authority. The WebSocket send-once state and write-loop push are removed.

## Client inbound routing

Developer readouts use the tooling receive path:

```text
sr.tooling
-> RealtimeTransportSession.tooling_packet_received
-> ClientConnectionService._on_tooling_packet_received
-> ToolingPacketRouter.dispatch
-> typed signal
-> existing devtools consumer
```

Current readout routes are:

```text
debug_status
-> ToolingPacketRouter.debug_status_received
-> ClientConnectionService.debug_status_received
-> SessionNetworkController
-> GameplaySessionController
-> GameplayComposition
-> GameplayDevtoolsContext
-> debug status readmodel/window refresh

debug_shape_catalog
-> ToolingPacketRouter.debug_shape_catalog_received
-> ClientConnectionService.debug_shape_catalog_received
-> SessionNetworkController
-> GameplaySessionController
-> GameplayComposition
-> server hitbox catalog/overlay flow
```

The normal WebSocket `ServerPacketRouter`, `ServerPacketDispatcher`, and `ClientInboundCoordinator` no longer classify these readouts. `telemetry_pong` still follows its existing path until continuous telemetry/ping migration.

## Client outbound routing

Runtime client devtools commands use the tooling outbound path:

```text
devtools hotkey or window action
-> devtools command, placement, or overlay context
-> GameplayDebugFlow or DevConnectionService builds request_id/trace_id payload
-> ClientConnectionService.send_tooling_packet(packet)
-> sr.tooling
```

The server networking/tooling route capability-checks commands and readout requests. `ToolingPacketRouter` receives correlated command results/errors plus `debug_status` and `debug_shape_catalog`. Continuous telemetry/ping remains the next migration slice.

## Build and runtime gates

Client-side gates are presentation/input gates only.

Current client gates include:

```text
public_build removes DevToggle0 through DevToggle9 input events
command sends require a configured connection path
placement sends require a non-empty placement result
continuous bullet stream sends require direction data
runtime command send requires an available sr.tooling channel
```

Server-side outbound debug output builders check `devtools.Enabled()` before building debug status or debug shape catalog packets.

The server source also defines `Enabled()` and `ShouldHandleCommand(...)` for default and `nodevtools` builds. The current tooling command route applies those command gates through `services/game-server/internal/networking/tooling/`; documentation for server build gates should be verified against that route when changing or relying on disabled-command behavior.

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
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/packets/packet_codec.gd
```

Client inbound devtools packet paths:

```text
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/tooling_packet_router.gd
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

Server tooling command routing:

```text
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/preflight.go
services/game-server/internal/networking/tooling/commands.go
services/game-server/internal/protocol/tooling/packets_generated.go
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
services/game-server/internal/networking/tooling/router.go
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

The `sr.tooling` transport foundation is the mandatory negotiated id 9 channel: reliable, ordered, bidirectional, and ready alongside the eight gameplay channels. Runtime measurement, runtime debug commands, debug-status subscription, and shape-catalog request/response are active end to end on this channel. The production telemetry provider and legacy overlay ping migration remain incomplete.

Unexpected required-channel closure preserves the WebSocket/session/room/game context while replacing only the WebRTC peer. Recovery uses a 10-second deadline; success preserves the active match and requests fresh world, overlay, and session baselines, while failure disables only single-player replay.

`debug_status` and `debug_shape_catalog` are devtools readout packets. They help the client render debug controls and overlays, but they do not replace normal lane-native realtime packets.

Measurement, telemetry, and readout request schemas belong to `shared/packets/tooling.toml`. Runtime debug command and readout response schemas remain in `shared/packets/debug.toml`. Their wire transport is `sr.tooling`; later schema consolidation is separate work.

`target_player_id` remains a devtools compatibility field for player-only debug commands. New gameplay targeting should continue to use canonical target kind/id fields instead of extending `target_player_id` further.
