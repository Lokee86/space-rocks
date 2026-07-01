# Player Counters

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side devtools surface for player counter mutation.

It covers the debug commands that set or add score and lives, how those commands resolve player targets, how they delegate into the authoritative game-server player counter seam, what the client may present, what telemetry confirms the change, and which code paths implement the behavior.

## Overview

Server devtools player counters are debug-only commands for mutating the authoritative per-match player counters that already exist in the game simulation.

The current counter commands are:

```text
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
```

These commands do not create a separate debug counter store. They route through `services/game-server/internal/devtools/player_counters.go`, resolve one or more player IDs, and then call narrow game-owned adapters in:

```text
services/game-server/internal/game/export_devtools_player_counters.go
```

Those adapters delegate to the normal game counter seam:

```text
services/game-server/internal/game/player_counters.go
```

The authoritative values remain stored on `playerSession`, not on the live `runtime.Ship` avatar. Score and lives therefore survive ship removal, death, pending respawn, and eliminated states as long as the player session still exists.

Counter values are clamped at zero by the gameplay counter seam. Devtools can set or add negative values, but the resulting score or lives value cannot become negative.

## Debug-only scope

Player counter devtools may:

* Set one player’s score.
* Add to or subtract from one player’s score.
* Set one player’s lives.
* Add to or subtract from one player’s lives.
* Apply the same score or lives command to all known player targets.
* Use later lane packet readback as confirmation that the counter changed.

Player counter devtools must not:

* Persist score or lives directly to player-data storage.
* Bypass `Game.SetPlayerScore`, `Game.AddPlayerScore`, `Game.SetPlayerLives`, or `Game.AddPlayerLives`.
* Mutate `runtime.Ship` fields as a substitute for session-owned counters.
* Create implicit player sessions when a target player ID is missing.
* Treat the client’s requested value as confirmed before the server projects the updated state.
* Duplicate scoring, life-loss, respawn, pickup, or match-result rules in debug-only code.

This surface exists for controlled local development and diagnostics. It is not normal gameplay progression, match reward policy, or player-data persistence.

## Server authority

The server owns all player counter consequences.

The command path is:

```text
client debug packet
-> WebSocket read loop
-> inbound packet envelope decode
-> inbound devtools packet classification
-> packetcodec.Decode(..., devtools.DebugCommand)
-> devtools.HandleCommand(room.GameInstance(), currentGamePlayerID, command)
-> handleDebugSetScore / handleDebugAddScore / handleDebugSetLives / handleDebugAddLives
-> resolveCommandTargetPlayerIDs
-> Game.DevtoolsSetPlayerScore / Game.DevtoolsAddPlayerScore / Game.DevtoolsSetPlayerLives / Game.DevtoolsAddPlayerLives
-> Game.SetPlayerScore / Game.AddPlayerScore / Game.SetPlayerLives / Game.AddPlayerLives
-> playerSession.Score or playerSession.Lives
```

The devtools package owns command dispatch and target resolution. It does not own authoritative counter storage.

The game package owns the mutation seam. Public counter methods lock `game.mu`, find the player session, clamp the resulting value, write the session counter, and return a `PlayerCounterChange`.

The returned change includes:

```text
PlayerID
Found
Before
After
Delta
```

Devtools handlers use `Found` to decide whether a command affected at least one target. A command returns `false` when no target player session was found or when the target game instance is nil.

Missing player sessions are not created by score or lives commands. If a requested player ID does not exist in `game.playerSessions`, the game counter seam returns `Found: false` and does not mutate state.

## Target resolution

Player counter commands use the shared devtools player-target resolver.

Current target scopes are:

```text
single_player
all_players
```

For `target_scope = "all_players"`, the server asks the game instance for `DevtoolsTargetPlayerIDs()`. That game-owned helper returns the sorted union of known player session IDs and active ship IDs, excluding empty IDs. The counter seam still only mutates IDs that resolve to existing player sessions.

For any other scope, including an empty or unknown scope, the command is treated as single-player targeting:

```text
target_player_id if present
else requesting player ID
```

The requesting player ID comes from the current WebSocket session’s active game player ID.

A valid devtools packet requires a current room and a current active game player before command handling is reached. If either is missing, inbound devtools routing consumes the packet and applies no command.

## Client presentation

The client presents player counter controls in the devtools window and sends command packets through the normal networking path.

The server-side player counter doc does not own the client UI, but the server expects these client-originated packet shapes:

```text
debug_set_score
fields: type, target_scope, optional target_player_id, score

debug_add_score
fields: type, target_scope, optional target_player_id, amount

debug_set_lives
fields: type, target_scope, optional target_player_id, lives

debug_add_lives
fields: type, target_scope, optional target_player_id, amount
```

The client may offer target rows such as local player, selected player, Game Target, or All Players. Those are presentation conveniences. Server authority starts only after the command packet reaches the game server and is resolved against the current room and game instance.

For player-only counter commands, non-player canonical targets must not become `target_player_id` values. The server counter surface accepts only player IDs and scopes.

## Commands or controls

### `debug_set_score`

Sets each resolved target player session’s score to `command.Score`.

The resulting score is clamped at zero.

```text
score = max(command.Score, 0)
```

### `debug_add_score`

Adds `command.Amount` to each resolved target player session’s score.

The resulting score is clamped at zero.

```text
score = max(current_score + command.Amount, 0)
```

Negative amounts are allowed for debugging, but they cannot reduce score below zero.

### `debug_set_lives`

