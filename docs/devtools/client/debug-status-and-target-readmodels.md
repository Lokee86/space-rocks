## Debug Status And Target Readmodels

Parent index: [Client](./!README.md)

## Purpose

This document describes the client devtools readmodels that consume server debug status packets and normalized gameplay state.

It covers how the client devtools window shows server-owned debug status, player target rows, game-target rows, and raw local/target telemetry without taking gameplay authority away from the server.

## Overview

Client debug status and target readmodels are diagnostic presentation surfaces.

They combine two server-fed inputs:

```text
debug_status packet
-> debug_status
-> debug_statuses

state packet normalized by gameplay state reader
-> self_id
-> server_players
-> player_sessions
-> player_lifecycle
-> server_asteroids
-> server_bullets
-> server_pickups
-> optional server_enemies / enemies
```

The debug status packet provides current debug-toggle state. The normalized gameplay state provides the entity/session maps needed to build target selectors and raw telemetry panels.

The client uses these inputs to refresh:

```text
world freeze status labels
per-player invincible target rows
per-player infinite-lives target rows
per-player freeze target rows
kill-player target rows
respawn-player target rows
score/lives target rows
game target options
local player telemetry
target telemetry
```

The readmodels do not mutate gameplay directly. Window controls send debug command packets or targeting requests through normal client networking. The server remains authoritative for whether commands are accepted and how they affect gameplay.

## Debug-only scope

This documentation covers client-only devtools presentation and readmodel coordination.

Debug-only surfaces include:

```text
DevtoolsWindow status labels
DevtoolsWindow target selectors
LocalPlayerTelemetry panel
TargetTelemetry panel
Game Target selector
All Players selector rows
feature-state labels such as Active / Inactive
```

These surfaces are not player-facing HUD. They are diagnostic views over server state and packet state.

Debug status and target readmodels must not become alternate gameplay state. They may display server-fed data and emit debug command requests, but they must not locally apply invincibility, lives, freeze, target, spawn, respawn, score, or clear-entity effects.

## Server authority

The server owns debug status.

The server emits `debug_status` packets from the WebSocket write loop when debug status output is eligible. Current eligibility requires:

```text
room exists
room has a game instance
devtools are enabled
room state is InGame or GameOver
session has a current game player id
```

The write loop sends debug status on a slower cadence than gameplay state. Current code writes normal gameplay presentation state every server write tick, then sends debug status every `debugStatusWriteIntervalTicks`, currently `8`.

The server packet contains:

```text
type: debug_status
debug_status: status for the receiving/current player
debug_statuses: map of every match player id to that player's status
```

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

`debug_status` drives receiver-local/global status labels in the devtools window. `debug_statuses` drives per-player selector labels.

The client does not request these status snapshots directly. They arrive as outbound server telemetry. Client controls can cause later status changes only by sending normal debug command packets or target request packets.

## Client presentation

### Debug status packet reading

`DebugStatusPacketReader.read(packet)` extracts only:

```text
debug_status
debug_statuses
```

If either value is malformed, the reader replaces it with an empty dictionary. This keeps window refresh code tolerant of bad or incomplete debug packets.

The normalized result is:

```gdscript
{
	"debug_status": Dictionary,
	"debug_statuses": Dictionary,
}
```

### Debug status fanout

Current inbound flow:

```text
ClientConnectionService.debug_status_received
-> GameplaySessionController.handle_debug_status_packet
-> GameplayComposition.apply_devtools_debug_status_packet
-> GameplayShellFlow.apply_devtools_debug_status_packet
-> GameplayFlowComposer.apply_devtools_debug_status_packet
-> GameplayDevtoolsContext.apply_debug_status_packet
-> DevtoolsGameplayStateContext.apply_debug_status_packet
-> DebugStatusPacketReader.read
-> DevtoolsDisplayRefreshFlow.apply_debug_status_packet
```

`DevtoolsDisplayRefreshFlow.apply_debug_status_packet` does two things:

```text
debug_statuses -> DevtoolsPlayerTargetModel.apply_debug_statuses
debug_status -> DevtoolsWindowController.apply_debug_status
```

Then it refreshes the player-target selectors so feature state labels reflect the latest per-player status map.

### Gameplay state fanout

Target readmodels are refreshed from normalized gameplay state, not from raw packet dictionaries.

Current flow:

```text
state packet
-> GameplayStatePacketReader.read
-> GameplayComposition.apply_gameplay_state
-> GameplayShellFlow / GameplayFlowComposer
-> GameplayDevtoolsContext.apply_gameplay_state
-> DevtoolsGameplayStateContext.apply_gameplay_state
-> DevtoolsDisplayRefreshFlow.refresh_gameplay_state
-> DevtoolsPlayerTargetModel.apply_gameplay_state
```

The target model caches:

```text
self_id
server_players
player_sessions
server_asteroids
server_bullets
server_pickups
server_enemies
player_lifecycle
game_target_kind
game_target_id
game_target_player_id
```

