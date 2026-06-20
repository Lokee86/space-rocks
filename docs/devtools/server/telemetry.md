## Telemetry

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes server-side devtools telemetry for Space Rocks.

It covers the diagnostic packet surfaces the game server emits or answers, the server-owned runtime facts those packets expose, how devtools telemetry stays separate from gameplay mutation, and which implementation paths own the current behavior.

## Overview

Server devtools telemetry means live diagnostic visibility, not analytics, tracing, or durable metrics storage.

The server currently participates in four telemetry-facing surfaces:

```text
debug_status
  periodic server devtools status snapshot

debug_shape_catalog
  room-scoped diagnostic collision shape catalog snapshot

telemetry_ping / telemetry_pong
  client/server timing diagnostic packet pair

state.server_sent_msec
  server timestamp on normal gameplay presentation state packets
```

These surfaces use the normal realtime packet codec and WebSocket writer path. They do not create a separate debug transport.

The high-level server flow is:

```text
game/runtime state
-> game export seam or packet projection
-> devtools/outbound networking packet builder
-> packetcodec.Encode
-> websocket writer
-> client devtools consumer
```

For ping/pong timing diagnostics, the flow is inbound first:

```text
client telemetry_ping
-> websocket packet decode
-> inbound telemetry route
-> telemetry_pong packet
-> same session outbound channel
```

Server telemetry observes runtime state, exposes diagnostic packets, and responds to timing probes. It does not own client overlay rendering, devtools-window layout, HUD behavior, durable player data, or analytics storage.

## Debug-only scope

Server devtools telemetry is development tooling.

It is allowed to expose:

```text
debug toggle state
per-player debug status maps
world-freeze sub-state flags
collision shape catalog definitions
server receive/send timestamps for timing probes
gameplay state send timestamps
```

It is not allowed to:

```text
replace gameplay state authority
store telemetry history
publish analytics
mutate gameplay by being observed
create a parallel debug simulation path
bypass the realtime packet codec
bypass normal WebSocket session routing
```

The server may expose facts that client devtools present as labels, overlays, or raw readouts. The server does not own those presentation surfaces.

## Server authority

The server owns the facts emitted by server devtools telemetry.

`debug_status` comes from server runtime state through the devtools export seam:

```text
devtools.StatusFor
-> game.DevtoolsStatusFor
-> DebugStatus
```

The status snapshot reads:

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

The player-specific flags are read from the server-owned player session or active player instance. World and subsystem freeze flags are read from `worldSimulationOptions`.

`debug_statuses` is a map keyed by match player ID. It is built from the current match decision player list and calls the same status projection for each player.

`debug_shape_catalog` comes from the server collision-shape catalog. The server loads imported collision shape definitions, converts them to collision bodies, derives outline points, and emits shape definitions keyed by stable devtools shape IDs.

Shape catalog entries include:

```text
id
kind
shape_type
points
```

Current shape kinds include:

```text
player
asteroid
bullet
pickup
```

`telemetry_ping` and `telemetry_pong` are timing diagnostics only. The server preserves the client ping sequence and client send timestamp, then adds server receive and server send timestamps. The packet pair does not mutate room, player, or simulation state.

Normal gameplay `state` packets are also stamped with `server_sent_msec` before outbound encoding. Client telemetry uses that timestamp together with ping/pong-derived clock offset estimates to calculate packet age.

## Client presentation

The client consumes server telemetry, but the server does not own presentation.

Current client consumers include:

```text
world telemetry overlay
network player labels
devtools window debug status labels
devtools window target selectors
raw local player telemetry
raw target telemetry
server hitbox or shape debug presentation
```

The server-facing distinction is:

```text
server owns emitted diagnostic facts
client owns rendering, layout, hotkeys, labels, and readout formatting
```

The world telemetry overlay consumes `telemetry_pong` and gameplay state timing data. It also counts server-owned state dictionaries after the client normalizes gameplay state.

The devtools window consumes `debug_status` and `debug_statuses` for current debug toggle status and per-player selector labels.