Sets each resolved target player session’s lives to `command.Lives`.

The resulting lives value is clamped at zero.

```text
lives = max(command.Lives, 0)
```

This directly changes the session-owned lives counter. It does not create or destroy a ship by itself.

### `debug_add_lives`

Adds `command.Amount` to each resolved target player session’s lives.

The resulting lives value is clamped at zero.

```text
lives = max(current_lives + command.Amount, 0)
```

Negative amounts are allowed for debugging, but they cannot reduce lives below zero.

### Command result behavior

Each command iterates all resolved target player IDs and applies the requested counter operation.

The handler returns `true` when at least one target player session is found and mutated. It returns `false` when the game instance is nil or no resolved player ID maps to an existing player session.

There is no dedicated success or failure response packet for these commands. Confirmation is observed through subsequent authoritative lane packet readback.

## Telemetry

Player counter devtools do not emit a dedicated telemetry packet.

Counter visibility comes from normal server state projection:

```text
lane packet.lives
lane packet.player_sessions[player_id].score
lane packet.player_sessions[player_id].lives
```

`overlay lane receiver-local lives/readout` is the requesting player's lives convenience projection.

`session lane player records` is the multi-player read model for durable session counters. Devtools readouts that compare local player or target state should read score and lives from player session state, not from active ship state.

Counter command effects are visible only after the server mutates game state and a later lane packet reaches the client. The client should not treat the outgoing command packet as proof that the counter changed.

Debug status packets currently report flags such as invincible, infinite lives, world frozen, and player frozen. They do not carry score or lives values.

## Build/runtime gates

Player counter commands depend on the general server devtools command path.

Relevant gates:

```text
default builds:
devtools.Enabled() == true

nodevtools builds:
devtools.Enabled() == false

devtools.ShouldHandleCommand(packet_type):
IsCommandType(packet_type) && Enabled()
```

Player counter handlers themselves are not the build gate. They assume the command has already been routed into `devtools.HandleCommand`.

Runtime command routing also requires:

```text
current room exists
current active game player ID exists
packet decodes as devtools.DebugCommand
command type matches a devtools counter packet
resolved target player session exists
```

If the current room or active game player ID is missing, inbound devtools routing consumes the packet and applies no mutation.

If packet decode fails, networking logs the decode failure and applies no mutation.

If the command resolves to no existing player session, the handler returns `false` and applies no mutation.

## Relationship to gameplay counters

Player counter devtools intentionally reuse the gameplay counter seam.

The authoritative gameplay counter doc owns the broader score/lives model:

```text
services/game-server/internal/game/player_counters.go
```

That seam is also used by normal gameplay flows such as score awards, fatal damage life loss, pickup life effects, match facts, and lane packet projection.

Devtools only supplies an alternate command path into that seam. It does not own:

* scoring policy
* score awards from asteroid destruction
* fatal damage rules
* infinite-lives behavior
* pickup effect rules
* respawn eligibility
* match-over policy
* match-result summaries
* player-data persistence

## Code map

Primary server devtools files:

```text
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/packets_generated.go
```

Game-owned counter adapters:

```text
services/game-server/internal/game/export_devtools_player_counters.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/export_devtools_player_spawn.go
```

Inbound routing:

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/protocol/packetcodec/codec.go
```

Build gates:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Packet source and generated output:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Client presentation and send path references:

```text
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/devtools_target_resolver.gd
client/scripts/devtools/devtools_window.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
```

Related tests:

```text
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/tests/game/devtools_test.go
client/tests/unit/devtools/gameplay_debug_flow_test.gd
client/tests/unit/devtools/context/test_devtools_command_context.gd
client/tests/unit/test_devtools_target_resolver.gd
client/tests/unit/devtools/devtools_window_test.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/scoring/
services/game-server/internal/game/damage/
services/game-server/internal/game/pickups/
services/game-server/internal/rooms/
services/game-server/internal/networking/outbound/
services/player-data/
client/scripts/gameplay/
client/scripts/ui/
```

## Tests

Server devtools player counter tests cover:

* Setting score to an exact value.
* Clamping negative score to zero.
* Adding positive score.
* Adding negative score.
* Clamping score below zero.
* Setting lives to an exact value.
* Clamping negative lives to zero.
* Adding positive lives.
* Adding negative lives.
* Clamping lives below zero.
* Applying score commands to all players.
* Applying lives commands to all players.
* Targeting another player by explicit `target_player_id`.
* Falling back to the requesting player when no target is supplied.
* Returning all target player IDs for `all_players`.
* Treating unknown scopes as single-player scopes.
* Default and `nodevtools` devtools gate behavior.

Broader game tests verify that devtools counter mutations appear through lane packet readback and that the same gameplay counter seam remains authoritative for score and lives projection.

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/...
go test -buildvcs=false ./tests/game -run 'Debug.*Score|Debug.*Lives|SetPlayerScore|SetPlayerLives'
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Devtools](../client/!INDEX.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Game Server Simulation Players](../../services/game-server/simulation/players/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

The legacy devtools documentation correctly identified the important boundary: score and lives devtools adapters must delegate to the shared player counter seam instead of mutating debug-only state.

The current server counter command surface has a lower bound of zero and no upper bound.

`debug_add_score` and `debug_add_lives` intentionally accept negative amounts for development use. The game-owned counter seam is responsible for clamping the result.

`target_player_id` remains a devtools/player-only command field. Normal gameplay targeting should continue using canonical target identity rather than adopting `target_player_id` as a general gameplay model.
