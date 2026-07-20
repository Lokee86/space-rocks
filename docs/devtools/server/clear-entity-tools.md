---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-75ce-880d-63e9b3310aab
document_type: general
policy_exempt: false
summary: This document describes the server-side devtools clear-entity tools.
---
# Clear Entity Tools

Parent index: [Server](./!INDEX.md)

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

Both commands are room/global commands. They do not target a player, do not use placement coordinates, and do not resolve through the canonical gameplay target. `GameplayDebugFlow` or `DevConnectionService` builds the command payload with `request_id` and `trace_id`; `ClientConnectionService.send_tooling_packet()` sends it through `sr.tooling`; and server networking/tooling capability-checks and dispatches it to the existing devtools controller before the authoritative `Game` entity store is mutated through `Control.ClearBullets` and `Control.ClearAsteroids`.

Debug/devtools entity creation and clearing are reflected to clients through normal realtime entity-family lanes: asteroid/bullet lifecycle lanes for existence and hot lanes for movement.

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
Control.ClearBullets()
Control.ClearAsteroids()
```

Those methods lock the authoritative entity maps, count the current entities, delete every entry in the corresponding entity map, and return the removed count.

Current map ownership:

```text
bullets   -> game.entities.Projectiles
asteroids -> game.entities.Asteroids
```

The returned count is currently an internal helper result. Successful clear commands return a correlated `tooling_command_result` without the count; rejected or unapplied commands return a correlated `tooling_error`.

Asteroid clears produce asteroid lifecycle deletes on the authoritative server. The next client hot movement updates cannot recreate deleted asteroids.

Bullet clears produce bullet lifecycle deletes on the authoritative server. The next client hot movement updates cannot recreate deleted bullets.

## Client presentation

The client presentation role is request-only.

The devtools window and command context expose clear controls. Pressing a clear control sends one of the generated packets:

```gdscript
Packets.debug_clear_bullets_packet()
Packets.debug_clear_asteroids_packet()
```

Cleared bullets/projectiles disappear through bullets_lifecycle deletes or full-state correction, cleared asteroids disappear through asteroids_lifecycle deletes or full-state correction, and cleared pickups disappear through world/pickup state readback.

This keeps clear tools aligned with the normal server-authoritative presentation model:

```text
client button
-> GameplayDebugFlow or DevConnectionService builds request_id/trace_id command payload
-> ClientConnectionService.send_tooling_packet()
-> sr.tooling
-> server networking/tooling capability check and devtools controller dispatch
-> game-owned entity store mutation
-> tooling_command_result or tooling_error
-> ToolingPacketRouter
-> next authoritative family-specific realtime readback
-> client world sync removes missing entities
```

## Commands and controls

| Command                 | Generated client builder         | Server handler              | Authoritative mutation          |
| ----------------------- | -------------------------------- | --------------------------- | ------------------------------- |
| `debug_clear_bullets`   | `debug_clear_bullets_packet()`   | `handleDebugClearBullets`   | `Control.ClearBullets()`   |
| `debug_clear_asteroids` | `debug_clear_asteroids_packet()` | `handleDebugClearAsteroids` | `Control.ClearAsteroids()` |

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

The route is:

```text
GameplayDebugFlow or DevConnectionService
-> ClientConnectionService.send_tooling_packet()
-> sr.tooling
-> networking/tooling preflight
-> Controller.HandleCommand
-> handleDebugClearBullets / handleDebugClearAsteroids
-> Control.ClearBullets / Control.ClearAsteroids
```

The tooling route requires the room attachment and `tooling.control` capability. A current game player ID is not required merely to dispatch a room/global clear command.

If the command cannot decode into `devtools.DebugCommand`, the tooling route returns `tooling_error` and does not mutate state.

If `Controller.HandleCommand` receives a nil game target, the clear handler returns `false` and performs no mutation.

## Telemetry

Clear-entity tools do not currently emit a dedicated acknowledgement packet, removal-count packet, or clear-specific presentation event.

Observable effects are indirect:

```text
bullets_lifecycle deletes   -> empty or reduced after bullet clear
asteroids_lifecycle deletes -> empty or reduced after asteroid clear
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

Clear-entity packet policy and dispatch are part of the general `sr.tooling` command route. Command-routing documentation owns the cross-command room and capability gates; this document only covers the clear-entity command behavior itself.

## Code map

Primary implementation:

```text
services/game-server/internal/devtools/controller.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/game/control_clear.go
```

Command dispatch and packet classification:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/preflight.go
services/game-server/internal/networking/tooling/commands.go
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
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_window.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/inbound/tooling_packet_router.gd
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

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game Server Runtime](../../services/game-server/simulation/runtime/!INDEX.md)
* [World Simulation](../../services/game-server/simulation/world/!INDEX.md)
* [Realtime Protocol](../../protocol/realtime-websocket-protocol.md)
* [Data](../../data/!INDEX.md)

## Notes

`debug_clear_bullets` clears live projectile entities and clears continuous streams in the Controller-selected runtime.
