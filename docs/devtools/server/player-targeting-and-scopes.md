## Player Targeting And Scopes

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes server-side devtools player targeting and scope resolution.

It covers how server debug commands interpret `target_player_id` and `target_scope`, which commands support all-player fanout, how player targets are resolved before mutation, and how this player-only debug surface stays separate from canonical gameplay targeting.

## Overview

Server devtools commands are packet-driven debug requests. The client sends a debug packet, networking decodes it into a generated `DebugCommand`, and the server routes it through `devtools.HandleCommand`.

Player-targeted devtools commands use two debug fields:

```text
target_player_id
target_scope
```

`target_player_id` is a player-only debug command field. It is not canonical gameplay target state.

`target_scope` controls whether a supported command applies to one effective player or to every current server-known player target. Current server scope constants are:

```text
single_player
all_players
```

Only `all_players` receives special server handling. Empty, `single_player`, and unknown scope strings behave as single-player command requests.

The shared server resolver is:

```text
resolveCommandTargetPlayerIDs(game, requesting_player_id, command)
```

Resolution behavior is:

```text
target_scope == all_players
-> return game.DevtoolsTargetPlayerIDs()

otherwise
-> use target_player_id when present
-> fall back to requesting_player_id when target_player_id is empty
```

The client may expose UI labels such as `Game Target` or `All Players`, but the server does not receive those UI sentinel values as authority. By the time a player-only debug command reaches the server, the client has either emitted a concrete `target_player_id`, emitted `target_scope=all_players`, or emitted no target so the server can use the requesting player fallback where that command supports it.

Canonical gameplay targeting remains separate:

```text
target_kind
target_id
```

Non-player canonical targets such as `asteroid`, `bullet`, `pickup`, or `enemy` may be valid gameplay targets for inspection and readback, but they must not become `target_player_id` values for player-only server devtools commands.

## Debug-only scope

This system is development/debug tooling only.

It applies to server devtools command handling for:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_player
debug_kill_player
debug_respawn_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_spawn_entity when entity_type=player
```

It also constrains how player-only command fields are interpreted by server handlers.

This system does not define normal gameplay targeting, weapon targeting, HUD target presentation, spectate behavior, durable profile identity, account identity, matchmaking identity, or permanent player selection rules.

## Server authority

The server owns all gameplay mutation behind player-targeted devtools commands.

The client may request:

```text
make a player invincible
give a player infinite lives
freeze a player
kill a player
respawn a player
set or add score
set or add lives
spawn a player slot
apply one of those operations to all players
```

The server decides:

```text
which player IDs the command resolves to
whether the game instance exists
whether the requesting session has a current game player ID
whether a target player exists
whether the command applies to session state, active avatar state, or both
whether an all-player fanout has any targets
whether a respawn target is dead enough to respawn
whether a requested debug-spawn player ID can be normalized and reserved
```

All-player fanout is game-owned. `Game.DevtoolsTargetPlayerIDs()` returns a sorted unique list from both:

```text
playerSessions
entities.Players
```

That means all-player commands can reach session-only player state and active ship state where the owning command supports those states. The client does not send a fake player ID for `All Players`.

The requesting player ID still matters. It is passed from the network session into command handling and is used for fallback targeting, debug kill damage source identity, bullet-spawn owner identity, and log context.

## Client presentation

The server does not own target labels, dropdown rows, telemetry widgets, or window controls.

Client-side devtools presentation may expose:

```text
explicit player rows
All Players rows
Game Target rows
local player fallback behavior
target telemetry
target status labels
```

Those are presentation and request-building concerns. The server receives only packet fields.

Important presentation boundary:

```text
All Players
-> target_scope = all_players
-> target_player_id = ""

Game Target for a player
-> target_player_id = <player id>

Game Target for a non-player entity
-> no effective player-only command target
```

Server docs should describe only the server packet interpretation and command effects. Client docs own selector labels, target readmodels, window controls, and packet construction behavior.

## Commands or controls

The server command surface is generated from `shared/packets/debug.toml` and decoded into `DebugCommand`.

Relevant fields are:

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

Player-targeted command groups behave differently.

Commands that use the shared target resolver:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_player
debug_kill_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
```

