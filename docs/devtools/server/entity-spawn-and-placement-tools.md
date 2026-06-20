# Entity Spawn And Placement Tools

Parent index: [Server](./!README.md)

## Purpose

This document describes the server-side devtools entity spawn and placement command handling.

It covers the authoritative server behavior behind debug placement requests for players, asteroids, bullets, and pickups. Client-side mouse placement, visual position conversion, target readouts, and devtools window presentation are covered by client devtools documentation.

## Overview

Entity spawn and placement tools let a developer request authoritative game-state mutations from the client devtools surface.

The server accepts placement-related devtools packets, decodes them into `devtools.DebugCommand`, routes them through the server devtools command handler, and applies the requested spawn through game-owned seams.

Current spawn surfaces:

```text
debug_spawn_entity
-> player
-> asteroid
-> bullet

debug_spawn_pickup
-> pickup
```

The server owns whether a request is valid, which entity is created, which ID is assigned, which game store receives the entity, and how the spawned entity appears in later state packets.

The client owns only the request UI and placement input. It does not create authoritative entities locally.

## Debug-only scope

Entity spawn and placement tools are development tooling.

They are not player-facing gameplay features, not normal game-mode rules, and not production progression systems.

Debug spawn tools may create or mutate authoritative runtime state, but they must still route through server-owned gameplay seams. They must not introduce a parallel debug-only entity model.

Covered debug-only behavior:

```text
spawn player at requested placement position
spawn asteroid at requested placement position
spawn bullet at requested placement position
spawn pickup at requested placement position
use optional placement direction for asteroid and bullet spawning
use optional target_player_id as a debug player-slot request for player spawning
```

Not owned by this document:

```text
normal timed asteroid spawning
asteroid fragment spawning
normal player respawn requests
continuous bullet stream runtime cadence
client placement input flow
client placement preview or readout behavior
pickup collection effects
drop-table spawning
```

## Server authority

The server remains authoritative for every spawn result.

The command path is:

```text
client devtools packet
-> websocket inbound routing
-> inbound.HandlePlacementDevtoolsPacket
-> packetcodec.Decode(raw message, devtools.DebugCommand)
-> devtools.HandleCommand
-> spawn-specific devtools handler
-> game-owned devtools export seam or game runtime method
-> normal state packet readback
```

Placement packets require a current room and a current active game player before command handling is invoked. If either is missing, the packet is consumed by devtools routing and no spawn is applied.

Spawn handlers return `true` when they recognize and consume a command, even when the requested spawn is ignored. Ignored spawn requests are logged rather than converted into client-facing failure packets.

## Client presentation

Server spawn tools do not own client presentation.

The client may expose hotkeys, devtools window buttons, player selectors, pickup selectors, and click placement flows. Those controls send packets only.

The server does not send a dedicated spawn confirmation packet. Spawn confirmation is implicit:

```text
command accepted by server
-> authoritative entity store changes
-> later state packet includes the entity
-> client world sync renders the entity
```

If a spawn request is ignored, the current server behavior is logging only. The client should not infer success from packet send completion.

## Commands or controls

### `debug_spawn_entity`

`debug_spawn_entity` is the shared packet for player, asteroid, and bullet placement.

Relevant command fields:

```text
type
entity_type
x
y
has_direction
direction_x
direction_y
target_player_id
```

Supported `entity_type` values:

```text
player
asteroid
bullet
```

Unknown entity types are consumed and logged as not implemented.

### `debug_spawn_pickup`

`debug_spawn_pickup` is the pickup placement packet.

Relevant command fields:

```text
type
pickup_type
x
y
```

The server converts `pickup_type` into a pickup definition through the pickup registry. Unknown pickup types are ignored and logged.

Current known pickup definitions include:

```text
1_up
torpedo
```

The client pickup selector may discover presentation entries, but the server remains the authority for whether the requested pickup type exists.

## Spawn behavior

### Player spawn

Debug player spawn uses `entity_type = "player"`.

