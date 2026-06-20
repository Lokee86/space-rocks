# Clear Entity Tools

Parent index: [Server](./!README.md)

## Purpose

This document describes the server-side devtools clear-entity tools.

Clear-entity tools are debug commands that remove selected classes of live server-owned entities from the current room game. They exist to reset noisy debug state during active development without giving the client authority over gameplay state.

## Overview

Clear-entity tools are client-triggered and server-authoritative.

The current clear tools are:

```text
debug_clear_bullets
debug_clear_asteroids
```

Both commands are room/global commands. They do not target a player, do not use placement coordinates, and do not resolve through the canonical gameplay target. The requesting client sends a generated debug packet, networking routes it through the devtools command path, and the server mutates the authoritative `Game` entity store through narrow game-owned export seams.

The client does not remove bullets or asteroids locally. Removed entities disappear from presentation because the next server state packet no longer includes them.

## Debug-only scope

Clear-entity tools are devtools controls, not production gameplay mechanics.

They may remove:

```text
debug_clear_bullets   -> all live projectiles in the current Game entity store
debug_clear_asteroids -> all live asteroids in the current Game entity store
```

They do not remove:

```text
players
player sessions
camera views
pickups
enemies
score
lives
match state
target state
pending presentation events
```

The clear handlers are intentionally direct. They delete current live entities from the authoritative maps rather than simulating collisions, despawn timers, scoring, drops, or death consequences.

## Server authority

The authoritative mutation is owned by `internal/game`.

`internal/devtools` receives and dispatches the command, but it does not directly own the entity maps. The actual mutations happen through:

```text
Game.DevtoolsClearBullets()
Game.DevtoolsClearAsteroids()
```

Those methods lock the `Game`, count the current entities, delete every entry in the corresponding entity map, and return the removed count.

Current map ownership:

```text
bullets   -> game.entities.Projectiles
asteroids -> game.entities.Asteroids
```

The returned count is currently an internal helper result. The clear command handlers do not send an acknowledgement packet with the count.

Clear-asteroids does not award score, split asteroids, spawn drops, or emit asteroid-destruction presentation events. Existing test coverage verifies that clearing asteroids removes asteroids while preserving the player session score.

Clear-bullets removes live projectiles from the game entity store. Continuous bullet stream runtime state is owned separately under `services/game-server/internal/devtools/streamruntime/`; the inspected clear-entity handler does not call `ClearContinuousBulletStreams()`.

## Client presentation

The client presentation role is request-only.

The devtools window and command context expose clear controls. Pressing a clear control sends one of the generated packets:

```gdscript
Packets.debug_clear_bullets_packet()
Packets.debug_clear_asteroids_packet()
```

The client logs that the request was sent, but it does not apply the mutation locally. Bullet and asteroid nodes are removed through normal world sync after the server projects a state packet without those entities.

This keeps clear tools aligned with the normal server-authoritative presentation model:

```text
client button
-> generated debug packet
-> websocket send path
-> server devtools command route
-> game-owned entity store mutation
-> next authoritative state packet
-> client world sync removes missing entities
```

## Commands and controls

| Command                 | Generated client builder         | Server handler              | Authoritative mutation          |
| ----------------------- | -------------------------------- | --------------------------- | ------------------------------- |
| `debug_clear_bullets`   | `debug_clear_bullets_packet()`   | `handleDebugClearBullets`   | `Game.DevtoolsClearBullets()`   |
| `debug_clear_asteroids` | `debug_clear_asteroids_packet()` | `handleDebugClearAsteroids` | `Game.DevtoolsClearAsteroids()` |

Packet bodies:

```json
{
  "type": "debug_clear_bullets"
}
```

```json
{
  "type": "debug_clear_asteroids"
}
```

The generated `DebugCommand` type contains shared fields used by other devtools commands, but clear commands currently use only `type`.

## Routing behavior

Clear commands are classified as simple devtools packets by the inbound networking route.

The route is:

```text
networking.handleClientPacket
-> inbound.RouteClientPacket
-> inbound.HandleSimpleDevtoolsPacket
-> inbound.handleDevtoolsCommandPacket
-> devtools.HandleCommand
-> handleDebugClearBullets / handleDebugClearAsteroids
-> Game.DevtoolsClearBullets / Game.DevtoolsClearAsteroids
```

If the websocket session has no current room or no current game player ID, the devtools packet is consumed and no mutation is applied.

If the command cannot decode into `devtools.DebugCommand`, networking logs a devtools command decode warning and consumes the packet.

If `devtools.HandleCommand` receives a nil game target, the clear handler returns `false` and performs no mutation.

## Telemetry

Clear-entity tools do not currently emit a dedicated acknowledgement packet, removal-count packet, or clear-specific presentation event.

Observable effects are indirect:

```text
StatePacket.bullets   -> empty or reduced after bullet clear
StatePacket.asteroids -> empty or reduced after asteroid clear
```

Client-side devtools controls log that the clear request was sent. Server-side clear handlers do not currently log the removed count.

Debug status output is separate. It reports devtools toggle state such as invincibility, infinite lives, freeze state, and per-player debug status. It is not the acknowledgement channel for clear-entity commands.

## Build and runtime gates

The devtools package has build-tag gates:

```text
default build  -> devtools.Enabled() == true
nodevtools     -> devtools.Enabled() == false
```

`devtools.ShouldHandleCommand(packetType)` combines command-type recognition with the build gate and is covered by devtools tests.

Clear-entity packet classification and dispatch are part of the general inbound devtools command routing. Command-routing documentation owns the cross-command gate policy; this document only covers the clear-entity command behavior itself.

## Code map

Primary implementation:

```text
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/game/export_devtools_clear_entities.go
```

Command dispatch and packet classification:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
```

Entity storage:

```text
services/game-server/internal/game/runtime/state.go
```

Packet source and generated outputs:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
```

Client request path:

```text
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_window.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
```

Related runtime state outside the clear handler:

```text
services/game-server/internal/devtools/streamruntime/
```

Important non-ownership boundaries:

```text
internal/devtools does not own the Game entity maps.
internal/game does not import internal/devtools for clear behavior.
client devtools do not locally remove gameplay entities.
clear commands do not run combat, scoring, drop, despawn, or presentation-event logic.
```

## Tests and verification

Focused tests:

```text
services/game-server/internal/devtools/clear_entities_test.go
```

Current coverage verifies:

```text
debug_clear_bullets removes all bullets
debug_clear_bullets is safe when no bullets exist
debug_clear_asteroids removes all asteroids
debug_clear_asteroids is safe when no asteroids exist
debug_clear_asteroids preserves player session score
```

Supporting tests:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
```

Verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/...
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

## Related docs

* [Server Devtools](./!README.md)
* [Devtools](../!README.md)
* [Game Server](../../services/game-server/!README.md)
* [Game Server Runtime](../../services/game-server/simulation/runtime/!README.md)
* [World Simulation](../../services/game-server/simulation/world/!README.md)
* [Realtime Protocol](../../protocol/realtime/!README.md)
* [Data](../../data/!README.md)

## Notes

Legacy documentation described clear bullets as also clearing active persistent debug bullet streams. The current inspected clear-entity handler removes live projectiles from the `Game` entity store and does not call the separate stream runtime clear method.
