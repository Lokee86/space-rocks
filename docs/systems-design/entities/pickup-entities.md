## Pickup Entities

Parent index: [Entities](./!README.md)

## Purpose

This document describes the systems-design model for pickup entities.

It defines what a pickup entity is, how pickup identity is split between class and type, which systems own pickup existence, what invariants must be preserved, and how pickup entities relate to collection, effects, presentation, drop tables, targeting, and runtime state.

## Overview

Pickup entities are temporary, server-authoritative world entities that exist during a live match. They can be spawned by gameplay systems or devtools, projected to clients through gameplay state packets, collected by players through authoritative collision, or removed by server-owned expiry.

A pickup entity is not the same thing as a pickup effect. The entity is the world object: it has an id, position, type, health, age, and lifespan. The effect is the gameplay mutation attempted after a player collects the entity.

Current implemented pickup entity types are:

```text
1_up
torpedo
```

Current implemented pickup classes are:

```text
powerup
weapon
```

The current entity loop is:

```text
spawn source
-> authoritative pickup entity created
-> pickup stored in server entity map
-> pickup projected in StatePacket.pickups
-> client presents pickup from packet state
-> server advances pickup age
-> pickup is removed by collection or expiry
-> event lane reports the semantic removal reason
```

Pickup entities are match-runtime objects. They do not create durable inventory, account ownership, profile records, hangar state, or progression grants by default. If a future pickup-like system needs durable grants, that durable grant flow should be explicit and separate from normal runtime pickup collection.

## Conceptual model

A pickup entity combines these conceptual parts:

```text
entity identity
class identity
type identity
lifecycle state
world position
collection eligibility
presentation state
```

Entity identity is the unique runtime id assigned by the server. Current ids use the server-owned `pickup_<number>` pattern.

Class identity is the broad pickup family. It is used for class-level behavior such as collision-shape lookup and client scene-family selection. Current classes are `powerup` and `weapon`.

Type identity is the specific gameplay identity. It is used for effect intent resolution and badge/icon selection. Current types are `1_up` and `torpedo`.

Lifecycle state is the authoritative server-owned age and lifespan data. The server increments pickup age and removes pickups whose positive lifespan has been reached. Current generated pickup definitions give implemented pickups a 12.0 second lifespan.

World position is server-owned. The client may translate server position into visual position for toroidal rendering, but it does not own the authoritative pickup location.

Collection eligibility is owned by the server collision phase. A pickup is collected only when authoritative player/pickup collision is detected during active collision processing.

Presentation state is client-only. The client chooses local scenes, icons, glow, pulse, blink, audio, and collection particles from server-supplied pickup facts.

## Entity lifecycle

Pickup lifecycle has three phases:

```text
creation
active existence
removal
```

Creation happens when game-owned code requests a pickup spawn using a known pickup type and position. The server validates the pickup type against pickup definitions, assigns an id, initializes health, age, and lifespan, and inserts the pickup into the active entity map.

Active existence means the pickup remains in authoritative runtime state. While active, it is projected to clients through `StatePacket.pickups`.

Current packet-facing pickup fields are:

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

Removal currently happens through collection or expiry.

Collection removal means a player touched the pickup during authoritative collision processing. The pickup is removed from active state, `pickup_collected` is recorded, and then the pickup effect intent is resolved and applied if valid.

Expiry removal means server-owned age reached lifespan. The server records `pickup_expired` and removes the pickup without applying collection effects.

Removal reason matters. A missing pickup in the next state packet means the entity is no longer active, but events explain whether the meaningful reason was collection, effect application, drop creation, or expiry.

## Authority rules

The game server owns authoritative pickup entity state.

The server owns:

* pickup creation
* pickup id assignment
* pickup type validation
* pickup class derivation
* pickup position
* pickup health value
* pickup age and lifespan
* pickup expiry
* pickup collision bodies
* pickup collection
* pickup removal
* pickup event recording
* pickup state packet projection

The client owns presentation only.

The client may:

* render a pickup node from `StatePacket.pickups`
* select a local pickup scene family from `pickup_class`
* select a badge/icon from pickup `type`
* derive end-of-life blink from `age_seconds` and `lifespan_seconds`
* play spawn and collection presentation effects
* remove a local pickup node when the server no longer reports it

The client must not:

* create authoritative pickups
* collect pickups locally
* expire pickups locally
* apply pickup effects locally
* invent pickup type/class mappings that contradict server state
* receive or trust server-sent scene paths

Drop tables are authoritative data inputs, but they do not own pickup lifecycle. A drop table may produce a pickup result. The game server then creates a normal authoritative pickup entity. After creation, the pickup uses the same lifecycle, collection, and expiry model as any other pickup source.

Devtools may request a pickup spawn, but devtools do not bypass pickup entity authority. The server still validates the type and creates the authoritative pickup.

## Collection and effect boundary

Pickup collection is an entity-consumption fact. Pickup effect application is a gameplay-mutation fact.

Those stages are intentionally separate:

```text
pickup entity collected
-> entity removed
-> pickup_collected recorded
-> effect intent resolved
-> game-owned mutation attempted
-> pickup_effect_applied recorded only if mutation succeeds
```

This separation lets the world present the pickup disappearing even when an effect is empty, unknown, invalid, or fails to mutate gameplay state.

Current effect mappings are:

```text
1_up
-> add_lives
-> amount 1

torpedo
-> equip_weapon
-> weapon torpedo
-> secondary slot
-> ammo +1
```