The server builds a `SpawnEntityRequest`, normalizes the requested position into world space, resolves a debug player ID, ensures a player session exists, and creates a player ship.

Player ID behavior:

```text
target_player_id present
-> normalize requested ID as player-N
-> reject invalid IDs
-> reject occupied IDs
-> reserve the requested ID
-> spawn that player slot

target_player_id absent
-> allocate the first available player-N ID
-> respect playerids.MaxPlayers
-> reserve the allocated ID
-> spawn that player slot
```

The `target_player_id` field is a devtools player-slot request in this command. It is not the generic gameplay target model.

Player spawn uses:

```text
DevtoolsEnsurePlayerSession
DevtoolsSpawnPlayerShip
DummyPlayerCameraConfig
```

Creating the ship also ensures the spawned player has a camera view.

### Asteroid spawn

Debug asteroid spawn uses `entity_type = "asteroid"`.

The server creates a debug asteroid spawn plan through the normal asteroid spawning model:

```text
position
velocity
size
variant
reason
```

Asteroid spawn behavior:

```text
position = normalized requested position
direction = requested direction when valid, otherwise random unit vector
speed = random asteroid speed from the game spawner
size = random value from 1 through 4
variant = RandomDebugSpawnVariantIndex()
reason = SpawnReasonDebugAsteroid
```

The server then applies the plan through the same game-owned asteroid spawn application path used by other asteroid spawn sources.

Debug asteroid spawning uses the asteroid variant catalog’s debug spawn weighting. It should not hardcode a variant count or choose variants independently from the catalog helpers.

### Bullet spawn

Debug bullet spawn uses `entity_type = "bullet"`.

The requesting player is used as the bullet owner. If there is no owner player ID, the spawn is ignored.

Bullet spawn behavior:

```text
position = normalized requested position
direction = requested direction when valid, otherwise random unit vector
speed = BasicCannonProjectileSpeed
lifetime = BasicCannonProjectileLifetime
rotation = derived from normalized direction
id = next bullet ID from the game spawner
store = game.entities.Projectiles
```

One-shot debug bullet spawn is separate from persistent continuous bullet streams.

### Pickup spawn

Pickup spawn uses `debug_spawn_pickup`.

The server normalizes the requested position, resolves the pickup definition, creates a pickup ID, and stores the pickup in the authoritative pickup map.

Pickup spawn behavior:

```text
pickup_type = requested pickup type
definition = pickups.DefinitionFor(pickup_type)
id = pickup_<nextPickupID>
position = normalized requested position
health = definition health
lifespan = definition lifespan
store = game.entities.Pickups
```

Unknown pickup types return no entity and no gameplay mutation.

## Placement request data

Placement request helpers live in the server devtools package.

`SpawnEntityRequest` owns the server-side interpretation of entity placement data:

```text
EntityType
X
Y
HasDirection
DirectionX
DirectionY
TargetPlayerID
```

`Position()` converts request coordinates to a physics vector.

`DirectionOr(fallback)` returns a normalized requested direction when `has_direction` is true and the requested direction is nonzero. Otherwise it returns the normalized fallback direction.

This means client drag direction is advisory input. The server still normalizes it and supplies fallback direction when needed.

## Build/runtime gates

Placement spawn handlers are downstream command handlers. They do not own the full devtools build gate.

Runtime gates before spawn execution:

```text
packet type must be classified as a placement devtools packet
session must have a current room
session must have a current active game player ID
packet must decode as devtools.DebugCommand
devtools.HandleCommand must recognize the command type
spawn-specific validation must pass
```

Build-tag files define devtools enabled state:

```text
enabled_default.go
enabled_nodevtools.go
disabled.go
```

The local spawn handlers do not perform an independent `Enabled()` check. Command-routing and build-gate behavior belongs to the broader server devtools command-routing surface.

## Relationship to real gameplay implementation areas

Entity spawn and placement tools intentionally reuse real server implementation areas.

