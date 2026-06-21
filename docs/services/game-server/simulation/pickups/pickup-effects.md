# Pickup Effects

Parent index: [Game Server Simulation Pickups](./!INDEX.md)

## Purpose

This document describes the game-server service boundary for pickup effect resolution and application.

It covers how collected pickup types become effect intents, how the game server applies those intents to player runtime/session state, and how applied effects are projected as gameplay events.

## Overview

Pickup effects are server-authoritative.

The client does not decide which effect a pickup applies. The game server detects player/pickup contact, removes the pickup entity, resolves the pickup type into an effect intent, applies the effect while holding game-state authority, and emits packet-facing presentation events for clients to observe.

The current pickup effect path is intentionally two-stage:

```text
pickup collection
-> effect intent resolution
-> game-owned state mutation
-> pickup_effect_applied event
```

The `pickups` package owns effect intent classification. It decides what a pickup type means, but it does not mutate `Game`, player sessions, live ships, packets, or client state.

The root `game` package owns effect application because it owns the mutable player/session/runtime state that pickup effects change.

Current implemented pickup effects are:

```text
1_up
-> add_lives
-> +1 life on the player session

torpedo
-> equip_weapon
-> equip torpedo in the live secondary slot
-> add 1 secondary ammo
```

Unknown pickup types can still be collected if a pickup entity exists, but they resolve to an empty no-op effect intent. Empty, unknown, or failed effect intents do not emit `pickup_effect_applied`.

## Code root

```text
services/game-server/internal/game/pickups/
services/game-server/internal/game/pickup_effects.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/events.go
services/game-server/internal/game/events/
services/game-server/internal/game/weapons/
shared/constants/pickups.toml
shared/constants/weapon_pickups.toml
shared/packets/gameplay.toml
```

## Responsibilities

The game-server pickup effects boundary owns:

* Mapping collected pickup types to effect intents.
* Keeping collection classification separate from state mutation.
* Applying supported pickup effects to game-owned runtime/session state.
* Adding lives through the player-session counter path.
* Equipping pickup-granted weapons onto the active live player ship.
* Adding pickup-granted weapon ammo to the relevant live weapon slot.
* Emitting `pickup_effect_applied` only when effect application succeeds.
* Keeping `pickup_collected` and `pickup_effect_applied` as separate event semantics.
* Returning no-op intent data for unknown pickup types instead of inventing a mutation.
* Keeping pickup effects server-authoritative.

## Does not own

Pickup effects do not own:

* Pickup entity creation.
* Pickup entity lifetime or expiry.
* Pickup drop-table evaluation.
* Pickup spawn authority.
* Player/pickup collision detection.
* Collision shape data.
* Client pickup rendering.
* Client pickup collection particles or sounds.
* Weapon fire policy.
* Projectile creation.
* Damage resolution.
* Radial effect stepping.
* Durable inventory ownership.
* Future loadout validation.
* Player-data persistence.
* Packet schema source-of-truth files.
* WebSocket transport.

Those responsibilities belong to pickup lifecycle, pickup drop integration, world/physics, combat, weapons, client presentation, data, protocol, or player-data documentation.

## Domain roles

### Effect intent resolver

`services/game-server/internal/game/pickups/collection.go` resolves a collected pickup into a `CollectionResult`.

The resolver receives:

```text
player_id
pickup_id
pickup_type
x
y
```

It returns:

```text
collected flag
player_id
pickup_id
pickup_type
x
y
effect intent
```

The resolver requires non-empty player ID, pickup ID, and pickup type. If any required identifier is empty, the result is not collected.

If the pickup type is recognized, the result includes a populated `EffectIntent`.

If the pickup type is unknown, the result is still collected but the `EffectIntent` is empty. That means collection can consume a pickup without applying a gameplay mutation.

### Game-owned effect applicator

`services/game-server/internal/game/pickup_effects.go` applies effect intents to game-owned state.

The applicator is intentionally in the root `game` package because effects mutate state that is not owned by the pickup rules package:

```text
playerSessions
game.entities.Players
runtime.Ship.ShipWeapons
runtime.Ship.WeaponState
pendingPresentationEvents
```

Effect application returns a boolean success result. A successful effect records `pickup_effect_applied`. A failed or empty effect does not.

### Event producer

Pickup effects use the normal game-server domain-event adapter.

The game server records a domain event:

```text
events.EventPickupEffectApplied
```

Then `eventStateForDomainEvent` converts it into a packet-facing event state:

```text
type: pickup_effect_applied
player_id
pickup_id
pickup_type
effect_type
amount
lives_after
```

The event is queued into each player session’s pending presentation events and is drained into that player’s next state packet.

## Current effect intents

### `add_lives`

The `1_up` pickup resolves to:

```text
effect_type: add_lives
amount: 1
```

Application calls the player-session life counter path:

```text
game.addPlayerLivesLocked(player_id, amount)
```

The life mutation applies to `playerSessions`, not directly to the live ship/avatar entity.

When successful, the server emits:

```text
pickup_effect_applied
effect_type: add_lives
amount: 1
lives_after: <updated session lives>
```

If the player session cannot be found, the effect fails and no `pickup_effect_applied` event is emitted.

### `equip_weapon`

The `torpedo` pickup resolves to:

```text
effect_type: equip_weapon
weapon_id: torpedo
slot: secondary
ammo: 1
```

Application requires an active live player ship in:

```text
game.entities.Players[player_id]
```

When successful, the server equips the requested slot with a limited-ammo weapon:

```text
runtime.Ship.ShipWeapons.Secondary.ID = torpedo
runtime.Ship.ShipWeapons.Secondary.AmmoPolicy = limited
runtime.Ship.WeaponState.Secondary.AmmoRemaining += 1
```

Collecting another torpedo pickup adds ammo to the existing secondary slot. It does not reset ammo to one.

The current equip effect mutates the active runtime ship. It does not update durable inventory, future loadout state, or the player session armory.

When successful, the server emits:

```text
pickup_effect_applied
effect_type: equip_weapon
pickup_type: torpedo
```

The current event does not include weapon ID, slot, or ammo fields. The equipped weapon and ammo state are observed through normal player ship state projection.

## Effect application flow

Pickup effect application happens after player/pickup collision detection.

The current runtime flow is:

```text
Game.Step
-> stepCollisions
-> handlePlayerPickupCollisions
-> detectPlayerPickupCollision
-> removePickupLocked
-> pickups.ResolveCollection
-> record pickup_collected
-> applyPickupEffectIntentLocked
-> record pickup_effect_applied when mutation succeeds
```

The pickup entity is removed before the effect is applied.

This means the collection event and entity removal are not dependent on the effect mutation succeeding. If the effect intent is empty, unknown, invalid, or cannot be applied, the pickup remains consumed and no effect-applied event is emitted.

The current collision loop breaks after one pickup collision is handled for a player in that pass.

## Protocols and APIs

Pickup effects do not expose a public client request API.

The runtime surface is an internal game-server call path. Clients consume the results through gameplay state packets.

The packet-facing surface is:

```text
StatePacket.events[]
```

Relevant event types are:

```text
pickup_collected
pickup_effect_applied
```

`pickup_collected` means the pickup entity was consumed by player contact and removed from the authoritative pickup map.

`pickup_effect_applied` means a gameplay mutation succeeded after collection.

The server owns authority behind both events. The client may present sounds, particles, HUD updates, or readback state from these events, but it does not validate collection or apply the effect locally.

For `add_lives`, the effect-applied event carries `amount` and `lives_after`.

For `equip_weapon`, the current effect-applied event does not carry ammo or weapon-slot details. The client observes equipment and ammo through the normal player state fields:

```text
ShipState.secondary_weapon_id
ShipState.secondary_ammo_policy
ShipState.secondary_ammo_remaining
```

The packet schema and generated packet structs are owned by the realtime protocol and data pipeline, not by the pickup effects boundary.

## Data ownership

Pickup effects own no durable storage.

Current mutations are limited to game-server runtime state:

```text
playerSessions[player_id].Lives
runtime.Ship.ShipWeapons
runtime.Ship.WeaponState
pendingPresentationEvents
```

Pickup type metadata is sourced from shared constants:

```text
shared/constants/pickups.toml
shared/constants/weapon_pickups.toml
```

Generated Go constants currently live in:

```text
services/game-server/internal/constants/powerups.go
services/game-server/internal/constants/weapon_pickups.go
```

Pickup definitions consume generated constants for type, class, health, and lifespan.

The current effect resolver itself is code-defined in:

```text
services/game-server/internal/game/pickups/collection.go
```

That resolver maps `1_up` and `torpedo` to effect intents directly. Generated constants define current pickup metadata and weapon-pickup tuning values, but the current policy dispatcher is still the pickup collection code.