For those commands:

```text
target_scope=all_players
-> apply to every DevtoolsTargetPlayerIDs() result

target_player_id=<id>
-> apply to that player ID

target_player_id empty and not all_players
-> apply to requesting player ID
```

All-player toggle commands use a group toggle rule:

```text
if any target player is inactive/missing for that feature or currently disabled
-> set the feature enabled for every resolved target

if every resolved target is currently enabled
-> set the feature disabled for every resolved target
```

This applies to:

```text
invincible
infinite_lives
player_frozen
```

`debug_kill_player` resolves target players through the same shared resolver, but it only applies the kill to players that are currently active according to match decision state.

Score and lives commands resolve target players through the same shared resolver, then call the game-owned player counter seam:

```text
DevtoolsSetPlayerScore
DevtoolsAddPlayerScore
DevtoolsSetPlayerLives
DevtoolsAddPlayerLives
```

Counter clamping belongs to the game counter implementation, not the devtools target resolver.

`debug_respawn_player` has a narrower single-target rule. It supports `target_scope=all_players` by expanding through the shared resolver and then respawning each resolved player target. For a single-player respawn, the command requires an explicit `target_player_id`; an empty target is ignored instead of falling back to the requesting player.

Respawn target IDs are normalized through the debug player-ID parser. Valid debug spawn/respawn player IDs use this shape:

```text
player-<positive number>
```

`Player-<positive number>` is accepted and normalized to lowercase `player-<number>`.

Respawn refuses to apply when the normalized target player is already active. When accepted, the server chooses a safe respawn position and forces the player back into the game through the game-owned respawn adapter.

`debug_spawn_entity` uses `target_player_id` only when spawning a player entity. In that path, `target_player_id` is an optional requested player slot. If it is present, the server normalizes and reserves that debug player ID. If it is empty, the server allocates the first available `player-N` ID within the configured player ID range.

Player spawn does not use `target_scope`. It is a placement/spawn request, not a player-command fanout request.

World and entity-wide commands do not use player target resolution:

```text
toggle_debug_freeze_world
debug_spawn_pickup
debug_spawn_entity for asteroid or bullet
debug_begin_continuous_bullet_stream
debug_clear_bullets
debug_clear_asteroids
```

`toggle_debug_freeze_world` uses `freeze_target` instead of `target_player_id`:

```text
all
asteroids
bullets
spawning
spawns
collisions
```

Empty `freeze_target` means `all`.

## Telemetry

Player targeting and scope resolution produce indirect telemetry through logs, debug status packets, and ordinary authoritative state readback.

Server command handlers log relevant target context such as:

```text
target_player_id
enabled
freeze_target
spawned_player_id
has_target_player_id
x
y
```

Debug status output exposes current debug feature state by player ID. The status packet contains:

```text
debug_status
debug_statuses
```

`debug_status` is the receiver/current-player status snapshot. `debug_statuses` is a map of player ID to debug status for every projected player status entry.

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

`target_scope` is not itself a durable server state field and is not emitted as telemetry. It is command input used to choose effective targets for the current command.

After mutations, clients observe the result through normal authoritative readback:

```text
debug_status packets
state packets
player lifecycle
player session score/lives state
active player ship state
entity maps
```

The server should not add a separate debug-only state model for player targeting when existing game state and debug status output already expose the result.

## Build/runtime gates

Server devtools have a build-tag gate:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Default builds return:

```text
devtools.Enabled() == true
```

`nodevtools` builds return:

```text
devtools.Enabled() == false
```

The devtools package exposes:

```text
ShouldHandleCommand(packet_type)
```

That helper combines command-type classification with `Enabled()`.

Outbound debug status and debug shape catalog presentation check `devtools.Enabled()` before sending debug output.

Inbound command routing is implemented through the networking inbound devtools handlers. Those handlers classify debug packet types before normal gameplay packet decode, require an active room and current game player ID, decode the packet into `DebugCommand`, and call `devtools.HandleCommand`.

