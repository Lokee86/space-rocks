# Tooling Channel Migration Contract

Parent index: [Design](./!INDEX.md)

## Purpose

This document is the authoritative inventory and migration contract for moving runtime developer tooling onto the dedicated `sr.tooling` WebRTC DataChannel.

It fixes the transport boundary, packet ownership, capability requirements, attachment requirements, request/result rules, and the exact current migration surface before command routing and telemetry plumbing are changed.

## Status

Phase 1 of the devtools transport migration is complete.

Implemented foundations:

```text
sr.tooling negotiated channel id 9
mandatory readiness with the eight gameplay channels
reliable, ordered, bidirectional delivery
lane-aware client and server routing
measurement request/response consumer
measurement report lifecycle
partial telemetry subscription router
```

Not yet migrated:

```text
runtime debug mutation commands
debug status subscription and delivery
debug shape catalog request and delivery
legacy world-overlay telemetry ping path
production telemetry snapshot provider
privileged capability enforcement
observer/ghost/active-operator attachment mechanics
```

## Locked transport boundary

After migration, the transport split is:

```text
WebSocket
-> authentication
-> WebRTC signaling
-> lobby and room setup
-> session setup and teardown
-> non-tooling control required before WebRTC readiness

sr.tooling
-> runtime measurement
-> runtime telemetry
-> runtime developer commands
-> developer readouts and catalogs
-> future room administration commands and readouts
```

Gameplay state continues to use its existing lane-native WebRTC channels. Tooling packets must never be inserted into gameplay world, overlay, session, event, asteroid, bullet, or lifecycle lanes.

The presence of `sr.tooling` is mandatory for connection readiness. Its presence does not grant privileged capabilities.

## Roles and capabilities

The locked privileged capability vocabulary is:

| Capability | Meaning |
| --- | --- |
| `tooling.read` | Read privileged room diagnostics, developer status, catalogs, and future developer inspection data. |
| `tooling.control` | Apply runtime developer mutations through authoritative owning seams. |
| `admin.control` | Apply future administrative actions through audited owning seams. |

The role vocabulary is:

```text
dev
admin
```

Roles grant capabilities; packet handlers authorize capabilities. Code must not authorize packets by checking a role name directly.

Runtime measurement and the public-safe telemetry subset are an explicit exception. They require an eligible connected session and any packet-specific room attachment, but they do not require `dev` or `admin` clearance.

## Capability grant source

The capability-checking seam is implemented independently from the source that grants capabilities.

Current migration policy preserves existing devtools access:

```text
every connected session receives tooling.read
every connected session receives tooling.control
admin.control is not granted automatically
```

This temporary all-connections grant applies equally to authenticated accounts and guest/local connections. Packet handlers must still check the packet policy through the capability seam; they must not bypass authorization merely because the current grant provider allows everyone.

The eventual production grant source is an account flag or account-backed entitlement. Authentication will project that account flag into the session's tooling capabilities. Accounts without the flag, and guest connections without an equivalent explicit grant, will not receive privileged tooling capabilities. The exact persisted field name and account-management workflow are deferred until the account/admin implementation slice.

Changing from the temporary grant provider to the account-backed provider must not require changing packet policies or command handlers.

## Authorization order

Every client-to-server tooling packet follows this order:

```text
identify packet type
-> find packet policy
-> verify channel/session state
-> verify required room attachment
-> verify required capability
-> decode the packet-specific payload
-> validate request ownership and target scope
-> call the owning controller/service seam
-> emit exactly one terminal result or error when the packet is request-like
-> audit state-changing privileged actions
```

Unknown packet types, missing capabilities, invalid attachments, malformed payloads, invalid targets, and unavailable owning systems must be rejected with `tooling_error` and must not mutate state.

Capability checks occur server-side. Client build flags, hidden UI, removed hotkeys, or absent menu controls are not authorization boundaries.

## Attachment contract

One tooling session may attach to at most one room at a time.

Final room attachment modes are:

| Mode | Gameplay participation | Physical presence | Capacity/team/result effects |
| --- | --- | --- | --- |
| `observer` | None | None | None |
| `ghost_operator` | None | Non-interacting controllable presence | None |
| `active_operator` | None | Gameplay-interacting controllable presence | None; run becomes tooling-affected |
| `participant` | Normal | Normal player ship | Normal participation rules |

Tooling packet authorization uses session identity, room attachment, and capabilities. It must not require `GamePlayerID` merely to prove tooling access.

Consequences:

```text
room-global readouts work for observers
room-global commands work without a participant slot when authorized
explicit player targets work without the operator being a player
implicit "self" fallback is allowed only when an attached gameplay identity exists
a missing GamePlayerID must not silently block room-global tooling
```

Observer, ghost-operator, and active-operator attachments do not consume player capacity and do not affect teams, balancing, readiness, objectives, scoring, victory, rewards, stats, rankings, or match results.

## Packet ownership rules

`shared/packets/tooling.toml` is the target source of truth for all `sr.tooling` wire schemas.

During migration, existing runtime debug packet types remain generated from `shared/packets/debug.toml`. They retain their current type names and semantic ownership while their physical route moves. Schema ownership moves to `tooling.toml` only when the corresponding migration slice updates generated outputs and removes the old route.

Runtime command semantics remain owned by:

```text
services/game-server/internal/devtools
-> developer command policy and dispatch

services/game-server/internal/game/control*.go
-> authoritative gameplay mutation capabilities

services/game-server/internal/networking/tooling
-> transport routing, packet policy, subscriptions, request correlation, and capability enforcement
```

The tooling transport must not duplicate gameplay mutation logic.

## Current public-safe tooling inventory

These packets already belong to `shared/packets/tooling.toml`.

### Client to server

| Packet | Class | Current route | Final route | Capability | Attachment | Owner |
| --- | --- | --- | --- | --- | --- | --- |
| `telemetry_subscribe` | Subscription | `sr.tooling` router exists; production provider not wired | `sr.tooling` | Public-safe | Room | Tooling router + telemetry provider |
| `telemetry_unsubscribe` | Subscription | `sr.tooling` router exists | `sr.tooling` | Public-safe | Room | Tooling router |
| `telemetry_ping` | Request | Tooling route exists; legacy overlay still sends a WebSocket copy | `sr.tooling` only | Public-safe | Connection | Tooling router |
| `measurement_start` | Command | `sr.tooling` | `sr.tooling` | Public-safe | Room | Measurement controller |
| `measurement_stop` | Command | `sr.tooling` | `sr.tooling` | Public-safe | Room | Measurement controller |
| `measurement_reset` | Command | `sr.tooling` | `sr.tooling` | Public-safe | Room | Measurement controller |
| `measurement_snapshot_request` | Request | `sr.tooling` | `sr.tooling` | Public-safe | Room | Measurement controller |

### Server to client

| Packet | Class | Trigger | Capability context | Attachment | Owner |
| --- | --- | --- | --- | --- | --- |
| `telemetry_snapshot` | Server push | Active telemetry subscription | Public-safe subscription | Room | Telemetry provider |
| `telemetry_pong` | Response | `telemetry_ping` | Public-safe request | Connection | Tooling router |
| `measurement_started` | Response | `measurement_start` | Public-safe request | Room | Measurement controller |
| `measurement_snapshot` | Response | `measurement_snapshot_request` | Public-safe request | Room | Measurement controller |
| `measurement_stopped` | Response | `measurement_stop` or finalization flow | Public-safe request/run | Room | Measurement controller |
| `tooling_error` | Error response | Any rejected tooling request | Inherits originating request | Connection or room | Tooling router |

## Runtime developer command migration inventory

All commands below currently travel as JSON text over the gameplay WebSocket and are classified before normal `game.ClientPacket` decoding. Their final physical route is `sr.tooling`. Their packet type names remain unchanged.

Every command requires `tooling.control`, an attached room, and no gameplay participant slot.