Packet-facing event structure is generated from the packet source of truth:

```text
shared/packets/gameplay.toml
services/game-server/internal/game/packets.go
```

Generated files should not be edited manually.

## Code map

Primary implementation files:

* `services/game-server/internal/game/pickups/collection.go` - Collection result and pickup effect intent classification.
* `services/game-server/internal/game/pickup_effects.go` - Game-owned application of pickup effect intents.
* `services/game-server/internal/game/pickup_collisions.go` - Player/pickup collision integration and call site for effect application.
* `services/game-server/internal/game/events/events.go` - Domain event types and event payload fields.
* `services/game-server/internal/game/events.go` - Domain event to packet event conversion and presentation-event queueing.
* `services/game-server/internal/game/player_counters.go` - Player score/life counter mutation helpers.
* `services/game-server/internal/game/weapons/types.go` - Weapon IDs, slots, ammo policies, equipped weapon state, and default armory shapes.
* `services/game-server/internal/game/runtime/ship.go` - Projection of live ship weapon and ammo state.
* `services/game-server/internal/game/state_packet.go` - State packet construction and event draining.
* `services/game-server/internal/game/packets.go` - Generated packet-facing event and state structs.
* `services/game-server/internal/game/entities/pickups/definitions.go` - Pickup definition lookup from generated constants.

Source-of-truth and generated files:

* `shared/constants/pickups.toml`
* `shared/constants/weapon_pickups.toml`
* `services/game-server/internal/constants/powerups.go`
* `services/game-server/internal/constants/weapon_pickups.go`
* `shared/packets/gameplay.toml`
* `services/game-server/internal/game/packets.go`

Important non-ownership boundaries:

* `services/game-server/internal/game/entities/pickups/` owns pickup entity definitions, not effect application.
* `services/game-server/internal/game/drops/` owns drop-table evaluation, not collection or effects.
* `services/game-server/internal/game/weapons/` owns weapon fire policy, not pickup collection.
* `services/game-server/internal/game/damage/` owns damage resolution, not pickup rewards.
* `services/game-server/internal/game/effects/radial/` owns radial effect timing and coverage, not pickup effects.
* `client/` owns pickup presentation and event effects, not authoritative pickup mutations.
* `shared/` owns packet and constant source-of-truth data.

## Tests

Relevant focused tests include:

* `services/game-server/internal/game/pickups/collection_test.go`
* `services/game-server/internal/game/pickup_effects_test.go`
* `services/game-server/internal/game/events_test.go`

Current coverage verifies:

* `1_up` resolves to `add_lives` amount `1`.
* Unknown pickup types resolve to a no-op effect intent.
* `torpedo` resolves to `equip_weapon`, weapon `torpedo`, slot `secondary`, ammo `1`.
* Applying a torpedo effect adds ammo to an already equipped secondary weapon.
* Applying a torpedo effect equips an empty secondary slot and adds one ammo.
* `pickup_effect_applied` converts to packet-facing event fields.
* State packets drain queued presentation events after delivery.

Broader verification should include the game-server Go test suite when pickup effects touch collection, player counters, weapon state, event projection, packet state, or collision integration.

## Related docs

* [Game Server Simulation Pickups](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Services](../../../!INDEX.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Pickup Presentation](../../../client/world-sync/pickup-presentation.md)
* [Gameplay Events And Effects](../../../client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Gameplay packets](../../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Constants pipeline](../../../../data/stubs/constants-pipeline.md) - Stub: constants pipeline documentation.
* [Packet schema pipeline](../../../../data/stubs/packet-schema-pipeline.md) - Stub: packet schema pipeline documentation.
* [Drop-table pipeline](../../../../data/stubs/drop-table-pipeline.md) - Stub: drop-table pipeline documentation.
* [Pickup entities](../../../../systems-design/entities/stubs/pickup-entities.md) - Stub: pickup entity design documentation.

## Notes

The current effect system is intentionally narrow. It supports lives and runtime weapon pickup behavior only.

`pickup_collected` and `pickup_effect_applied` should remain separate. Collection is the entity-consumption fact. Effect application is the gameplay-mutation fact.

The torpedo pickup currently changes the active runtime ship, not durable player inventory or future loadout state. Future loadout or inventory work should not be documented here as current behavior until implemented.

The current resolver uses code-defined pickup-type mappings. If pickup effect metadata later becomes fully data-driven, that belongs in the data pipeline documentation as well as this service boundary.
