## Toggles

Parent index: [Server](./!README.md)

## Purpose

This document describes the server-side devtools toggle system for gameplay-affecting debug controls.

It covers command packets, target resolution, server-owned mutation, debug status output, runtime/build gates, and the implementation seams that keep debug behavior routed through normal authoritative game systems.

## Overview

Server devtools toggles are debug-only command handlers that mutate authoritative game-server state for local development and diagnostics.

Current server-side toggles are:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
```

The server receives these commands as generated debug packets, decodes them into `devtools.DebugCommand`, routes them through `devtools.HandleCommand`, and applies the result through narrow game-owned export seams.

The client may request a toggle from a hotkey or devtools window control, but the client does not apply toggle effects locally. Confirmation comes back through normal gameplay state, debug status packets, visible entity behavior, or both.

The toggle system has three kinds of server-owned state:

```text
player damage options
-> invincibility

player session life options
-> infinite lives

world simulation options
-> aggregate and granular world freeze flags

player suspension state
-> player freeze
```

The toggle system does not create a parallel simulation. It changes existing runtime options that normal simulation, collision, life, input, movement, and firing code already consult.

## Debug-only scope

Server toggles are development and diagnostic controls.

They may:

```text
prevent selected players from taking collision damage
prevent selected sessions from losing lives
freeze all or part of world simulation behavior
freeze selected player sessions
report current toggle state through debug status packets
log toggle changes through the server logger
```

They must not:

```text
be treated as player-facing gameplay features
persist to player profile data
move authority to the client
bypass the normal game aggregate
duplicate damage, lives, suspension, or simulation rules in devtools-only code
be used as a production balancing or match-rule system
```

Debug toggles are intentionally scoped to active server runtime state. They are not stored in the player-data service, the API server, or any durable account/profile record.

## Server authority

The server owns whether a toggle is accepted and how it affects gameplay.

Current command flow:

```text
client debug packet
-> game-server networking inbound devtools route
-> packetcodec.Decode into devtools.DebugCommand
-> devtools.HandleCommand
-> toggle-specific handler in internal/devtools/toggles.go
-> game-owned export seam in internal/game/export_devtools_toggles.go
-> normal game runtime state
-> normal simulation/state/debug-status output
```

If the current WebSocket session has no room or no current game player ID, the inbound devtools route consumes the packet without applying a command. If packet decode fails, the command is not applied and a network warning is logged.

Toggle mutation happens behind the game aggregate:

```text
services/game-server/internal/game/export_devtools_toggles.go
```

That file exposes narrow devtools-facing methods. The devtools package does not reach directly into unrelated game internals.

The server keeps authoritative ownership split as follows:

| Toggle         | Authoritative state            | Owning runtime seam                |
| -------------- | ------------------------------ | ---------------------------------- |
| Invincible     | `DamageOptions.Invincible`     | player session and live ship state |
| Infinite Lives | `LifeOptions.InfiniteLives`    | player session state               |
| World Freeze   | `WorldSimulationOptions` flags | game aggregate simulation options  |
| Player Freeze  | `Suspension.DevFrozen`         | player session suspension state    |

## Commands or controls

### Invincibility

Packet type:

```text
toggle_debug_invincible
```

Invincibility prevents the selected player from taking normal collision damage through the server collision/damage capability path.

Current behavior:

```text
target_player_id omitted
-> target requesting player

target_player_id present
-> target that player

target_scope = all_players
-> target all current devtools target player IDs
```

The handler resolves target players, toggles or sets `DamageOptions.Invincible`, and logs the result as:

```text
debug invincibility set
```

Invincibility state is written to both places where it may matter:

```text
player session DamageOptions
live player entity DamageOptions
```

This keeps status and live collision behavior aligned. Collision damage checks use the normal `playerCanTakeCollisionDamage` path, which also respects suspension and temporary player invulnerability.

Invincibility does not freeze movement, shooting, scoring, or session timers.

### Infinite lives

Packet type:

```text
toggle_debug_infinite_lives
```

Infinite lives lets the selected player die normally without decrementing the session lives counter.

Current behavior:

```text
target_player_id omitted
-> target requesting player

target_player_id present
-> target that player

target_scope = all_players
-> target all current devtools target player IDs
```

The handler resolves target players, toggles or sets `LifeOptions.InfiniteLives`, and logs the result as:

```text
debug infinite lives set
```

Infinite lives is session-owned. A player can still take fatal damage, despawn, emit death/damage events, and use the respawn flow. The difference is that the life counter does not decrement while `LifeOptions.CanLoseLives()` is false.

Because the flag is session-owned, it persists across respawns for the same player session.

### World freeze

Packet type:

```text
toggle_debug_freeze_world
```

World freeze controls server simulation gates through `WorldSimulationOptions`.

The command accepts an optional `freeze_target` field.

Current freeze targets are:

```text
empty
all
asteroids
bullets
spawning
spawns
collisions
```

An empty `freeze_target` is treated as `all`.

Aggregate world freeze toggles all world freeze flags together:

```text
all or empty
-> FreezeAsteroids
-> FreezeBullets
-> FreezeSpawning
-> FreezeCollisions
```

If all flags are already frozen, the aggregate toggle unfreezes all of them. If the world is only partially frozen, the aggregate toggle freezes all flags.

Granular targets toggle only their own flag:

```text
asteroids
-> asteroid movement gate