| Packet | Scope | Authoritative owner | Required result |
| --- | --- | --- | --- |
| `toggle_debug_invincible` | Player target | Devtools controller -> game control | `tooling_command_result` or `tooling_error` |
| `toggle_debug_infinite_lives` | Player target | Devtools controller -> game control | `tooling_command_result` or `tooling_error` |
| `toggle_debug_freeze_world` | Room or freeze sub-target | Devtools controller -> game control | `tooling_command_result` or `tooling_error` |
| `toggle_debug_freeze_player` | Player target | Devtools controller -> game control | `tooling_command_result` or `tooling_error` |
| `debug_kill_player` | Player target or `all_players` | Devtools controller -> damage/lifecycle control | `tooling_command_result` or `tooling_error` |
| `debug_spawn_entity` | Room position, optional player target | Devtools controller -> spawn control | `tooling_command_result` or `tooling_error` |
| `debug_spawn_pickup` | Room position | Devtools controller -> pickup spawn control | `tooling_command_result` or `tooling_error` |
| `debug_begin_continuous_bullet_stream` | Room position and direction | Devtools controller -> stream runtime/spawn control | `tooling_command_result` or `tooling_error` |
| `debug_respawn_player` | Player target or `all_players` | Devtools controller -> respawn control | `tooling_command_result` or `tooling_error` |
| `debug_set_score` | Player target | Devtools controller -> counter control | `tooling_command_result` or `tooling_error` |
| `debug_add_score` | Player target | Devtools controller -> counter control | `tooling_command_result` or `tooling_error` |
| `debug_set_lives` | Player target | Devtools controller -> counter control | `tooling_command_result` or `tooling_error` |
| `debug_add_lives` | Player target | Devtools controller -> counter control | `tooling_command_result` or `tooling_error` |
| `debug_clear_bullets` | Room | Devtools controller -> entity control | `tooling_command_result` or `tooling_error` |
| `debug_clear_asteroids` | Room | Devtools controller -> entity control | `tooling_command_result` or `tooling_error` |

All migrated commands must carry a non-empty `request_id`. Existing fields such as `trace_id`, `target_scope`, and `target_player_id` remain valid during migration.

`target_player_id` remains a compatibility field for these player-only devtools commands. New non-devtools targeting continues to use `target_kind` plus `target_id`.

## Developer readout migration inventory

| Packet | Current behavior | Final behavior | Capability | Attachment |
| --- | --- | --- | --- | --- |
| `debug_status_subscribe` | Not implemented | Start change-driven debug-status delivery | `tooling.read` | Room |
| `debug_status_unsubscribe` | Not implemented | Stop debug-status delivery | `tooling.read` | Room |
| `debug_status` | Built by WebSocket write-loop path | Change-driven `sr.tooling` server push while subscribed | `tooling.read` | Room |
| `debug_shape_catalog_request` | Not implemented | One-shot request for current room catalog | `tooling.read` | Room |
| `debug_shape_catalog` | Pushed once per room by WebSocket write loop | One-shot `sr.tooling` response; client caches by catalog/version | `tooling.read` | Room |
| `tooling_command_result` | Not implemented | Terminal success response for every migrated mutation command | Inherits `tooling.control` request | Room |

The status stream is subscription-owned, not participation-owned. It may include room-global status and authorized player status maps without requiring the receiving session to control a player.

The shape catalog is static metadata for a room/runtime contract and should not be sent every write tick. The request/response route replaces connection-local "sent room id" tracking.

## Request and result contract

Every client-initiated tooling request, command, subscribe, and unsubscribe packet carries a non-empty `request_id`.

Request correlation rules:

```text
request_id is unique within the connection lifetime
responses echo request_id
one request receives at most one terminal success response
rejections and execution failures return tooling_error
server-pushed subscription data does not reuse the subscription request_id as an event id
unknown or late responses are ignored client-side and recorded diagnostically
```

Mutation completion rules:

```text
accepted is not the same as applied
success is emitted only after the owning seam reports application
no-op and invalid-target outcomes are explicit results or errors, never silent success
state projection packets remain authoritative for resulting state
```

`tooling_command_result` must identify the original command type and whether the command was applied. Optional command-specific result data may be carried in a bounded result map. It must not become a second gameplay-state protocol.

## Subscription contract

Subscriptions are connection-local and room-attachment-scoped.