`game_target_kind` and `game_target_id` are read from the local player's synced ship state in `server_players[self_id]`.

If the current state has no generic `target_kind` / `target_id` but still has a compatibility `target_player_id`, the model treats that as a player target.

### Target rows

`target_rows()` builds raw player rows from the union of `player_lifecycle` and `server_players`.

Each row contains:

```text
player_id
status
alive
is_self
label
```

The status label is currently:

```text
ALIVE when lifecycle is active or the player exists in server_players
DEAD otherwise
```

`target_rows()` does not include `All Players`. It is the base player row set used by more specific selector builders.

### All Players rows

Several debug controls support an `All Players` selector option.

The model uses:

```text
player_id: __all_players__
label: All Players
```

The window controller converts that selected row into:

```text
target_scope: all_players
target_player_id: ""
```

All-player rows are used for:

```text
Kill Player
Respawn Player
Invincible
Infinite Lives
Freeze Player
Set Score
Add Score
Set Lives
Add Lives
```

Spawn Player and Game Target do not use all-player targeting.

### Game Target rows

The model exposes a compact `Game Target` row only when the canonical game target is a player.

The special row uses:

```text
player_id: __game_target__
label: Target : P<n>
```

Examples:

```text
player-2  -> Target : P2
player-10 -> Target : P10
```

For non-player canonical targets such as `asteroid`, `bullet`, `pickup`, or `enemy`, player-only controls do not receive a `Game Target` row.

Non-player targets can still appear in raw target telemetry.

### Feature status rows

Feature-specific selector rows are built from `debug_statuses`.

Current feature-specific row builders:

```text
invincible_target_rows()
infinite_lives_target_rows()
player_frozen_target_rows()
```

Each row label uses feature state wording:

```text
<player-id>: Active
<player-id>: Inactive
```

The feature values are read from the per-player `debug_statuses` map:

```text
invincible
infinite_lives
player_frozen
```

If a player's debug status is missing or malformed, the feature is treated as inactive for selector-label purposes.

### Local and target telemetry

The devtools window has two raw telemetry panels:

```text
LocalPlayerTelemetry
TargetTelemetry
```

Each panel has a source selector. Current source options are:

```text
StatePacket.entities
StatePacket.player_world_states
```

Internal source keys are:

```text
players
player_world_states
```

For local telemetry:

```text
players
-> server_players[self_id]

player_world_states
-> player_sessions[self_id]
```

For target telemetry:

```text
players
-> selected player, asteroid, bullet, pickup, or enemy state

player_world_states
-> selected player session state only
```

`player_world_states` returns empty for non-player targets.

Target telemetry includes:

```text
target_kind
target_id
raw sorted key/value body
```

The telemetry renderer does not hand-map score, lives, health, shields, or target fields into custom UI widgets. It renders raw dictionary keys in sorted order.

## Commands or controls

Debug status and target readmodels support controls indirectly.

The readmodels produce selector rows. The window and controller convert selected rows into command requests.

Current relevant controls include:

```text
Invincible
Infinite Lives
Freeze World
Freeze Asteroids
Freeze Bullets
Freeze Spawns
Freeze Collisions
Freeze Player
Kill Player
Respawn Player
Set Score
Add Score
Set Lives
Add Lives
Set Game Target
Clear Game Target
```

Player-targeted controls resolve through `DevtoolsTargetResolver`.

Target resolution rules:

```text
explicit selected player wins
Game Target resolves only when canonical target_kind is player
All Players resolves to target_scope=all_players
local player fallback applies only where the command path supports it
non-player canonical targets do not become player targets
```

The Set Game Target button sends a normal target request packet with `target_kind=player` and the selected player ID. Clear Game Target sends the same path with an empty target.

The client blocks single-player command requests when target resolution produces an empty player ID.

## Telemetry behavior

Debug status is live diagnostic telemetry, not analytics.

The client surfaces:

```text
receiver/global status labels from debug_status
per-player feature labels from debug_statuses
raw local state from normalized gameplay state
raw target state from normalized gameplay state
canonical target kind and id
```

The status packet and the gameplay state packet are separate lanes.

`debug_status` explains current debug-control state. The gameplay `state` packet explains entity/session world state. The client combines them only for devtools presentation.

## Build/runtime gates

Client-side devtools input has a public-build gate:

```text
client/scripts/devtools/dev_tools_build_flags.gd
```

When `public_build` is true, the script removes `DevToggle0` through `DevToggle9` input events from `InputMap`.

Debug status readmodel code still remains ordinary client code. The server-side `nodevtools` build tag controls whether the game server emits and handles devtools status/command behavior. When server devtools are disabled, debug status packets are not eligible to send.

## Data ownership

The client readmodels own only transient presentation state.

Owned transient state includes:

```text
latest debug_status
latest debug_statuses
cached gameplay state maps
latest selector rows
latest telemetry source choices
latest local/target telemetry dictionaries
latest game target kind/id
latest local player id
```

The readmodels do not persist data.

Source packet data comes from:

```text
shared/packets/debug.toml
shared/packets/gameplay.toml
```

Generated client packet constants and builders live in:

```text
client/scripts/generated/networking/packets/packets.gd
```

Server-side debug status generation lives in game-server devtools and outbound networking code. Client readmodels consume the packet output only.

## Invariants

Debug status and target readmodels must preserve these rules:

```text
server owns gameplay mutation
client reads debug status as telemetry
client reads gameplay state as a presentation readmodel
client selectors may emit command requests but must not apply command effects locally
Game Target appears in player-only controls only when target_kind is player
All Players is a target scope, not a fake player id
non-player targets may feed raw target telemetry but not player-only command targeting
malformed debug status payloads degrade to empty dictionaries
raw telemetry panels render packet dictionaries rather than bespoke per-stat UI state
```

## Code map

Primary client readmodel files:

```text
client/scripts/devtools/debug_status_packet_reader.gd
client/scripts/devtools/devtools_player_target_model.gd
client/scripts/devtools/devtools_display_refresh_flow.gd
client/scripts/devtools/devtools_target_resolver.gd
```

Devtools context and fanout files:

```text
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/context/devtools_gameplay_state_context.gd
client/scripts/devtools/context/devtools_state_context.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
```

Window presentation files:

```text
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_window.gd
client/scenes/devtools/devtools_window.tscn
```

Client inbound routing files:

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/gameplay/gameplay_composition.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
```

Gameplay state reader files:

```text
client/scripts/gameplay/state/gameplay_state_packet_reader.gd
client/scripts/gameplay/state/gameplay_state_apply_flow.gd
```

Generated and source packet files:

```text
shared/packets/debug.toml
shared/packets/gameplay.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
```

Server producer files:

```text
services/game-server/internal/devtools/status.go
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/websocket_write.go
```

Relevant server build-gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Important non-ownership boundaries:

```text
client/scripts/world/
client/scripts/ui/
services/game-server/internal/game/
services/game-server/internal/devtools/
services/game-server/internal/networking/
docs/services/client/
docs/devtools/server/
docs/data/
docs/protocol/
```

`world` owns entity presentation and visual sync.

`ui` owns player-facing UI. Devtools window controls are not normal HUD behavior.

`game` owns authoritative simulation state and gameplay readmodels.

`devtools` owns server debug command/status behavior.

`networking` owns transport and outbound packet timing.

`data` owns packet source-of-truth and generated-output pipeline documentation.

`protocol` owns packet and transport protocol documentation.

## Tests and verification

Relevant client tests include:

```text
client/tests/unit/devtools/debug_status_packet_reader_test.gd
client/tests/unit/test_devtools_player_target_model.gd
client/tests/unit/test_devtools_display_refresh_flow.gd
client/tests/unit/test_devtools_target_resolver.gd
client/tests/unit/test_devtools_window_controller.gd
client/tests/unit/devtools/devtools_window_test.gd
client/tests/unit/test_gameplay_devtools_context.gd
client/tests/unit/devtools/context/test_devtools_state_context.gd
client/tests/unit/devtools/context/test_devtools_command_context.gd
```

Relevant inbound routing and gameplay state tests include:

```text
client/tests/unit/test_gameplay_state_packet_reader.gd
client/tests/unit/test_gameplay_state_apply_flow.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/test_gameplay_session_controller.gd
```

Relevant server tests include:

```text
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/target_player_ids_test.go
```

Run client tests after changing the reader, target model, display refresh flow, window controller, or devtools context wiring.

Run server outbound/debug tests after changing the debug status packet shape, eligibility, cadence, or status field projection.

Run packet generation checks when changing shared packet source files.

## Related docs

* [Client Devtools](./!README.md)
* [Devtools](../!README.md)
* [Server Devtools](../server/!README.md)
* [Devtools Design](../design/!README.md)
* [Client Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
* [Client Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
* [Client Input And Targeting](../../services/client/input-and-targeting.md)
* [Game Server](../../services/game-server/!README.md)
* [Packet Schema Pipeline](../../data/packet-schemas.md) - shared packet schema and generated output documentation.

## Notes

The legacy devtools notes correctly separated raw devtools telemetry from the player-facing HUD. That rule still applies: `LocalPlayerTelemetry` and `TargetTelemetry` are inspection panels, not gameplay UI.

The current target readmodel supports non-player target telemetry for asteroids, bullets, pickups, and enemies where matching state maps exist. Player-only commands still require a player target.

`Game Target` is a compact selector affordance, not a separate server entity. It resolves to the current canonical player target only when the local synced ship state reports `target_kind=player`.

`All Players` is represented as `target_scope=all_players`. It should not be serialized or stored as a real player ID.
