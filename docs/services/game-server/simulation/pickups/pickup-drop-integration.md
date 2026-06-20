# Pickup Drop Integration

Parent index: [Game Server Simulation Pickups](./!INDEX.md)

## Purpose

This document describes how the game server integrates pickup drops into asteroid destruction.

It explains where drop-table evaluation happens, how successful drop results become authoritative pickup entities, what events are emitted, and which boundaries remain owned by drop tables, pickup entities, pickup collection, data sync, and client presentation.

## Overview

Pickup drop integration is a game-server simulation responsibility owned by `services/game-server/internal/game`.

The current runtime flow is:

```text
projectile destroys asteroid
-> asteroid destruction consequences run
-> score is awarded
-> asteroid is marked pending despawn
-> asteroid fragments may spawn
-> basic asteroid drop table is evaluated
-> successful drop result spawns a pickup entity
-> pickup_dropped event is recorded
-> pickup appears in StatePacket.pickups
-> pickup lifecycle, collection, and effects run through normal pickup systems
```

Drop integration sits between asteroid destruction and pickup lifecycle. It does not decide pickup effects, collection behavior, collision behavior, or client rendering. It only decides whether a destroyed asteroid produces an authoritative pickup entity.

The drop-table package evaluates pure drop rules. Root game code owns the mutation that turns a successful drop result into a stored pickup.

## Code root

```text
services/game-server/internal/game/
```

Primary supporting packages and data roots:

```text
services/game-server/internal/game/drops/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/pickups/
shared/drop_tables/
tools/data_sync/
```

## Responsibilities

Pickup drop integration owns the game-server side of:

* Triggering pickup drop evaluation from asteroid destruction consequences.
* Looking up the active drop table used for destroyed asteroids.
* Respecting the drop table active-pickup cap before and during drop spawning.
* Building a drop source from the destroyed asteroid.
* Supplying random roll values to the drop-table evaluator.
* Converting successful drop results into authoritative pickup entities.
* Recording `pickup_dropped` events after successful pickup spawn.
* Leaving spawned pickups in the normal runtime pickup store for lifecycle, state projection, collision, collection, and effects.

## Does not own

Pickup drop integration does not own:

* Drop-table source-of-truth TOML parsing.
* Drop-table Go generation.
* Pure drop-table roll evaluation rules.
* Pickup entity definition data.
* Pickup collision shapes.
* Pickup age, lifespan, expiry, or removal outside spawn failure handling.
* Player/pickup collision detection.
* Pickup collection rules.
* Pickup effect intent resolution.
* Player lives, weapon equip, or ammo mutation from collected pickups.
* Client pickup rendering.
* Client audio, effects, HUD, telemetry, or readout presentation.
* WebSocket transport or packet codec behavior.
* Durable player profile or account persistence.

Those systems participate after a pickup exists, but they own their own boundaries.

## Domain roles

Pickup drop integration participates in the combat and reward loop by allowing destroyed asteroids to produce temporary pickups.

The server is authoritative for:

* whether an asteroid was destroyed
* whether a drop table is evaluated
* whether a drop succeeds
* which pickup type is spawned
* where the pickup spawns
* whether active pickup caps allow the spawn
* when `pickup_dropped` is recorded

The client observes the result. It does not roll drop tables, instantiate authoritative pickups locally, or decide whether a dropped pickup exists.

## Asteroid destruction integration

Pickup drops currently run from projectile-caused asteroid destruction.

The destruction consequence path is:

```text
applyProjectileAsteroidDestruction
-> evaluate scoring policy
-> award score through game-owned score mutation
-> mark asteroid pending despawn
-> spawn asteroid fragments
-> maybeDropPickupFromAsteroidLocked
```

Pickup drops only occur after a projectile damage result destroys an asteroid. Nonfatal projectile hits do not drop pickups. Ship/asteroid collision does not drop pickups.

The drop helper runs while the game lock is held. It mutates game-owned runtime state only after the drop evaluator returns one or more successful results.

## Drop table selection

The current implementation uses the generated `basicasteroids` table directly:

```text
table id = basicasteroids
source type = asteroid
drop mode = single
max drops per source = 1
max active pickups = 2
```

The table source is:

```text
shared/drop_tables/basicasteroids.toml
```

The generated server output is:

```text
services/game-server/internal/game/drops/drop_tables.go
```

`Game.New` installs generated drop tables into the game instance:

```go
dropTables: drops.GeneratedTables
```

`maybeDropPickupFromAsteroidLocked` looks up `basicasteroids` from `game.dropTables.ByID`. If the table is missing, the helper returns without spawning or recording an event.

## Current drop entries

The current generated `basicasteroids` table contains these entries:

| Pickup type | Chance | Source size range |
| ----------- | -----: | ----------------- |
| `1_up`      | `0.01` | `1` through `4`   |
| `torpedo`   | `0.15` | `2` through `4`   |

The table uses `single` mode. Entries are evaluated in order, and evaluation stops after the first successful matching entry.

