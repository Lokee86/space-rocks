# Respawn Tools

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side devtools respawn command.

It covers how `debug_respawn_player` is routed, how target players are resolved, what server authority applies, how force-respawn differs from normal gameplay respawn, what client-visible effects confirm the command, and which code paths implement and verify the behavior.

## Overview

Respawn tools are debug-only server commands for recreating a player ship during an active development session.

The client can request a debug respawn from the devtools hotkey path or the devtools window, but the server owns whether the command is accepted and how the player is recreated. The command enters the server as:

```text
debug_respawn_player
```

The command routes through the server devtools packet path, not through normal gameplay `respawn` packet handling.

Current high-level flow:

```text
client devtools respawn request
-> websocket inbound envelope decode
-> inbound devtools packet routing
-> devtools.HandleCommand(...)
-> handleDebugRespawnPlayer(...)
-> target player validation
-> Game.DevtoolsSafeRespawnPosition(...)
-> Game.DevtoolsForceRespawnPlayer(...)
-> lane-native readback reflects recreated ship
```

Debug respawn uses the game-owned respawn placement and ship recreation seams, but it is not the same as normal gameplay respawn. Normal gameplay respawn requires the player session to have lives remaining, a zero respawn cooldown, and no active ship. Debug respawn is a force tool: it requires a valid target player session and blocks active players, but it resets respawn cooldown and recreates the ship through the devtools export seam.

## Debug-only scope

Respawn tools are for local development and diagnostics.

They may:

* Force a selected non-active player session back into active ship state.
* Force all non-active targetable player sessions back into active ship state.
* Reset the target session's respawn cooldown to zero.
* Recreate the player's active `runtime.Ship` from session-owned state.
* Recreate or update the player's server-side camera view.
* Use safe respawn placement near the session spawn position.

They must not:

* Become the normal player-facing respawn path.
* Let the client choose authoritative respawn position.
* Duplicate respawn logic in client code.
* Bypass the server-owned game instance.
* Mutate player-data or account persistence.
* Treat client UI selection as authority without server-side target validation.
* Route through normal `Game.HandlePacket` gameplay respawn handling.

The normal gameplay respawn packet remains separate:

```text
respawn
-> Game.HandlePacket(...)
-> respawnPlayer(...)
```

The debug command remains separate:

```text
debug_respawn_player
-> devtools.HandleCommand(...)
-> DevtoolsForceRespawnPlayer(...)
```

## Server authority

The server owns all gameplay consequences of debug respawn.

Inbound routing only classifies the packet and decodes it into `devtools.DebugCommand`. The command effect is owned by `services/game-server/internal/devtools/` and the narrow game-owned export seam in `services/game-server/internal/game/export_devtools_respawn.go`.

Debug respawn requires:

```text
current room exists
current game player ID exists
debug command packet decodes successfully
target player ID resolves to normalized player-N form
target game instance exists
target player session exists
target player is not currently active
safe respawn position can be resolved
force respawn succeeds
```

A single-player debug respawn requires a `target_player_id`. The server-side respawn handler does not fall back to the requesting player when `target_player_id` is empty. The client local hotkey path supplies the local player ID before sending the packet.

For all-player respawn, the command sends:

```text
target_scope = "all_players"
```

The server expands that scope through `Game.DevtoolsTargetPlayerIDs()`, which includes player IDs known through sessions and active ship state. Each resolved player ID then runs through the same single-target respawn handler. Active players are ignored instead of being recreated.

Debug respawn normalizes player IDs through the debug player ID helper. Accepted target IDs use the canonical form:

```text
player-1
player-2
player-3
```

Invalid or empty IDs are ignored.

The command currently logs request `x` and `y` values if they are present in the shared debug command shape, but those values do not choose the respawn position. Server respawn placement comes from:

```text
Game.DevtoolsSafeRespawnPosition(playerID)
```

That method delegates to the game-owned safe respawn placement logic using the target session's stored spawn position and collision shape.

## Client presentation

The client devtools layer only requests respawn and observes the result.

Client-side command paths include:

```text
DevToggle7
-> request_respawn_local_player()
-> request_respawn_player("single_player", local_player_id)
-> DevConnectionService.send_respawn_player(...)
-> debug_respawn_player packet
```

and devtools window controls:

```text
Respawn Player button
-> target resolver
-> request_respawn_player(target_scope, target_player_id)
-> DevConnectionService.send_respawn_player(...)
```

The client packet builder creates one of these shapes:

```json
{
  "type": "debug_respawn_player",
  "target_scope": "single_player",
  "target_player_id": "player-1"
}
```

or:

```json
{
  "type": "debug_respawn_player",
  "target_scope": "all_players"
}
```

The client does not locally recreate a ship. Confirmation comes through normal authoritative server state:

* World lane ship records include the recreated active ship.
* Session lane lifecycle records for `playerID` become `active`.
* Session/overlay lane respawn cooldown readback shows zero.
* Player camera/world state updates from the server-side camera view.

If the server ignores the command, the client receives no special rejection packet. The next lane readback simply continues to reflect the unchanged server state.

## Commands or controls

### `debug_respawn_player`

`debug_respawn_player` force-respawns one or more target player sessions.

Single-target command shape:

```text
type = "debug_respawn_player"
target_scope = "single_player"
target_player_id = "<player-id>"
```

All-player command shape:

```text
type = "debug_respawn_player"
target_scope = "all_players"
```

Current server behavior:

```text
if target_scope == "all_players":
    resolve target player IDs from game sessions and active player entities
    apply the single-target respawn path to each resolved player
    return handled

else:
    require target_player_id
    normalize target_player_id
    ignore if target is currently active
    resolve safe respawn position from the game session
    reset target session respawn cooldown to zero
    recreate target ship from session state
    create or update target camera view
    return handled
```

Single-target debug respawn does not allocate a new player session. If the target session is missing, the command is ignored.

Debug respawn does not preserve the old active ship. It writes a new ship into:

```text
game.entities.Players[playerID]
```

from the target session's `NewShip(...)` method.

The recreated ship inherits session-owned state:

```text
player ID
ship type ID
resolved ship stats
client config
damage options
selected primary weapon
selected secondary weapon
targeting state
```

Debug respawn resets weapon runtime state as part of creating a new ship. The selected weapons come from the session armory; transient per-ship weapon state starts fresh.

### Target scope behavior

`single_player` requires an explicit target player ID.

`all_players` expands to every known player ID from:

```text
game.playerSessions
game.entities.Players
```

The all-player path is intentionally not a fake player ID. It is a scope instruction.

Each all-player target is handled independently. Active players are ignored, while non-active sessions can be force-respawned.

## Telemetry

Debug respawn emits structured game logs through `logging.Game`.

Current log messages include:

```text
debug respawn player received
debug respawn player ignored
debug force respawn applied
```

The received log includes:

```text
player_id
target_player_id
x
y
```

Ignored logs include:

```text
player_id
target_player_id
```

Successful force-respawn logs include:

```text
player_id
target_player_id
x
y
```

where `x` and `y` are the server-selected respawn position.

There is no dedicated outbound `debug_respawn_result` packet. The result is observable through normal state projection and logs.

Debug status packets do not currently carry respawn-specific success or failure state. They report debug toggle state such as invincibility, infinite lives, freeze state, and player frozen state.

## Build/runtime gates

Server devtools are enabled in default game-server builds and disabled in `nodevtools` builds.

Default builds include devtools command handling through the enabled devtools gate. `nodevtools` builds disable devtools command handling before command effects run.

Relevant build-gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Runtime gates for respawn command application are:

```text
inbound packet type must be a devtools command type
current room must exist
current game player ID must exist
packet must decode into devtools.DebugCommand
devtools.HandleCommand must recognize debug_respawn_player
target game instance must exist
target player ID must normalize successfully
target player must not currently be active
target player session must exist
safe respawn position must resolve
force respawn must succeed
```

If the current session has no room or no current game player ID, inbound devtools routing consumes the packet without applying a command. This prevents debug command packets from falling through into normal gameplay packet routing.

## Code map

Primary server implementation:

```text
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/player_camera.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/packets_generated.go
```

Game-owned export seams:

```text
services/game-server/internal/game/export_devtools_respawn.go
services/game-server/internal/game/export_devtools_player_spawn.go
```

Gameplay respawn implementation used for comparison and safe placement:

```text
services/game-server/internal/game/session.go
services/game-server/internal/game/input.go
services/game-server/internal/game/players.go
services/game-server/internal/game/match.go
services/game-server/internal/game/rules/match.go
```

Packet source and generated output:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Client request-side references:

```text
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
client/scripts/devtools/devtools_target_resolver.gd
```

Related tests:

```text
services/game-server/tests/game/devtools_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/networking/inbound/
```

owns packet-family routing, not respawn behavior.

```text
services/game-server/internal/game/session.go
```

owns normal gameplay respawn rules and safe respawn placement.

```text
services/game-server/internal/devtools/
```

owns debug command dispatch and debug-only request handling.

```text
client/scripts/devtools/
```

owns request construction and UI presentation, not server mutation.

## Tests

Relevant server tests include:

```text
services/game-server/tests/game/devtools_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
```

Current focused coverage verifies that:

* `debug_respawn_player` is recognized as a devtools command type.
* all-player debug respawn recreates eligible non-active player entities.
* all-player debug respawn ignores active players.
* debug force-respawn creates a camera view when one is missing.
* debug force-respawn uses the dummy devtools camera config when creating that camera view.
* `session.NewShip(...)` copies player armory selections into recreated ship weapons.
* default builds handle devtools command types.
* `nodevtools` builds do not handle devtools command types.

Useful verification commands from `services/game-server/`:

```bash
go test -buildvcs=false ./internal/devtools/...
go test -buildvcs=false ./internal/game/...
go test -buildvcs=false ./tests/game -run 'Devtools|Respawn'
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Gameplay Network Adapter](../../services/game-server/networking/gameplay-network-adapter.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Active Player Avatar State](../../services/game-server/simulation/players/active-player-avatar-state.md)
* [Player Camera View State](../../services/game-server/simulation/players/player-camera-view-state.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Logging And Diagnostics](../../services/game-server/observability/logging-and-diagnostics.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

Debug respawn intentionally bypasses normal gameplay respawn cooldown gates. That makes it useful for development, but it should not be reused for player-facing respawn behavior.

The current debug respawn command uses safe server-selected respawn placement. It does not use client-sent `x` or `y` values as authoritative placement.

The server-side force-respawn camera path uses a dummy 1280 by 720 camera config when it must create a missing camera view. Normal gameplay respawn uses the existing player camera view and preserves valid client viewport configuration through `setPlayerCameraViewLocked`.

Earlier legacy notes described respawn tools as using the existing per-player respawn guards. Current code uses a debug-specific force-respawn path instead: active players are ignored, invalid targets are ignored, and non-active sessions are recreated through the devtools export seam.