Player spawn routes through game-owned player session and ship creation seams. It does not construct a separate debug player entity type.

Asteroid spawn routes through `spawning.AsteroidSpawnPlan` and the normal asteroid application path. Debug spawn only changes the spawn reason and placement source.

Bullet spawn routes through the game’s debug bullet helper, but still creates a normal runtime bullet and stores it in the authoritative projectile map.

Pickup spawn routes through the game pickup spawn path and pickup definition registry.

This preserves the main invariant:

```text
devtools may request a spawn, but gameplay-owned server code owns the resulting entity.
```

## Telemetry

Spawn and placement commands use server logs and normal state readback.

Current logs include:

```text
debug player spawned
debug player spawn ignored
debug asteroid spawned
debug asteroid spawn ignored
debug bullet spawned
debug bullet spawn ignored
debug pickup spawned
debug pickup spawn ignored
debug spawn entity not implemented for entity type
```

Spawn logs include relevant request and result fields such as player ID, spawned entity ID, owner player ID, pickup type, coordinates, and direction presence.

There is no dedicated spawn acknowledgement packet, error packet, or client-visible server response for ignored placement requests.

## Code map

Primary server devtools spawn files:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/placement_requests.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/packets_generated.go
```

Inbound routing:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
```

Game-owned devtools seams:

```text
services/game-server/internal/game/export_devtools_spawn.go
services/game-server/internal/game/export_devtools_player_spawn.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/pickups.go
```

Related gameplay implementation:

```text
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/entities/pickups/definitions.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/space/
services/game-server/internal/game/physics/
```

Packet source and generated outputs:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Build/runtime gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/command_types.go
```

Important non-ownership boundaries:

```text
client/scripts/devtools/
client/scripts/gameplay/devtools/
client/scripts/gameplay/input/
client/scripts/world/
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/streamruntime/
docs/protocol/
docs/data/
```

## Tests

Relevant server tests include:

```text
services/game-server/tests/game/devtools_test.go
services/game-server/tests/networking/inbound_devtools_test.go
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/internal/game/asteroids/variants_test.go
```

Covered behavior includes:

```text
debug pickup spawn creates a pickup
debug pickup spawn uses the requested position
unknown pickup types are rejected without storing a pickup
placement devtools routing recognizes debug_spawn_entity
placement devtools routing recognizes debug_spawn_pickup
debug command type recognition includes spawn packet types
default builds report devtools enabled
nodevtools builds report devtools disabled through the devtools gate helpers
debug player ship spawn uses dummy camera config
debug bullet spawn requires a valid owner player ID
debug asteroid variant helpers include all current variants
```

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools
go test -buildvcs=false ./internal/game -run 'Devtools|Debug|Spawn|Pickup|Asteroid'
go test -buildvcs=false ./tests/game -run 'DebugSpawn|Devtools'
go test -buildvcs=false ./tests/networking -run 'PlacementDevtools'
go test -buildvcs=false -tags nodevtools ./internal/devtools
```

Packet drift verification:

```bash
data-sync -check -packets -go -gds
data-sync -diff -packets -go -gds
```

## Related docs

* [Server Devtools](./!README.md)
* [Devtools](../!README.md)
* [Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!README.md)
* [Game Server Networking](../../services/game-server/networking/!README.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Game Server Simulation](../../services/game-server/simulation/!README.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Asteroid Variants Data](../../data/asteroid-variants-data.md)
* [Constants](../../data/constants.md)

## Notes

The server spawn handlers consume placement requests even when the specific spawn cannot be applied. This keeps malformed or ineligible debug requests from falling through into normal gameplay packet handling.

`target_player_id` has command-specific meaning in debug player spawn. In this context it identifies a requested debug player slot. It should not be generalized into the canonical gameplay target model.

Continuous bullet streams share placement-style input on the client, but their server runtime is a separate persistent stream system. One-shot debug bullet spawn is the only bullet-spawn behavior owned by this document.
