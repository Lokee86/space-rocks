## Debug Status Output

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the game-server debug status output surface.

It covers how the server builds and emits `debug_status` packets for devtools clients, what state the packet exposes, which runtime gates control emission, and which code owns the server-side projection.

## Overview

Debug status output is a server-authored devtools telemetry packet.

It reports the current state of server-owned debug controls so the client devtools window can display accurate toggle status and per-player feature state. It does not mutate gameplay, does not replace gameplay state packets, and does not act as a command response.

The server emits debug status over the normal WebSocket connection after gameplay presentation state has been written. The status packet is built from the authoritative game instance and encoded through the shared packet codec.

Current output shape:

```text
type: debug_status
debug_status: DebugStatus
debug_statuses: map[player_id]DebugStatus
```

`debug_status` is the status view for the receiving player. `debug_statuses` is a per-player map used by client devtools target selectors and feature-state labels.

Current `DebugStatus` fields:

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

The packet is devtools-only diagnostic output. It is separate from:

```text
gameplay state packets
debug shape catalog packets
collision-body telemetry
client-only overlays
player-facing HUD
```

## Debug-only scope

Debug status output exists only to support development and debugging.

It exposes debug-control state that helps inspect and operate the current room:

```text
player damage override state
player life override state
player suspension state
global or partial world-freeze state
```

It does not expose production gameplay policy and should not be used as player-facing UI state.

The status packet should stay focused on the status of devtools controls. Richer world, collision, shape, and entity inspection belongs in separate devtools telemetry surfaces.

## Server authority

The server owns all debug status values.

`services/game-server/internal/devtools/status.go` builds outward-facing `DebugStatus` values from the game-owned devtools status projection:

```text
devtools.StatusFor
-> game.DevtoolsStatusFor
-> devtools.DebugStatus
```

`game.DevtoolsStatusFor` reads authoritative state from the game instance:

```text
worldSimulationOptions
-> world_frozen
-> asteroids_frozen
-> bullets_frozen
-> spawning_frozen
-> collisions_frozen

player session damage/life/suspension state
-> invincible
-> infinite_lives
-> player_frozen

live player entity damage state
-> invincible
```

The devtools package does not own the underlying gameplay state. It owns the debug-facing projection and packet shape. The game package owns the actual simulation, player session state, player entity state, and world simulation options.

`StatusesForAllPlayers` builds the `debug_statuses` map by iterating match players from `MatchDecision().Players` and projecting each player through `StatusFor`.

## Output lifecycle

Debug status output is written by the WebSocket write loop.

Current flow:

```text
writeServerMessages
-> gameplay write ticker
-> BuildGameplayPresentationStateResponse
-> WriteServerMessage
-> writeDebugShapeCatalogMessage
-> debugStatusTick++
-> writeDebugStatusMessage every debugStatusWriteIntervalTicks
-> BuildDebugStatusResponse
-> packetcodec.Encode
-> WriteServerMessage
```

`debugStatusWriteIntervalTicks` is currently `8`.

The write loop sends gameplay presentation state every eligible server write tick. Debug status is slower and is attempted only after a successful gameplay state write path reaches the debug-status cadence check.

`writeDebugStatusMessage` requires a current game player id and an eligible room before building the packet.

## Send eligibility

`CanSendDebugStatus` controls whether outbound debug status is eligible.

Current requirements:

```text
room is not nil
room has a game instance
devtools.Enabled() is true
room state is InGame or GameOver
```

The WebSocket writer also requires:

```text
session.currentGamePlayerID is not empty
```

Debug status is not sent for lobby-only sessions, sessions without a current game player id, rooms without a game instance, or `nodevtools` builds.

`GameOver` remains eligible so devtools can continue showing final debug state while the room is in its resolved end-of-match state.

## Packet shape

The status packet source shape is defined in the shared packet schema.

Source:

```text
shared/packets/debug.toml
```

Generated server output:

```text
services/game-server/internal/devtools/packets_generated.go
```

Generated packet structs:

```go
type DebugStatus struct {
	Invincible       bool `json:"invincible"`
	InfiniteLives    bool `json:"infinite_lives"`
	WorldFrozen      bool `json:"world_frozen"`
	AsteroidsFrozen  bool `json:"asteroids_frozen"`
	BulletsFrozen    bool `json:"bullets_frozen"`
	SpawningFrozen   bool `json:"spawning_frozen"`
	CollisionsFrozen bool `json:"collisions_frozen"`
	PlayerFrozen     bool `json:"player_frozen"`
}

type DebugStatusPacket struct {
	Type          string                 `json:"type"`
	DebugStatus   DebugStatus            `json:"debug_status"`
	DebugStatuses map[string]DebugStatus `json:"debug_statuses"`
}
```

The server sets `Type` to:

```text
debug_status
```