The shape and hitbox debug surfaces consume server-owned shape or collision-body facts. Shape catalog output is a separate packet from gameplay state and debug status.

## Commands or controls

Telemetry output itself is mostly passive.

The current controls and triggers are:

```text
telemetry_ping
  client sends while the world telemetry overlay is visible;
  server responds with telemetry_pong to the same session

debug_status
  server sends periodically from the websocket write loop when eligible

debug_shape_catalog
  server sends once per room session when eligible

state.server_sent_msec
  server stamps every outgoing gameplay presentation state packet
```

Debug commands can change facts later reported by telemetry, but telemetry is not the command path. For example, toggling invincibility changes server runtime state through the devtools command handler; a later `debug_status` packet reports the new `invincible` value.

`debug_status` is emitted on a slower cadence than gameplay state. The write loop sends gameplay presentation state on the server write tick and sends debug status every `debugStatusWriteIntervalTicks`, currently `8`.

`debug_shape_catalog` is sent once per room ID for a session after gameplay presentation begins. The write loop tracks the last room ID used for shape catalog output and does not resend the catalog for the same room unless the tracked room changes.

## Telemetry surfaces

### Gameplay state timestamp

Gameplay presentation state is not a devtools packet, but it carries timing data used by devtools telemetry.

Before encoding an outbound gameplay state packet, the server sets:

```text
server_sent_msec
```

The timestamp is generated in the outbound networking path immediately before packet encoding.

The state packet also contains world data that client telemetry can count:

```text
players
player_sessions
player_lifecycle
bullets
asteroids
pickups
total_asteroids
events
```

Client overlays may count these dictionaries, but the server treats them as normal gameplay state projection, not a separate telemetry packet.

### Telemetry ping and pong

`telemetry_ping` is decoded as a normal client packet after the early devtools envelope routes and auth route.

The server handles only packets with:

```text
type = telemetry_ping
```

The response uses:

```text
type = telemetry_pong
sequence
client_sent_msec
server_received_msec
server_sent_msec
```

`sequence` and `client_sent_msec` are copied from the ping. `server_received_msec` is captured when the server begins handling the ping. `server_sent_msec` is captured immediately before encoding the pong.

The response is written to the same WebSocket session outbound channel. It is not broadcast to the room.

### Debug status

`debug_status` is an outbound devtools packet.

The packet shape is:

```text
type = debug_status
debug_status = DebugStatus for the receiving/current player
debug_statuses = map[player_id]DebugStatus for all match players
```

The status fields are:

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

Eligibility requires:

```text
room exists
room has a game instance
server devtools are enabled
room state is InGame or GameOver
session has a current game player ID
```

The packet does not include entity maps, collision bodies, gameplay state, or shape catalog data.

### Debug shape catalog

`debug_shape_catalog` is an outbound devtools packet.

The packet shape is:

```text
type = debug_shape_catalog
shapes = map[shape_id]DebugShapeDefinition
```

Each shape definition contains:

```text
id
kind
shape_type
points
```

The server builds the catalog from the imported collision shape catalog. Shape IDs are constructed by server devtools helpers:

```text
player:<ship_type>
asteroid:<variant>
bullet
pickup:<pickup_type>
```

Eligibility requires:

```text
room exists
room has a game instance
server devtools are enabled
room state is InGame or GameOver
```

The packet is catalog data only. It does not include live players, asteroids, bullets, pickups, or collision-body instances.

### Collision body telemetry seam

The game aggregate exposes a devtools collision body snapshot seam:

```text
Game.DevtoolsCollisionBodies()
```

That method reads the authoritative entity store under the game lock and derives collision body outline points for:

```text
players
asteroids
bullets
pickups
```

Each body contains:

```text
kind
id
shape
points
```

This is a server-owned observation seam over real collision bodies. It should remain tied to gameplay collision data and should not become a parallel client-only shape model.

Current outbound tests assert that `debug_status`, `debug_shape_catalog`, and normal gameplay presentation state do not inline `debug_collision_bodies`. Collision bodies are a distinct diagnostic seam, not hidden payload inside unrelated packets.