bullets
-> bullet movement, bullet lifetime advancement, and player bullet firing gate

spawning or spawns
-> asteroid spawn timer and asteroid spawn creation gate

collisions
-> ship/asteroid, bullet/asteroid, and player/pickup collision passes
```

Unknown freeze targets are ignored without changing freeze flags. The server logs:

```text
debug world freeze target ignored
```

Valid world-freeze changes log:

```text
debug world freeze toggled
```

World freeze is not a universal pause. It does not stop every simulation phase. Player session timers, player movement, weapon cooldown stepping, pickup stepping, radial effects, observer callbacks, and cleanup paths continue unless they are specifically behind one of the frozen world gates.

### Player freeze

Packet type:

```text
toggle_debug_freeze_player
```

Player freeze suspends selected player sessions through the same suspension capability model used by pause, but it uses a separate suspension cause.

Current behavior:

```text
target_player_id omitted
-> target requesting player

target_player_id present
-> target that player

target_scope = all_players
-> target all current devtools target player IDs
```

The handler resolves target players, toggles or sets `Suspension.DevFrozen`, and logs the result as:

```text
debug player freeze set
```

When enabling player freeze, the server clears the live player input if the player entity exists.

Player freeze contributes to:

```text
Suspension.IsSuspended()
```

That means the normal gameplay capability helpers block:

```text
receiving input
movement
shooting
collision damage
score receipt
```

Pause and dev freeze are separate suspension causes:

```text
Paused || DevFrozen
-> suspended
```

Clearing pause does not clear dev freeze. Clearing dev freeze does not clear pause.

## Target resolution

Player-targeted toggles share the same server target resolution helper.

Single-player resolution:

```text
target_player_id present
-> use target_player_id

target_player_id absent
-> use requesting player ID
```

All-player resolution:

```text
target_scope = all_players
-> Game.DevtoolsTargetPlayerIDs()
```

`DevtoolsTargetPlayerIDs()` returns the sorted union of player IDs from:

```text
player sessions
live player entities
```

This lets all-player controls reach players that have a session but no current ship entity, as well as live player entities.

All-player toggles use set-style behavior:

```text
if any resolved target is missing or inactive for that feature
-> set every resolved target enabled

if every resolved target is already enabled
-> set every resolved target disabled
```

This avoids a mixed all-player selection flipping enabled players off while enabling disabled players.

`target_scope = all_players` is a scope value. It is not serialized or stored as a fake player ID.

## Client presentation

The server does not own devtools window layout, hotkey routing, target selectors, or overlay rendering.

The server does own the command effects and status data consumed by the client. Current client-facing surfaces include:

```text
debug_status.debug_status
debug_status.debug_statuses
normal gameplay state packet changes
normal entity/session lifecycle changes
logs visible from the server process
```

The client devtools window and hotkeys may request:

```text
invincibility
infinite lives
world freeze
asteroid freeze
bullet freeze
spawn freeze
collision freeze
player freeze
```

Those controls send generated debug packets through the normal client networking path. The server applies the command, then later status/state output lets the client update labels, rows, and diagnostic readouts.

The client must not locally infer that a command succeeded just because a button was pressed. Server state remains the authority.

## Telemetry behavior

Toggle telemetry is emitted through server debug status packets.

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

The server builds two status views:

```text
debug_status
-> status for the receiving/current player

debug_statuses
-> map of player ID to status for all match players
```

`debug_status` is used for receiver-local and room/global status labels. `debug_statuses` supports per-player target rows in the devtools window.

Status projection flow:

```text
game.DevtoolsStatusFor(playerID)
-> devtools.StatusFor(game, playerID)
-> devtools.DebugStatusPacket
-> packetcodec.Encode
-> WebSocket write loop
```

Debug status output is sent only when:

```text
room exists
game instance exists
server devtools are enabled
room state is InGame or GameOver
session has a current game player ID
```

Debug status packets are written on a slower cadence than gameplay state. Current code sends gameplay presentation state every server write tick, then sends debug status every `debugStatusWriteIntervalTicks`, currently `8`.

## Build/runtime gates

Server devtools have a build-tag gate:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
```

Default builds return:

```text
devtools.Enabled() == true
```

`nodevtools` builds return:

```text
devtools.Enabled() == false
```

The devtools package also exposes:

```text
ShouldHandleCommand(packetType string)
```

That helper returns true only when the packet type is a devtools command type and devtools are enabled.

Outbound debug status and debug shape catalog eligibility directly check `devtools.Enabled()` before sending.