```text
room attachment change clears room subscriptions
channel close clears subscriptions
WebRTC recovery does not imply authorization loss
successful channel recovery requires explicit resubscription by the client
session close finalizes active measurement and clears all tooling state
```

Telemetry and debug-status subscriptions are independent. Opening or closing a UI panel may subscribe or unsubscribe, but UI visibility is not server authority.

## Local-only devtools inventory

These concerns remain client-local and do not migrate to `sr.tooling`:

```text
devtools window visibility
telemetry overlay visibility and layout
hitbox overlay visibility
player label visibility
local target selector UI state
placement preview and pointer rendering
measurement panel presentation
local report export presentation
hotkey bindings
formatting and unavailable-value display
```

A local control that requests server state or mutation still sends its request through `sr.tooling`; only its presentation state remains local.

## Explicit exclusions

The following are not devtools transport packets and must not be moved as part of this migration:

```text
authenticate_request and authenticate_result
WebRTC offer, answer, ICE, ready, smoke, and failure signaling
room create/join/leave/ready/start requests and room responses
normal input packets
normal respawn and pause requests
normal target selection requests
resync_request and resync_required
lane-native gameplay state and event packets
```

Admin packet types do not exist yet. Future admin packets use `sr.tooling`, require `admin.control`, use explicit request/result correlation, and call audited admin-owned seams. They do not reuse runtime debug commands merely because the transport is shared.

## Migration slices

The remaining implementation order is fixed:

```text
1. request/response foundation
   -> request_id on debug commands
   -> tooling_command_result
   -> capability policy enforcement seam

2. runtime developer commands
   -> client sends commands through sr.tooling
   -> server tooling router dispatches to existing devtools controller
   -> remove WebSocket command classification after equivalence coverage

3. developer readouts
   -> debug-status subscribe/unsubscribe
   -> debug shape catalog request/response
   -> remove WebSocket write-loop debug output

4. continuous telemetry
   -> move legacy overlay ping fully onto sr.tooling
   -> wire production telemetry provider
   -> repair transport metrics against final per-lane sources

5. authorization and attachment
   -> replace the temporary all-connections tooling grant with an account-flag grant source
   -> observer/ghost/active-operator room attachment
   -> audit and tooling-affected propagation
```

Command migration and readout migration may be implemented in parallel only after the request/result and policy foundation is merged.

## Exit criteria

This contract phase is complete when:

```text
every current runtime devtools packet is inventoried
current and target physical routes are explicit
every packet has a capability and attachment requirement
local-only concerns and explicit exclusions are listed
request/result and subscription semantics are fixed
room attachment is not coupled to gameplay participation
a machine-readable server packet-policy registry covers the inventory
```

The overall migration is complete only when no runtime devtools command or readout uses the WebSocket gameplay packet path and tests prove unauthorized sessions cannot invoke privileged routes.

## Code map

```text
shared/packets/tooling.toml
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/networking/tooling/packet_contract.go
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/devtools/
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/inbound/tooling_packet_router.gd
client/scripts/devtools/
```

## Core invariants

```text
sr.tooling transport availability is not itself the authorization source; the temporary grant provider currently grants tooling.read and tooling.control to every connected session.
Runtime measurement and public-safe telemetry remain available without dev/admin clearance.
Privileged readouts require tooling.read.
Runtime developer mutations require tooling.control.
Future administrative actions require admin.control.
No tooling command requires gameplay participation merely to authorize the operator.
One tooling session attaches to at most one room.
Every privileged mutation uses the real owning gameplay/service seam.
Every migrated request has explicit correlation and a terminal result or error.
WebSocket remains the pre-readiness auth/signaling/lobby/session-control transport.
Gameplay lanes never carry tooling traffic.
```

## Related docs

- [Devtools Packet Protocol](./devtools-packet-protocol.md)
- [Devtools Authority And Seams](./devtools-authority-and-seams.md)
- [Devtools And Telemetry](../../planning/devtools/devtools-and-telemetry.md)
- [Realtime WebRTC Gameplay Transport](../../protocol/realtime-webrtc-gameplay-transport.md)
- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