For size `2` through `4` asteroids, `1_up` is checked first. If it succeeds, `torpedo` is not evaluated. If it fails, `torpedo` may be evaluated. For size `1` asteroids, only `1_up` is in range.

There is no minimum drop count. A destroyed asteroid can produce no pickup.

## Drop evaluation model

The `drops` package owns pure evaluation.

It receives:

```text
table id
drop source
roll values
```

The source built for asteroid drops contains:

```text
Type = asteroid
ID   = asteroid.ID
Size = asteroid.Size
X    = asteroid.X
Y    = asteroid.Y
```

The evaluator rejects:

* missing table IDs
* source type mismatches
* entries outside the source size range
* failed chance rolls

A roll succeeds when the roll value is lower than the entry chance.

The evaluator returns `drops.Result` values containing:

```text
table id
pickup type
x
y
```

The evaluator does not spawn pickups, inspect game state, mutate runtime maps, record events, or talk to clients.

## Active pickup cap

`MaxActivePickups` is enforced in game-owned drop integration.

The helper checks the active pickup count before rolling the table:

```text
if active pickups >= table max active pickups
-> return without rolling or spawning
```

It checks again before spawning each returned result:

```text
for each result
-> if active pickups >= table max active pickups, return
-> spawn pickup
```

This matters for future multi-drop tables. A table can produce multiple results, but game-owned integration still prevents the active pickup count from exceeding the table cap.

The current `basicasteroids` table has:

```text
max_active_pickups = 2
```

## Pickup spawning

Successful drop results are converted into authoritative pickups through:

```go
game.spawnPickupLocked(...)
```

That spawn helper:

* validates the pickup type through pickup definitions
* allocates a stable pickup id such as `pickup_1`
* initializes type, position, health, age, and lifespan
* stores the pickup in `game.entities.Pickups`

Spawn position comes from the drop result, which currently uses the destroyed asteroid position.

If the result pickup type is unknown to pickup definitions, spawn fails and no `pickup_dropped` event is recorded for that result.

## Event recording

A successful drop spawn records:

```text
pickup_dropped
```

The event includes:

```text
pickup_id
pickup_type
source_type
source_id
table_id
x
y
```

`pickup_dropped` means the server successfully created a pickup from a drop table result. It does not mean the pickup was collected, consumed, expired, or applied to a player.

Pickup drop events are separate from:

```text
pickup_collected
pickup_effect_applied
pickup_expired
```

No event is recorded for:

* missing drop table
* failed roll
* source size mismatch
* active pickup cap blocking spawn
* unknown pickup type preventing spawn

## Protocol and APIs

Pickup drop integration has no inbound HTTP or WebSocket API.

Clients observe drop results through normal game-server state output:

```text
StatePacket.pickups
StatePacket.events
```

A dropped pickup appears in `StatePacket.pickups` as a normal pickup state with:

```text
id
type
pickup_class
x
y
health
age_seconds
lifespan_seconds
```

The `pickup_dropped` event is emitted through the normal event projection path. It is presentation data for clients and tooling. It is not a client authority surface.

The client may render the pickup and react to the event, but the server remains authoritative for the pickup entity, lifetime, collection, and effect application.

## Relationship to pickup lifecycle and effects

The drop seam ends once a pickup has been spawned and the drop event has been recorded.

After that point, normal pickup systems own behavior:

```text
stepPickups
-> age and expiry

handlePlayerPickupCollisions
-> collision and collection

pickups.ResolveCollection
-> collection result and effect intent

applyPickupEffectIntentLocked
-> player/session/weapon mutation

pickupStatesLocked
-> state packet projection
```

This separation preserves a narrow responsibility split:

```text
drops package
= decide whether a source produces a pickup result

game pickup drop integration
= spawn authoritative pickup from successful result

pickup entity lifecycle
= age, expire, and project pickup state

pickup collection/effects
= collect pickup and apply gameplay mutation
```

## Data ownership

Pickup drop integration reads generated drop-table data at runtime.

Source-of-truth data lives in:

```text
shared/drop_tables/basicasteroids.toml
```

Generated runtime data lives in:

```text
services/game-server/internal/game/drops/drop_tables.go
```

The game server stores generated tables in memory on `Game.dropTables`.

Pickup drop integration mutates:

```text
game.nextPickupID
game.entities.Pickups
game.pendingPresentationEvents
```

It reads:

```text
game.dropTables
game.entities.Pickups
destroyed asteroid id
destroyed asteroid size
destroyed asteroid position
pickup definitions
```

It does not persist account, profile, match-result, or player-data state.

## Invariants

Pickup drop integration must preserve these rules:

* Drop tables are server-authoritative.
* Drop-table evaluation is pure and does not mutate game state.
* Root game code owns spawning pickups from drop results.
* Pickup drops happen only after asteroid destruction consequences reach the drop helper.
* Nonfatal asteroid damage must not drop pickups.
* The active pickup cap must be checked before spawning.
* A successful drop must create an authoritative pickup before emitting `pickup_dropped`.
* `pickup_dropped` must not stand in for collection or effect application.
* Pickup collection and pickup effects remain owned by the pickup seam.
* The client must observe pickup drops through server state and events, not local authority.
* Drop-table source changes must flow through data sync before they affect generated server runtime tables.

## Code map

Primary integration files:

```text
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/pickup_drops.go
services/game-server/internal/game/pickups.go
services/game-server/internal/game/events.go
services/game-server/internal/game/events/events.go
services/game-server/internal/game/state_packet.go
```

Drop-table package:

```text
services/game-server/internal/game/drops/table.go
services/game-server/internal/game/drops/source.go
services/game-server/internal/game/drops/roll.go
services/game-server/internal/game/drops/drop_tables.go
```

Pickup entity and effect support:

```text
services/game-server/internal/game/entities/pickups/types.go
services/game-server/internal/game/entities/pickups/definitions.go
services/game-server/internal/game/entities/pickups/pickup.go
services/game-server/internal/game/pickup_lifecycle.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/pickup_effects.go
services/game-server/internal/game/pickups/collection.go
```

Source and generated data:

```text
shared/drop_tables/basicasteroids.toml
tools/data_sync/config.toml
tools/data_sync/data_sync/drop_tables_toml.py
tools/data_sync/data_sync/drop_tables_sync.py
tools/data_sync/data_sync/generators/go_drop_tables.py
tools/data_sync/data_sync/model/drop_tables.py
```

Relevant tests:

```text
services/game-server/internal/game/pickup_drops_test.go
services/game-server/internal/game/drops/table_test.go
services/game-server/internal/game/drops/drop_tables_test.go
services/game-server/tests/game/pickups_test.go
tools/data_sync/tests/test_drop_tables_toml.py
tools/data_sync/tests/test_drop_tables_generators.py
tools/data_sync/tests/test_drop_tables_sync.py
tools/data_sync/tests/test_final_flows.py
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/pickups/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/drops/
services/game-server/internal/networking/
client/
shared/packets/
shared/constants/
```

`drops` owns table evaluation only.

`entities/pickups` owns pickup entity definitions and collision body construction.

`game/pickups` owns collection result and effect intent rules.

`networking` owns transport only.

`client` owns presentation only.

`shared/packets` owns generated packet structure, not runtime drop decisions.

`shared/constants` owns generated pickup definition constants, not drop-table evaluation.

## Tests and verification

Current test coverage includes:

* generated drop tables are well formed
* missing table IDs return no drop result
* source type mismatches return no drop result
* source size mismatches return no drop result
* rolls that meet or exceed chance return no result
* rolls below chance return drop results
* single mode returns the first successful matching entry
* multi mode can return multiple results
* multi mode respects `MaxDropsPerSource`
* asteroid drop integration creates a pickup
* asteroid drop integration respects `MaxActivePickups`
* failed drop chance creates no pickup
* dropped pickups project into state packets
* projectile asteroid destruction can trigger pickup drops
* spawned pickups use definitions, lifecycle, state projection, collection, and expiry paths

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/drops
go test -buildvcs=false ./internal/game -run 'Drop|Pickup|AsteroidDestruction'
go test -buildvcs=false ./...
```

Useful data-sync verification commands:

```bash
data-sync -check -drop-tables -go
data-sync -diff -drop-tables -go
data-sync -validate -drop-tables
```

## Related docs

* [Game Server Simulation Pickups](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Simulation Combat](../combat/!INDEX.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Game Server Simulation World](../world/!INDEX.md)
* [Asteroid Spawning And Variants](../world/asteroid-spawning-and-variants.md)
* [Game Server Simulation Runtime](../runtime/!INDEX.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Pickup Entity Lifecycle](pickup-entity-lifecycle.md)
* [Pickup Collection](pickup-collection.md)
* [Pickup Effects](pickup-effects.md)
* [Drop Table Pipeline](../../../../data/stubs/drop-table-pipeline.md)
* [Data](../../../../data/!INDEX.md)
* [Protocol](../../../../protocol/!INDEX.md)
* [Systems Design Combat](../../../../systems-design/combat/!INDEX.md)
* [Client World Sync](../../../client/world-sync/!INDEX.md)
* [Pickup Presentation](../../../client/world-sync/pickup-presentation.md)
* [Data Sync](../../../../../tools/data_sync/!INDEX.md)

## Notes

Asteroid variant source data includes a `drop_table` field, and generated server asteroid variants retain that field. Current pickup drop integration does not yet select the table from the runtime asteroid variant. It uses `basicasteroids` directly in `maybeDropPickupFromAsteroidLocked`.

The current `pickup_dropped` event projection records the table id as `basicasteroids` from the integration helper. The drop result also carries a table id, but the current event construction does not read it from the result.

The current drop-table generated output is Go-only. Drop tables are not generated into client constants or packet outputs.