Inbound command handling has additional runtime gates:

```text
current room must exist
current game player ID must be non-empty
debug packet must decode into DebugCommand
HandleCommand must recognize the command type
target/game export seams must accept the requested mutation
```

Command handlers return `true` for recognized devtools command packets even when a specific target has no effect. That keeps command routing separated from gameplay outcome.

## Data ownership

Toggle packet schemas are source-of-truth data under:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
```

Generated server output lives in:

```text
services/game-server/internal/devtools/packets_generated.go
```

Generated client packet builders live in:

```text
client/scripts/generated/networking/packets/packets.gd
```

Runtime toggle state is not generated data and is not persisted data.

Runtime state ownership:

```text
DamageOptions.Invincible
-> player session and live player entity

LifeOptions.InfiniteLives
-> player session

WorldSimulationOptions
-> game aggregate

Suspension.DevFrozen
-> player session
```

No toggle state belongs to player-data persistence, Rails API storage, static data, or client-only presentation state.

## Invariants

Server toggles must preserve these rules:

```text
server owns gameplay-affecting toggle state
client sends requests only
debug toggles use normal game-owned runtime seams
devtools code does not duplicate collision, damage, lives, movement, firing, or pause logic
target_scope=all_players is not a player ID
single-player target fallback uses the requesting player
all-player toggles enable all unless every target is already enabled
world freeze is a set of granular simulation gates, not a universal pause
player freeze is separate from pause but participates in shared suspension checks
toggle status is diagnostic telemetry, not persistent player data
```

## Code map

Primary server devtools toggle files:

```text
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/packets_generated.go
```

Build and command gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Networking ingress and debug status output:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
```

Game-owned export seams:

```text
services/game-server/internal/game/export_devtools_toggles.go
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/game/export_devtools_player_spawn.go
```

Runtime option and capability files:

```text
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/pause.go
```

Simulation files that consume toggle-owned gates:

```text
services/game-server/internal/game/simulation.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/simulation_bullets.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_weapons.go
```

Packet source and generated-output files:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/devtools/packets_generated.go
```

Related tests:

```text
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/game/world_simulation_options_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/tests/networking/inbound_devtools_test.go
```

Important non-ownership boundaries:

```text
client/scripts/devtools/
docs/devtools/client/
docs/services/game-server/
docs/data/
docs/protocol/
```

Client devtools owns input and presentation. Server devtools owns command handling and status projection. Game-server simulation docs own normal gameplay behavior. Data docs own packet schema and generation rules. Protocol docs own packet transport and compatibility behavior.

## Tests and verification

Relevant focused tests verify:

```text
single-player invincibility toggles on and off
all-player invincibility uses set-style behavior
invincible players do not die from asteroid collision
single-player infinite lives toggles on and off
all-player infinite lives uses set-style behavior
infinite-lives players die without losing lives
infinite lives persists across respawn
aggregate world freeze toggles all freeze flags
partial freeze plus aggregate freeze enables all flags
asteroid-only freeze stops asteroid movement only
bullet-only freeze stops bullet movement and expiry only
spawning-only freeze stops asteroid spawning only
spawns alias freezes spawning
collision-only freeze stops collision consequences only
unknown world-freeze target leaves freeze flags unchanged
frozen world stops asteroid movement
frozen world stops bullet movement and expiry
frozen world stops bullet spawning
frozen world stops asteroid spawning
frozen world skips ship/asteroid collision damage
frozen world skips bullet/asteroid collision, score, and asteroid split consequences
player freeze contributes to suspension behavior
all-player player freeze uses set-style behavior
debug status reflects player and world toggle state
debug status reports granular world-freeze flags
default builds enable devtools
nodevtools builds disable the devtools package gate
```

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/...
go test -buildvcs=false ./internal/game/...
go test -buildvcs=false ./tests/game/...
go test -buildvcs=false ./tests/networking/...
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

Packet schema changes should also be checked with the packet generation workflow:

```bash
data-sync -check -packets -go -gds
data-sync -diff -packets -go -gds
```

## Related docs

* [Server Devtools](./!README.md)
* [Devtools](../!README.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Devtools Window](../client/devtools-window.md)
* [Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Game Server](../../services/game-server/!README.md)
* [Game Server Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Game Server Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md)
* [Player Pause And Suspension](../../services/game-server/simulation/players/player-pause-and-suspension.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

Legacy devtools notes grouped client hotkeys, client overlays, server toggles, targeting, and telemetry in one document. This document is narrower: it covers server-owned gameplay-affecting toggle behavior and status output.

`debug_kill_player` shares implementation file space with toggle handlers in `toggles.go`, but it is a command, not a toggle. It belongs with player targeting, lifecycle, and command-surface documentation rather than being treated as part of the toggle set.

World freeze should be described as granular simulation gating. It is useful for debugging asteroid movement, bullet motion, spawning, and collision passes, but it should not be mistaken for match pause or full simulation suspension.
