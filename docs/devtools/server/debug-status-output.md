---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7b7a-bdc5-8e5d06c063e3
document_type: general
policy_exempt: false
summary: 'This document describes the game-server debugstatus readout: its authoritative source, sr.tooling subscription lifecycle, eligibility gates, packet shape, and client presentation boundary.'
---
## Debug Status Output

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the game-server `debug_status` readout: its authoritative source, `sr.tooling` subscription lifecycle, eligibility gates, packet shape, and client presentation boundary.

## Current route

`debug_status` is a privileged developer readout delivered exclusively through the reliable ordered `sr.tooling` WebRTC DataChannel.

```text
client enters an active room and sr.tooling becomes ready
-> debug_status_subscribe(request_id)
-> tooling packet policy preflight
-> tooling.read capability and room attachment checks
-> connection-local subscription state
-> tooling router Tick
-> outbound.BuildDebugStatusPacket(...)
-> debug_status over sr.tooling
-> ToolingPacketRouter
-> ClientConnectionService.debug_status_received
-> existing gameplay/devtools presentation owners
```

The legacy WebSocket write-loop route is removed. `ServerPacketRouter`, `ServerPacketDispatcher`, and `ClientInboundCoordinator` do not classify or forward `debug_status`.

The client automatically subscribes once for each active room after both conditions are true:

```text
sr.tooling is ready
room state is in_game or game_over
```

Returning to a non-game room state clears the client’s per-room request marker. Replacing the realtime transport also clears the marker, so a recovered `sr.tooling` channel re-establishes the subscription for the still-active room. Entering a later active room sends a new subscription request.

## Subscription lifecycle

The request packets are defined in `shared/packets/tooling.toml`:

```text
debug_status_subscribe
debug_status_unsubscribe
```

Both carry a non-empty `request_id`, require an attached room, and require `tooling.read`.

Subscription state belongs to the connection-local tooling router. The first router tick after subscribing is eligible to emit a status packet. Later packets are sampled every eight tooling-router ticks, preserving the former bounded debug-status cadence without using the WebSocket writer.

The emitted `debug_status.request_id` echoes the active subscription request ID as stable stream correlation. It is not a unique event ID.

Unsubscribe clears the connection-local status subscription immediately. Closing the tooling router also discards the subscription.

## Server authority

The server owns every status value.

```text
outbound.BuildDebugStatusPacket(room, playerID, requestID)
-> game.NewControl(room game instance)
-> devtools.NewController(...)
-> controller.StatusFor(playerID)
-> controller.StatusesForAllPlayers()
-> devtools.DebugStatusPacket
```

The devtools package owns the debug-facing projection. The game package owns the underlying simulation and player/session state.

Current `DebugStatus` fields are:

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

`debug_status` is the receiving gameplay identity’s status when one exists. `debug_statuses` is the authoritative per-player map used by target selectors and status labels.

Room-global readout delivery does not require the tooling session to have a `GamePlayerID`. An observer-capable session may subscribe when attached to the room and authorized. An empty receiving player ID produces the room/global projection plus the available per-player map.

## Eligibility

`CanSendDebugStatus` requires:

```text
room is not nil
room has a game instance
devtools.Enabled() is true
room state is InGame or GameOver
```

The tooling preflight additionally requires:

```text
client-to-server packet direction
attached room
tooling.read capability
valid request_id
valid packet payload
```

When the room is temporarily ineligible, the router skips that sampled push and keeps the subscription. `GameOver` remains eligible so the final debug-control state can still be displayed.

`nodevtools` builds make status output unavailable through `devtools.Enabled() == false`.

## Packet shape

The readout packet remains defined in `shared/packets/debug.toml` and is generated into `internal/devtools`:

```go
type DebugStatusPacket struct {
    Type          string                 `json:"type"`
    RequestID     string                 `json:"request_id"`
    DebugStatus   DebugStatus            `json:"debug_status"`
    DebugStatuses map[string]DebugStatus `json:"debug_statuses"`
}
```

The tooling request structs and packet type constants are generated from `shared/packets/tooling.toml` into:

```text
services/game-server/internal/protocol/tooling/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

The outbound presentation owner returns a typed packet. Tooling transport serialization is owned by the router/sender boundary; the presentation builder no longer encodes a WebSocket byte response.

## Client presentation boundary

The transport migration preserves the existing public client signals and application owners:

```text
ToolingPacketRouter.debug_status_received
-> ClientConnectionService._on_debug_status_received
-> ClientConnectionService.debug_status_received
-> SessionNetworkController
-> GameplaySessionController
-> GameplayComposition
-> GameplayDevtoolsContext
-> debug status readmodel and window refresh
```

The client consumes the packet as diagnostic presentation data. It does not apply gameplay mutations locally.

`DebugStatusPacketReader` treats malformed `debug_status` or `debug_statuses` values as empty dictionaries. Missing player rows degrade to inactive/empty presentation rather than becoming alternate authority.

## Relationship to commands

Debug commands and status readouts are separate protocol interactions.

Commands travel through `sr.tooling`, require `tooling.control`, and return `tooling_command_result` or `tooling_error`. A later `debug_status` push reports the resulting authoritative state through the independent `tooling.read` subscription.

The status packet is not a command acknowledgement and must not become a second gameplay-state protocol.

## Code map

```text
shared/packets/tooling.toml
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/packet_contract.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/packets_generated.go
client/scripts/networking/inbound/tooling_packet_router.gd
client/scripts/networking/client_connection_service.gd
client/scripts/devtools/debug_status_packet_reader.gd
```

Authoritative gameplay sources remain under:

```text
services/game-server/internal/game/control_status.go
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
```

## Verification

Relevant coverage includes:

```text
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/networking/tooling/router_test.go
services/game-server/internal/devtools/controller_status_test.go
client/tests/unit/networking/test_tooling_packet_router.gd
client/tests/unit/test_client_connection_service.gd
client/tests/unit/devtools/debug_status_packet_reader_test.gd
```

Tests verify typed packet construction, request correlation, immediate subscribed delivery, unsubscribe behavior, tooling-router client dispatch, preservation of the existing client signal, eligibility guards, and status projection.

## Related docs

* [Server Devtools](./!INDEX.md)
* [Tooling Channel Migration Contract](../design/tooling-channel-migration-contract.md)
* [Devtools Packet Protocol](../design/devtools-packet-protocol.md)
* [Client Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