The pickup entity package does not apply these effects. Entity code owns the pickup object, definition lookup, class lookup, position read model, and collision body construction. The pickup rule seam resolves collection and effect intent. The root game aggregate applies effects because it owns player sessions, live ships, weapon state, and event queues.

## Class and type rules

Pickup class and pickup type must remain separate concepts.

Pickup class answers:

```text
What broad pickup family is this?
```

Pickup type answers:

```text
What specific pickup is this?
```

Current class responsibilities:

* server collision shape lookup
* client scene-family selection
* broad presentation grouping

Current type responsibilities:

* gameplay effect identity
* badge/icon identity
* event payload identity
* drop-table result identity

Collision shape lookup uses pickup class, not pickup type. This means `1_up` uses the `powerup` pickup collision shape and `torpedo` uses the `weapon` pickup collision shape.

Client scene selection also uses pickup class. Scene paths are client-side presentation details and are not part of server pickup definitions.

Client badge/icon selection uses pickup type. Pickup scene `Badge` children are expected to match packet pickup type strings.

## Targeting role

Pickup is a canonical gameplay target kind for readout, telemetry, and target candidate flows.

A pickup target is not a player target. It must not be accepted by systems or devtools commands that require `target_player_id`.

Targeting can observe pickup identity and position, but it does not own pickup collection, effects, expiry, or server validation. Server-owned systems remain responsible for deciding whether a target request is meaningful and whether any gameplay action can affect a pickup.

## Invariants

Pickup entity behavior must preserve these rules:

* Pickups are server-authoritative runtime entities.
* Pickup creation must pass through server-owned validation.
* Pickup source systems must converge on the same authoritative entity model.
* Pickup class and pickup type must remain separate.
* Pickup collision shape lookup uses pickup class.
* Pickup scene-family selection uses pickup class.
* Pickup badge/icon selection uses pickup type.
* Scene paths stay client-side and are not sent in gameplay packets.
* Pickup age and lifespan are server-owned lifecycle facts.
* Client lifespan blink is derived presentation, not expiry authority.
* Collection consumes the pickup entity before effect application.
* Collection and effect application remain separate stages.
* `pickup_collected` and `pickup_effect_applied` remain separate events.
* `pickup_expired` does not apply collection effects.
* Drop-table evaluation does not collect pickups or apply pickup effects.
* Runtime pickups do not create durable inventory, profile, hangar, or ownership state by default.
* Weapon pickup ammo is additive.
* Client presentation must not create, collect, expire, or mutate authoritative pickups.

## Participating systems

```text
game-server pickup entity lifecycle
```

Owns authoritative pickup creation, storage, age, expiry, removal, and state projection.

```text
game-server pickup collection
```

Owns player/pickup collision handling, collection consumption, and `pickup_collected` event production.

```text
game-server pickup effects
```

Owns effect intent resolution and game-owned mutation after collection.

```text
game-server drop integration
```

Owns turning successful drop-table results into authoritative pickup entities.

```text
realtime gameplay protocol
```

Carries pickup state and pickup events from server to client.

```text
data pipeline
```

Owns source-of-truth and generated data for pickup constants, packet fields, collision shapes, and drop tables.

```text
client world sync
```

Owns pickup node creation, scene-family selection, badge/icon visibility, interpolation, lifespan presentation, and collection presentation effects.

```text
targeting
```

Treats pickups as target candidates for readout and telemetry while preserving the distinction between pickup targets and player targets.

```text
devtools
```

May request debug pickup spawning and list pickup presentation options, but server authority still decides whether a pickup exists.

## Service implementation touchpoints

The main game-server entity files are:

```text
services/game-server/internal/game/entities/pickups/types.go
services/game-server/internal/game/entities/pickups/definitions.go
services/game-server/internal/game/entities/pickups/pickup.go
services/game-server/internal/game/pickups.go
services/game-server/internal/game/pickup_lifecycle.go
services/game-server/internal/game/state_packet.go
```

The main adjacent server behavior files are:

```text
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/pickups/collection.go
services/game-server/internal/game/pickup_effects.go
services/game-server/internal/game/pickup_drops.go
services/game-server/internal/game/collisions.go
```

The main client presentation boundary is:

```text
client/scripts/world/pickup_sync.gd
client/scripts/world/pickup_sync_state.gd
client/scripts/world/pickups/pickup_presentation_catalog.gd
client/scripts/entities/pickup.gd
client/scenes/pickups/powerup_pickup.tscn
client/scenes/pickups/weapon_pickup.tscn
```

These paths are implementation touchpoints, not a complete ownership map. Detailed implementation ownership belongs in the service docs.

## Active issues

* Pickup health currently exists as current health only; pickups do not have a `max_health` field.
* Bullet/pickup collision damage is not enabled.
* Torpedo radial effects currently exclude pickups as targets.

See [Current System Limits](../../limits/current-system-limits.md#combat-systems).

## Related docs

* [Entities](./!README.md)
* [Combat Pickups](../combat/pickups.md)
* [Targeting](../combat/targeting.md)
* [Weapons](../combat/weapons.md)
* [Radial Effects](../combat/radial-effects.md)
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

Pickup entities are documented here at the conceptual entity level. Combat pickup effects, server lifecycle implementation, client presentation, packet schema ownership, and drop-table data are documented in their owning folders.

The current server still steps pickup age and expiry in the match-over simulation branch, but match-over simulation does not run normal collision handling. That means pickups may expire after match over, but they are not collected through the normal player/pickup collision path once the match is over.

The current effect resolver is code-defined. Pickup metadata and packet fields use shared source data, but effect policy itself is not fully data-driven.
