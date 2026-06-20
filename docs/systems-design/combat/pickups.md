## Pickups

Parent index: [Combat](./!README.md)

## Purpose

This document describes the combat systems-design model for pickups.

It defines pickup authority, collection semantics, effect boundaries, drop integration, presentation boundaries, and invariants that should remain true as pickup types, loadouts, inventory, rewards, and future combat mechanics expand.

## Overview

Pickups are temporary, server-authoritative combat/reward entities that can exist in the live world, be collected by players, and apply runtime gameplay effects.

The current implemented pickup loop is:

```text
combat source destroys or creates something
-> optional drop/spawn path creates authoritative pickup
-> pickup exists in server world state
-> server advances pickup age and expiry
-> client presents pickup from state packets
-> player/pickup collision consumes pickup
-> server resolves pickup effect intent
-> server applies gameplay mutation if valid
-> server emits collection/effect presentation events
```

Pickups are not durable inventory by default. They are match-runtime entities. A pickup may change match-local session state, such as lives, or active runtime combat state, such as an equipped limited-ammo secondary weapon, but those changes do not create account ownership, profile inventory, or hangar records unless a separate durable grant flow is explicitly implemented.

Current implemented pickup types are:

```text
1_up    -> powerup pickup class -> add_lives effect
torpedo -> weapon pickup class  -> equip secondary torpedo and add ammo
```

The pickup system deliberately separates four concepts:

```text
pickup entity
= the authoritative world object that exists, ages, expires, and can be collected

pickup class
= broad family used for collision/presentation grouping, such as powerup or weapon

pickup type
= gameplay identity, such as 1_up or torpedo

pickup effect
= server-owned gameplay mutation attempted after collection
```

This separation lets new pickup families, visual presentation, collision shapes, and effect rules evolve without turning every pickup type into a separate entity system.

## Conceptual model

A pickup has two identities at the same time:

```text
class identity
type identity
```

Class identity is broad and structural. It answers which pickup family the entity belongs to. Current class identities include:

```text
powerup
weapon
```

Type identity is specific and behavioral. It answers what the pickup is. Current type identities include:

```text
1_up
torpedo
```

The pickup class is used for generic family concerns, such as class-level collision shape lookup and client scene-family selection. The pickup type is used for gameplay effect resolution and badge/icon identity.

A pickup lifecycle has three conceptual phases:

```text
creation
-> active world existence
-> removal
```

Creation can come from normal gameplay, such as asteroid drop integration, or from devtools spawn commands. After creation, all pickup sources converge on the same authoritative pickup entity model.

Active world existence means the pickup appears in server state packets with identity, position, health, age, and lifespan data. Clients render this state but do not own it.

Removal currently happens by collection or expiry:

```text
player/pickup collision
-> pickup_collected
-> effect intent application path

age reaches lifespan
-> pickup_expired
```

Collection and effect application are separate concepts. Collection means the pickup entity was consumed and removed. Effect application means the server successfully mutated gameplay state after resolving the collected pickup type.

## Authority rules

The game server is authoritative for:

* pickup creation
* pickup IDs
* pickup position
* pickup type and class
* pickup age and lifespan
* pickup expiry
* pickup collision and collection
* pickup removal
* pickup effect intent resolution
* pickup effect application
* pickup-related gameplay events

The client is authoritative for none of those outcomes.

The client owns presentation only:

* rendering active pickup nodes
* selecting local pickup scene families from `pickup_class`
* showing badge/icon state from pickup `type`
* deriving end-of-life blink from server age/lifespan values
* playing spawn and collection presentation effects
* removing local pickup nodes when the server no longer reports them

The client must not instantiate authoritative pickups locally. It must not decide that a pickup was collected. It must not apply pickup effects locally.

Drop tables are authoritative data/rule inputs, but they do not own pickup lifecycle or pickup effects. Drop-table evaluation can produce a pickup result; the game server then creates the authoritative pickup entity. Once the pickup exists, normal pickup lifecycle, collection, and effect rules own the rest of the behavior.

## Collection and effect semantics

Pickup collection is a contact outcome, not a damage outcome.

A player collects a pickup when the server detects player/pickup collision during the authoritative collision phase. Current collection skips pending-despawn players and respects the global collision phase gate.

Collection is intentionally two-stage:

```text
stage 1: collection
-> consume pickup entity
-> emit pickup_collected

stage 2: effect application
-> apply server-owned gameplay mutation if valid
-> emit pickup_effect_applied when mutation succeeds
```

This means a pickup can be collected even if its effect intent is empty or the later mutation fails. The pickup entity is already consumed once collection succeeds.

Current effects are narrow:

```text
1_up
-> increments player session lives by 1

torpedo
-> equips torpedo into the live secondary weapon slot
-> uses limited ammo
-> adds 1 secondary ammo
```

Torpedo pickup ammo is additive. Collecting a torpedo pickup while torpedo is already equipped increases secondary ammo instead of replacing the ammo count.

The `1_up` effect mutates match-local player session lives. It does not mutate ship health.

The torpedo effect mutates the active runtime ship weapon state. It does not create durable weapon ownership.

## Drop integration

Pickups can enter combat through drop-table integration.

The current normal gameplay drop source is destroyed asteroids. When projectile-caused asteroid destruction reaches the destruction consequence path, the server may evaluate the `basicasteroids` drop table. Successful drop results create authoritative pickups at the source position and emit `pickup_dropped`.

Current `basicasteroids` behavior is:

```text
source type: asteroid
drop mode: single
max drops per source: 1
max active pickups: 2
```

Current entries are:

| Pickup type | Chance | Source size range |
| ----------- | -----: | ----------------- |
| `1_up`      | `0.01` | `1` through `4`   |
| `torpedo`   | `0.15` | `2` through `4`   |

The drop seam ends once the authoritative pickup has been created and the drop event has been recorded. Drop tables do not collect pickups, apply effects, expire pickups, or render pickups.

## Event model

Pickup events describe different semantic facts and should not be collapsed.

Current pickup-related events are:

```text
pickup_dropped
pickup_collected
pickup_effect_applied
pickup_expired
```

`pickup_dropped` means a pickup entity was successfully created from a drop-table result.

`pickup_collected` means a pickup entity was consumed by player contact and removed from authoritative pickup state.

`pickup_effect_applied` means the gameplay mutation from the pickup effect intent succeeded.

`pickup_expired` means the pickup reached its server-owned lifespan and was removed without collection.

Presentation systems may use these events for different feedback. World pickup collection effects should use `pickup_collected`. HUD, result, or telemetry feedback tied to successful gameplay mutation should use `pickup_effect_applied`.

## Targeting and presentation model

Pickups are valid gameplay target candidates for readout and telemetry flows. They are not valid substitutes for player-only commands or systems that require a player target.

Pickup presentation is packet-driven:

```text
StatePacket.pickups
-> client world sync
-> pickup presentation node
```

Current packet-facing pickup fields include:

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

`pickup_class` selects the pickup scene family. `type` selects the pickup identity and badge/icon.

Scene paths are not part of server pickup definitions and are not sent by the server. The server sends class/type identity; the client maps that to local presentation assets.

Pickup end-of-life blink is client presentation derived from authoritative age/lifespan values. Actual expiry remains server-owned.

## Invariants

Pickup behavior must preserve these rules:

* Pickups are server-authoritative runtime entities.
* Client presentation must not create, collect, expire, or apply authoritative pickups.
* Pickup source systems must converge on the same server spawn path.
* Pickup class and pickup type remain separate concepts.
* Pickup collision shape lookup uses pickup class, not pickup type.
* Pickup scene-family selection uses pickup class, not server-sent scene paths.
* Pickup badge/icon selection uses pickup type.
* Collection and effect application remain separate stages.
* `pickup_collected` and `pickup_effect_applied` remain separate events.
* Drop-table evaluation does not collect pickups or apply pickup effects.
* Pickup expiry does not apply collection effects.
* Pickup effects mutate only server-owned runtime/session state unless a separate durable grant flow is implemented.
* Runtime pickups do not create durable inventory, profile, hangar, or ownership records by default.
* Weapon pickup ammo is additive.
* Pickup lifespan authority stays on the server.
* Client lifespan presentation is derived, not authoritative.

## Participating systems

The main participating systems are:

```text
game-server pickup lifecycle
```

Owns authoritative pickup existence, spawn/remove helpers, age, expiry, and state projection.

```text
game-server pickup collection
```

Owns player/pickup collision handling, authoritative consumption, and collection events.

```text
game-server pickup effects
```

Owns effect-intent classification and game-owned effect application.

```text
game-server drop integration
```

Owns turning successful drop-table results into authoritative pickups.

```text
realtime gameplay protocol
```

Carries pickup state and pickup events to clients.

```text
data pipeline
```

Owns source-of-truth and generated data for packet schemas, constants, collision shapes, and drop tables.

```text
client world sync
```

Owns pickup rendering, interpolation, lifespan presentation, scene-family selection, badge/icon selection, and collection-effect presentation.

```text
devtools
```

May request pickup spawning through debug tooling, but the server still creates the authoritative pickup.

## Active issues

* Pickup health currently exists as current health only; pickups do not have a `max_health` field.
* Bullet/pickup collision damage is not enabled.
* Torpedo radial effects currently exclude pickups as targets.

See [Current System Limits](../../limits/current-system-limits.md#combat-systems).

## Related docs

* [Combat](./!README.md)
* [Game Server Simulation Pickups](../../services/game-server/simulation/pickups/!README.md)
* [Pickup Entity Lifecycle](../../services/game-server/simulation/pickups/pickup-entity-lifecycle.md)
* [Pickup Collection](../../services/game-server/simulation/pickups/pickup-collection.md)
* [Pickup Effects](../../services/game-server/simulation/pickups/pickup-effects.md)
* [Pickup Drop Integration](../../services/game-server/simulation/pickups/pickup-drop-integration.md)
* [Client Pickup Presentation](../../services/client/world-sync/pickup-presentation.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Drop Tables](../../data/drop-tables.md)
* [Constants](../../data/constants.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Player Build And Loadouts](../../planning/domains/gameplay/player-build-and-loadouts.md)
* [Inventory And Hangar](../../planning/domains/gameplay/inventory-and-hangar.md)
* [Progression And Rewards](../../planning/domains/gameplay/progression-and-rewards.md)

## Notes

Pickup mechanics are part of combat and reward pacing, but normal runtime pickups are not durable progression grants by default.

Future loadout, inventory, hangar, rare-drop, or durable reward systems should not reuse normal pickup collection as implicit persistence. They should add an explicit durable grant boundary if a pickup-like event is meant to create ownership.

The current pickup effect resolver is code-defined. Pickup metadata and packet shapes already use shared source data, but effect policy itself is not fully data-driven.
