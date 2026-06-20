# Continuous Bullet Streams

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side devtools continuous bullet stream implementation.

It covers the debug command surface, stream runtime ownership, simulation-step integration, bullet spawning authority, build/runtime gates, current stop/clear behavior, telemetry, tests, and the relationship between this tool and normal gameplay firing.

## Overview

Continuous bullet streams are server devtools that repeatedly spawn debug bullets from a fixed world origin in a fixed direction.

They are created by the devtools command packet:

```text
debug_begin_continuous_bullet_stream
```

The command carries:

```text
x
y
has_direction
direction_x
direction_y
```

The server treats this as a debug-only request. The client chooses a placement origin and drag direction, but the server owns whether the stream begins, how the direction is normalized, when bullets are spawned, whether bullets can currently spawn, and how spawned bullets enter authoritative game state.

The current high-level path is:

```text
client devtools placement
-> debug_begin_continuous_bullet_stream packet
-> networking inbound devtools routing
-> devtools.HandleCommand
-> handleDebugBeginContinuousBulletStream
-> streamruntime.DefaultRuntime.BeginContinuousBulletStream
-> game.DevtoolsRegisterSimulationStepObserver
-> streamruntime.StepContinuousBulletStreams
-> game.DevtoolsSpawnDebugBullet
-> normal state packet bullet readback
```

A stream records:

```text
OwnerPlayerID
Origin
Direction
CooldownRemaining
```

The origin is normalized into wrapped world coordinates. The direction is normalized before storage. Each stream starts with `CooldownRemaining` set to `constants.BasicCannonCooldown`.

Stream ticking runs from the game simulation observer path. The observer is registered once per `*game.Game` when a valid stream command first reaches that game.

## Debug-only scope

Continuous bullet streams are development tooling.

They may:

* Spawn repeated debug bullets for local combat, collision, visibility, pacing, and rendering diagnostics.
* Use the normal authoritative game entity store so clients observe results through normal state sync.
* Use the same world bullet movement gate that freezes normal bullets.
* Reuse the game-owned debug bullet spawn adapter.

They must not:

* Become a player-facing weapon, loadout, or replay system.
* Give the client authority over bullet cadence.
* Apply client-only projectile mutations.
* Bypass the server `nodevtools` gate.
* Replace normal weapon fire policy.
* Move stream runtime state into normal `internal/game` gameplay ownership.

This tool injects debug projectiles. It does not simulate normal player input, consume weapon ammo, consume equipped weapon cooldowns, or resolve loadout legality.

## Server authority

The server owns stream creation and ticking.

`handleDebugBeginContinuousBulletStream` rejects incomplete stream requests by consuming the command without starting a stream when:

```text
has_direction == false
direction length == 0
streamruntime rejects the request
```

`streamruntime` rejects stream creation when:

```text
owner player ID is empty
normalized direction is zero
```

When creation succeeds, the server logs the normalized stream direction and registers a simulation step observer for the target game if one has not already been registered.

The current stream runtime owner is:

```text
services/game-server/internal/devtools/streamruntime
```

The game package does not import the devtools package and does not own stream state. Game-owned behavior is exposed through narrow devtools export seams:

```text
DevtoolsRegisterSimulationStepObserver
DevtoolsBulletsCanMove
DevtoolsSpawnDebugBullet
```

The stream observer runs at the end of `Game.Step(delta)`, after normal or reduced simulation phases. The observer still runs while `Game.Step` holds the game lock, so stream callbacks must stay small and route mutations through game-owned adapter functions.

## Stream ticking and bullet spawning

Each stream subtracts the simulation delta from `CooldownRemaining`.

When the cooldown reaches zero or below, the stream attempts to spawn a debug bullet only if:

```text
bulletsCanMove == true
```

The `bulletsCanMove` value comes from:

```text
game.DevtoolsBulletsCanMove()
```

That delegates to the world simulation bullet gate:

```text
WorldSimulationOptions.BulletsCanMove()
```

When bullets are frozen, stream cooldown still advances, but no projectile is spawned. A ready stream can spawn on a later observer tick after bullet movement is re-enabled.

When spawn succeeds, the stream cooldown resets to:

```text
constants.BasicCannonCooldown
```

Debug bullets are spawned through:

```text
game.DevtoolsSpawnDebugBullet(ownerPlayerID, origin, direction)
```

That delegates to `spawnDebugBullet`, which:

* rejects empty owner IDs
* rejects zero directions
* normalizes the spawn position into world bounds
* normalizes the direction
* uses `constants.BasicCannonProjectileSpeed`
* uses `constants.BasicCannonProjectileLifetime`
* allocates a bullet ID through the game spawner
* inserts the bullet into `game.entities.Projectiles`

Clients then see the projectile through normal state packet projection and world sync. There is no separate stream-specific outbound packet.

## Commands or controls

The server command is:

```text
debug_begin_continuous_bullet_stream
```

The current packet payload is:

```text
{
  "type": "debug_begin_continuous_bullet_stream",
  "x": <world x>,
  "y": <world y>,
  "has_direction": true,
  "direction_x": <direction x>,
  "direction_y": <direction y>
}
```

The requesting server game player becomes the stream owner. The current command does not use `target_player_id`.

Client-side placement currently sends this command from continuous bullet stream placement. The client requires a drag direction before sending the packet. If placement has no direction, the client packet builder returns an empty packet and no command is sent.

Server inbound routing treats this as a remaining devtools packet type:

```text
HandleRemainingDevtoolsPacket
-> handleDevtoolsCommandPacket
-> devtools.HandleCommand
```

The command does not route through `Game.HandlePacket`.

### Stop and clear behavior

There is no dedicated stop-stream packet in the current command surface.