### Player counters and status readout

Score and lives counter changes are command behavior, not telemetry behavior.

When devtools commands set or add score/lives, the server mutates the authoritative player session through game export seams. The resulting facts are visible through normal gameplay state and related client readmodels.

There is no separate server telemetry packet dedicated only to score/lives counter output.

## Build/runtime gates

Server devtools are enabled in default builds and disabled with the `nodevtools` build tag.

Current build-gate files are:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Default build behavior:

```text
devtools.Enabled() == true
```

`nodevtools` behavior:

```text
devtools.Enabled() == false
```

Outbound `debug_status` and `debug_shape_catalog` eligibility checks require `devtools.Enabled()`.

`telemetry_ping` and `telemetry_pong` are networking timing diagnostics, not devtools command packets. Their server route is handled under inbound networking telemetry, not under the devtools command handler.

Runtime gates also apply:

```text
debug_status requires a current game player id
debug_status requires InGame or GameOver room state
debug_shape_catalog requires InGame or GameOver room state
debug_shape_catalog requires a game instance
telemetry_pong requires only a valid telemetry_ping packet on a connected session
```

## Code map

Primary server devtools telemetry files:

```text
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Game export seams used by telemetry:

```text
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/game/export_devtools_collision_telemetry.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/export_devtools_player_counters.go
```

Networking inbound telemetry path:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/telemetry.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
```

Networking outbound telemetry path:

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

Shared packet sources and generated output:

```text
shared/packets/debug.toml
shared/packets/gameplay.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/game/packets.go
```

Collision shape sources:

```text
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_outline.go
services/game-server/internal/game/physics/collision.go
```

Important non-ownership boundaries:

```text
client/scripts/devtools/
  owns client presentation, overlay flow, labels, and raw readout formatting

services/game-server/internal/networking/
  owns WebSocket decode, routing, timestamped responses, and writes

services/game-server/internal/game/
  owns authoritative simulation state and export seams

services/player-data/
  does not own server devtools telemetry

docs/services/game-server/networking/
  owns broader realtime packet-routing documentation

docs/data/
  owns packet source-of-truth and generated output documentation
```

## Tests and verification

Focused server tests include:

```text
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
```

Related networking integration tests include:

```text
services/game-server/tests/networking/rooms_test.go
```

Relevant verification areas:

```text
debug_status packet includes status payloads
debug_status packet excludes collision-body payloads
debug_shape_catalog packet includes only shape catalog payload
debug_shape_catalog packet excludes live entity payloads
shape catalog generation produces usable outline points
server devtools build gates enable or disable debug command/status behavior
collision body telemetry uses server collision bodies
gameplay presentation state remains separate from devtools packets
```

Run game-server tests after changing telemetry packet shapes, status projection, shape catalog generation, WebSocket write cadence, inbound telemetry routing, build gates, or game export seams used by devtools.

Run packet generation checks after changing `shared/packets/debug.toml` or `shared/packets/gameplay.toml`.

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Telemetry Overlays](../client/telemetry-overlays.md)
* [Client Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Game Server Telemetry And Packet Routing](../../services/game-server/networking/telemetry-packet-routing.md)
* [Game Server Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
* [Game Server Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Game Server State Packet Projection](../../services/game-server/simulation/runtime/state-packet-projection.md)
* [Game Server Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)
* [Realtime Protocol](../../protocol/realtime-websocket-protocol.md)

## Notes

Telemetry in this document means live debug and diagnostic readouts. It does not mean production analytics.

`debug_status`, `debug_shape_catalog`, and gameplay `state` are intentionally separate packet surfaces. Tests should preserve that separation unless the packet contract is deliberately changed.

`packet_staleness_ms` and `packet_age_ms` are client-side calculations. The server supplies timestamps and state packets; the client owns the derived timing readout.

The server collision body telemetry seam observes real collision bodies. It should stay connected to the authoritative physics/collision implementation rather than duplicating shape facts in client-only debug logic.