The packet should remain a compact status projection. It should not absorb collision bodies, shape catalogs, raw gameplay state, or command acknowledgement payloads.

## Client presentation

The server does not own client presentation.

The client consumes `debug_status` packets through the devtools readmodel path and uses the result to refresh:

```text
world freeze status labels
per-player invincible selector labels
per-player infinite-lives selector labels
per-player freeze selector labels
devtools window status display
```

Client presentation treats `debug_status` as telemetry. It does not apply gameplay effects locally.

`debug_statuses` is used for per-player selector rows. Missing or malformed per-player status data should degrade as inactive or empty on the client side rather than creating alternate authority.

## Commands or controls

Debug status output has no direct request command.

The client does not ask the server for an immediate status snapshot. The server emits status snapshots on its own outbound cadence when eligible.

Client controls can change later status output by sending normal devtools command packets, such as:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
```

Those commands route through server devtools command handling and mutate only through game-owned seams. A later `debug_status` packet reports the resulting authoritative state.

The status output must not become a command transport, command acknowledgement format, or client-side mutation shortcut.

## Telemetry behavior

Debug status is live diagnostic telemetry, not analytics.

The packet reports current server state. It is transient and is not persisted.

Encoding failures are logged through the networking logger with room, player, and remote-address context. When encoding fails, the server skips that status write and keeps the session alive.

Current encode path:

```text
BuildDebugStatusResponse
-> packetcodec.Encode
-> logging.Network.Error on encode failure
```

Debug status output does not log every successful status packet. Routine successful writes stay quiet.

## Build/runtime gates

Server devtools are enabled in default builds and disabled by the `nodevtools` build tag.

Default build:

```text
services/game-server/internal/devtools/enabled_default.go
-> Enabled() == true
```

`nodevtools` build:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
-> Enabled() == false
```

Command routing uses:

```text
services/game-server/internal/devtools/disabled.go
-> ShouldHandleCommand(packetType)
```

Outbound status uses `devtools.Enabled()` directly through `CanSendDebugStatus`.

When `nodevtools` is active:

```text
debug commands are not handled
debug status packets are not eligible to send
```

This keeps both mutation and server-authored status output behind the same server-side devtools gate.

## Relationship to gameplay implementation

Debug status output observes real gameplay-owned state.

It must not introduce duplicate debug-only gameplay state. The devtools package may project and serialize status, but gameplay state remains in the owning game/runtime structures:

```text
DamageOptions
LifeOptions
Suspension
WorldSimulationOptions
MatchDecision player list
```

World-freeze fields are derived from `WorldSimulationOptions` capability methods rather than from independent devtools booleans.

Player freeze is derived from the player/session suspension state. It remains separate from normal pause, even though both contribute to player suspension behavior.

Invincibility is read from session/player damage options. The live player entity can override the session-derived value in the projection when present, matching the authoritative runtime state used by simulation.

## Code map

Primary server status files:

```text
services/game-server/internal/devtools/status.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/websocket_write.go
```

Game-owned status source files:

```text
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
```

Generated packet files:

```text
shared/packets/debug.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Build-gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Related command files that can affect later status output:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/inbound/router.go
```

Relevant tests:

```text
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/devtools/
-> owns debug command/status projection and generated debug packet structs

services/game-server/internal/game/
-> owns authoritative simulation and player/session state

services/game-server/internal/networking/
-> owns websocket read/write routing, cadence, and packet emission

client/scripts/devtools/
-> owns client devtools presentation and readmodels
```

## Tests and verification

`debug_status_presentation_test.go` verifies that the outbound response:

```text
encodes as JSON
uses type debug_status
includes debug_status
includes debug_statuses
does not include debug_collision_bodies
rejects nil room input
rejects rooms without a game instance
```

`enabled_default_test.go` verifies that default builds enable devtools and allow devtools command handling.

`disabled_test.go` verifies that `nodevtools` builds disable devtools and reject devtools command handling.

`toggles_test.go` verifies several command paths that affect status fields, including all-player behavior for invincibility, infinite lives, and player freeze.

Run server tests after changing:

```text
debug status packet shape
debug status send eligibility
debug status cadence
DebugStatus field projection
nodevtools gate behavior
toggle command effects that feed status output
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Devtools Design](../design/!INDEX.md)
* [Client Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)
* [Realtime WebSocket Protocol](../../protocol/stubs/realtime-websocket-protocol.md) - Realtime websocket protocol documentation.

## Notes

Debug status output and debug shape catalog output are separate outbound surfaces. Status is repeated on a slower cadence. Shape catalog output is sent once per room.

The legacy devtools notes correctly treated debug status as server-authored diagnostic state, not client authority. That rule still applies.

The status packet should remain small. New debug inspection data should be added to a separate telemetry surface when it is not directly a control-status boolean.