The stream runtime package exposes:

```text
ClearContinuousBulletStreams()
```

but the current inspected command path does not call it.

`debug_clear_bullets` currently routes to:

```text
handleDebugClearBullets
-> Game.DevtoolsClearBullets
```

That clears existing projectile entities from the authoritative game entity store. It does not currently clear active continuous stream records in `streamruntime`.

Do not document `debug_clear_bullets` as the authoritative stream stop path unless the command handler is changed to call the stream runtime clear seam.

## Client presentation

The server does not send stream-specific presentation state.

The client may request continuous stream placement, but after the command is sent, stream effects are visible only through ordinary server state readback:

```text
StatePacket.bullets
world sync bullet rendering
normal projectile movement and expiration
```

There is no current server-provided active-stream count, stream list, stream owner readout, stream cooldown readout, or stream indicator in `debug_status`.

Client-side documents own placement input, drag direction collection, and devtools window/hotkey presentation. This server doc owns command handling, server runtime state, simulation integration, and authoritative bullet spawning.

## Telemetry

Continuous bullet streams currently expose telemetry through logs and normal state readback.

Server logs include ignored begin requests for:

```text
has_direction is false
direction is zero
generic streamruntime rejection
```

Successful stream creation logs:

```text
debug continuous bullet stream started
```

with:

```text
player_id
x
y
direction_x
direction_y
```

Spawned bullets are not emitted as a stream-specific telemetry packet. They appear as normal bullets in gameplay state packets.

## Build/runtime gates

Server devtools are enabled in default builds through:

```text
services/game-server/internal/devtools/enabled_default.go
```

`nodevtools` builds disable devtools command handling through:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

The devtools command gate is:

```text
ShouldHandleCommand(packetType)
= IsCommandType(packetType) && Enabled()
```

Runtime command routing also requires:

```text
current room exists
current game player ID is non-empty
packet decodes into DebugCommand
command type is handled by devtools.HandleCommand
```

Stream-specific runtime gates require:

```text
has_direction == true
direction is nonzero
owner player ID is non-empty
bullets can move before each stream spawn
debug bullet spawn succeeds before cooldown reset
```

Client-side build flags can remove DevToggle inputs in public builds, but client-side gates are not the server authority boundary. Server devtools availability remains controlled by the server build/runtime gate.

## Relationship to real gameplay implementation areas

Continuous bullet streams deliberately reuse server-owned implementation seams.

They reuse:

```text
Game.Step observer cadence
WorldSimulationOptions.BulletsCanMove
space.NormalizePosition
runtime.Bullet
game.entities.Projectiles
state packet bullet projection
normal client world sync rendering
```

They do not reuse:

```text
normal player input fire consumption
weapon slot selection
weapon ammo consumption
equipped weapon cooldowns
loadout legality
client-side projectile spawning
```

The debug stream path should remain a devtools-only projectile injection tool. Normal gameplay firing belongs to the player weapon simulation and weapon policy seams. Future weapon or loadout work should not be implemented by expanding continuous bullet streams.

The observer seam is also not a general scheduling system. Current usage is narrow: server-paced devtools stream spawning that needs to run inside the authoritative simulation cadence.

## Code map

Primary server implementation:

```text
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/streamruntime/runtime.go
services/game-server/internal/devtools/streamruntime/simulation.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams.go
```

Command routing and build gates:

```text
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/packets_generated.go
```

Game-owned devtools adapters:

```text
services/game-server/internal/game/export_devtools_streams.go
services/game-server/internal/game/export_devtools_spawn.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/world_simulation_options.go
```

Packet source and generated outputs:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Related client request construction:

```text
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd
client/scripts/gameplay/devtools/debug_click_placement_flow.gd
client/scripts/devtools/context/devtools_placement_context.gd
client/scripts/devtools/dev_connection_service.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/weapons/
```

owns normal weapon policy and player firing.

```text
client/scripts/devtools/
```

owns client-side debug presentation and command request construction.

```text
services/game-server/internal/networking/
```

owns WebSocket packet routing and session context.

```text
services/game-server/internal/devtools/streamruntime/
```

owns stream runtime state outside the normal game package.

## Tests

Focused stream runtime tests:

```text
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
```

These verify:

* valid stream creation
* owner tracking
* origin storage
* direction normalization
* initial cooldown
* invalid owner rejection
* zero-direction rejection
* runtime clear behavior
* spawn after cooldown
* cooldown reset after successful spawn

Related server devtools tests:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/clear_entities_test.go
```

These verify command classification, build gates, and related clear-entity command behavior.

Related game adapter and integration tests:

```text
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/tests/game/continuous_bullet_stream_test.go
```

These verify:

* new games report that bullets can move
* game-owned debug bullet spawning works for valid owners
* stream runtime can spawn a bullet through the game-owned devtools bullet adapter after cooldown

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/... ./internal/game ./tests/game
```

Nodevtools gate verification:

```bash
cd services/game-server
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)

## Notes

Legacy server devtools notes correctly identified that continuous bullet streams are server-paced and that the client does not own cadence. The current implementation confirms that stream ticking is driven by the game simulation observer path and that bullet spawning remains server-authoritative.

Legacy notes also claimed that clear bullets clears active persistent streams. The current inspected implementation does not do that. Current docs should describe the live implementation: clear bullets removes existing projectile entities, while stream runtime clearing exists as a package method but is not currently wired to the clear-bullets command.

The current `DefaultRuntime` is package-level stream runtime state. Stream records carry owner player ID, origin, direction, and cooldown, but not a room ID or game ID. If concurrent multi-room stream behavior becomes important, stream ownership should be revisited before treating streams as room-scoped tooling.