Runtime command gates include:

```text
no current room -> command ignored
no current game player ID -> command ignored
debug packet decode failure -> warning log, command ignored
unknown command type -> HandleCommand returns false
all_players with no game instance -> no target IDs
respawn with empty target_player_id -> ignored
respawn of active player -> ignored
spawn-player with invalid requested player ID -> ignored
spawn-player with occupied requested player ID -> ignored
player counter command with no found target -> returns false
```

Devtools command handlers must keep routing through game-owned adapters. They should not mutate unrelated game internals directly when an existing gameplay seam owns the behavior.

## Code map

Primary server devtools targeting files:

```text
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
```

Player-targeted command handlers:

```text
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/player_ids.go
```

Networking command routing:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/protocol/packetcodec/
```

Build gates:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Game-owned devtools adapters:

```text
services/game-server/internal/game/export_devtools_player_spawn.go
services/game-server/internal/game/export_devtools_player_counters.go
services/game-server/internal/game/export_devtools_toggles.go
```

Related authoritative game state and targeting files:

```text
services/game-server/internal/game/targeting.go
services/game-server/internal/game/player_targeting.go
services/game-server/internal/game/session.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/state_packet.go
```

Packet source and generated output:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Important non-ownership boundaries:

```text
client/scripts/devtools/
client/scripts/gameplay/targeting/
services/game-server/internal/game/targeting/
services/game-server/internal/game/damage/
services/game-server/internal/game/rules/
services/game-server/internal/networking/
docs/services/game-server/simulation/targeting/
docs/devtools/client/
docs/data/
docs/protocol/
```

Client devtools owns selector presentation and packet construction.

Game targeting owns canonical `target_kind` and `target_id`.

Damage owns debug kill damage resolution.

Rules own match/player lifecycle classification.

Networking owns packet routing and session identity.

Data/protocol docs own packet schema source and generated packet expectations.

## Tests

Relevant server tests include:

```text
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
```

Current coverage verifies:

```text
all_players expands to server-known target player IDs
explicit target_player_id wins for single-player scope
empty single-player command target falls back to requesting player
unknown scope behaves like single-player scope
all-player invincible toggles all players
all-player infinite-lives toggles all players
all-player player-freeze toggles all players
all-player toggle commands enable everyone until everyone is enabled
all-player toggle commands disable everyone when everyone is enabled
all-player kill targets all active players
score and lives commands support all-player fanout
score and lives commands support explicit other-player targeting
score and lives commands fall back to the caller when target_player_id is empty
DevtoolsTargetPlayerIDs includes session-only and active-ship player targets
command type classification includes current debug command packet types
default builds enable ShouldHandleCommand for devtools command packet types
nodevtools builds disable ShouldHandleCommand
```

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools
go test -buildvcs=false ./internal/game -run 'DevtoolsTargetPlayerIDs|DevtoolsSpawnPlayerShip'
```

Useful nodevtools verification:

```bash
cd services/game-server
go test -buildvcs=false -tags nodevtools ./internal/devtools
```

Run packet checks when changing `DebugCommand`, debug packet fields, or generated packet output:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Client Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Devtools Window](../client/devtools-window.md)
* [Canonical Target State](../../services/game-server/simulation/targeting/canonical-target-state.md)
* [Target Selection And Status](../../services/game-server/simulation/targeting/target-selection-and-status.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Realtime Protocol](../../protocol/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

The most important boundary is that `target_player_id` remains a player-only debug command compatibility field. It must not become normal gameplay target state.

Server devtools targeting should stay command-specific. Adding a new player-targeted debug command should reuse `target_scope` and `resolveCommandTargetPlayerIDs` only when all-player fanout and requesting-player fallback are actually correct for that command.

`single_player` exists as a named scope constant, but the current server implementation only special-cases `all_players`. Unknown scopes therefore behave like single-player command input.

`Game Target` is a client-side selector concept. The server should not learn about a `__game_target__` sentinel; it should receive either concrete player command fields or no effective command target.
