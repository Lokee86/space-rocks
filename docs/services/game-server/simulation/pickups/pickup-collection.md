# Pickup Collection

Parent index: [Game Server Simulation Pickups](./!INDEX.md)

## Purpose

This document describes game-server pickup collection.

It explains how player/pickup collision is detected, how the authoritative pickup entity is consumed, how collection rules produce an effect intent, how the game applies that intent, and how collection/effect events are projected to clients.

## Overview

Pickup collection is an authoritative game-server simulation responsibility.

The current runtime flow is:

```text
active simulation tick
-> collision phase gate
-> player/pickup collision detection
-> authoritative pickup removal
-> pickup collection rule resolution
-> pickup_collected event recording
-> pickup effect intent application
-> pickup_effect_applied event recording when an effect succeeds
-> later state-packet projection
```

Pickup collection is intentionally two-stage:

```text
collection
= pickup entity was consumed and removed from the authoritative pickup map

effect application
= the gameplay mutation from the pickup type succeeded
```

Those stages produce separate events:

```text
pickup_collected
pickup_effect_applied
```

This lets client presentation react to collection separately from gameplay-result feedback. For example, the world can remove or animate the collected pickup from `pickup_collected`, while HUD or telemetry feedback can use `pickup_effect_applied`.

Current implemented pickup collection effects are:

```text
1_up    -> add_lives, amount 1
torpedo -> equip_weapon, secondary torpedo, ammo +1
```

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/pickups/
```

Primary supporting packages:

```text
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/physics/
services/game-server/internal/game/space/
services/game-server/internal/game/events/
services/game-server/internal/game/weapons/
```

## Responsibilities

Pickup collection owns the game-server side of:

* Detecting player/pickup collision during the authoritative collision phase.
* Respecting the world collision freeze gate.
* Skipping player entities that are pending despawn.
* Resolving pickup collision bodies through the server collision shape catalog.
* Using wrapped-space local placement for pickup collision checks.
* Removing the collected pickup from authoritative runtime state.
* Converting the pickup collision into a `pickups.CollectionRequest`.
* Resolving pickup collection rules through `pickups.ResolveCollection`.
* Recording `pickup_collected` when collection succeeds.
* Applying returned pickup effect intents through game-owned mutation helpers.
* Recording `pickup_effect_applied` when effect application succeeds.
* Keeping collection rules separate from effect mutation and packet projection.

## Does not own

Pickup collection does not own:

* Pickup spawning.
* Pickup drop-table evaluation.
* Pickup expiry and lifespan stepping.
* Pickup entity definitions.
* Collision primitive math.
* Collision shape export/import.
* Weapon firing policy.
* Weapon profile definition.
* Player input handling.
* Player death and respawn.
* Scoring policy.
* Room membership or match lifecycle.
* WebSocket transport.
* Packet codec behavior.
* Client pickup rendering, interpolation, audio, or effects.

Those systems may participate before or after pickup collection, but they own their own boundaries.

## Domain roles

Pickup collection participates in the player-facing pickup and reward flow by enforcing server authority over:

* whether a player touched a pickup
* whether a pickup was consumed
* which player collected it
* which pickup type was collected
* which gameplay effect should be attempted
* whether that effect changed authoritative game state
* which collection/effect presentation events clients receive

It also participates in the technical simulation flow by preserving collision-phase ordering and by keeping runtime mutation inside the `Game` aggregate instead of inside pure pickup rules.

## Simulation phase position

`Game.Step` runs pickup collection through the normal collision phase for active matches.

The active-match simulation path is:

```text
step player sessions
-> step player weapons
-> step players
-> remove ready players
-> step asteroid spawning
-> step asteroids
-> step bullets
-> step pickups
-> step collisions
-> step radial effects
-> simulation step observers
```

Pickup collection runs inside `stepCollisions` after ship/asteroid collision and bullet/asteroid collision:

```text
handleShipAsteroidCollisions
-> handleBulletAsteroidCollisions
-> handlePlayerPickupCollisions
```

The collision phase is gated by:

```go
game.worldSimulationOptions.CanRunCollisions()
```

When collisions are frozen, player/pickup collision does not consume pickups or apply pickup effects.

If the match is already over, `Game.Step` does not run the collision phase. It still steps pickups for age and expiry before returning.

## Collection detection model

Player/pickup collision detection is implemented by:

```go
detectPlayerPickupCollision(...)
```

The helper asks both runtime entities for collision bodies:

```text
player.CollisionBody(catalog)
pickup.CollisionBody(catalog)
```

Pickup collision shapes are selected by pickup class, not pickup type.

Current class keys include:

```text
powerup
weapon
```

This means `1_up` uses the `powerup` pickup collision shape and `torpedo` uses the `weapon` pickup collision shape.

For wrapped-world checks, the pickup body is temporarily placed in wrapped-local space near the player:

```text
delta = space.Delta(player.Position(), pickup.Position())
pickupBody.Position = player.Position().Add(delta)
```

The authoritative stored pickup position is not rewritten for collision detection. The adjusted position is only used for overlap testing across toroidal world boundaries.

If either collision body cannot be built, the helper returns no collision.

## Player eligibility

The current collection loop considers active player entities in `game.entities.Players`.

It skips a player only when:

```text
player.IsPendingDespawn()
```

Pickup collection does not use the same damage eligibility gates as player/asteroid collision. It does not check collision damage options, temporary invulnerability, or damage modifiers, because collection is not damage resolution.

The collision phase gate still applies. If collisions are frozen globally, pickup collection does not run.

## Collection flow

`handlePlayerPickupCollisions` owns the game adapter around the pure pickup rule.

The flow is:

```text
for each active player
-> skip pending-despawn player
-> for each pickup
-> skip nil pickup
-> detect player/pickup collision
-> remove pickup from authoritative pickup map
-> build pickup collection request
-> resolve collection through internal/game/pickups
-> if collected, record pickup_collected
-> apply returned effect intent
-> stop scanning more pickups for that player this pass
```

The pickup is removed before the effect intent is applied. Collection consumes the pickup entity even if the later effect application produces no gameplay mutation.

A player can collect at most one pickup per collision pass because the inner pickup loop breaks after the first detected collection.

## Collection rule model

Pure collection rules live in:

```text
services/game-server/internal/game/pickups/collection.go
```

The rule input is:

```go
type CollectionRequest struct {
    PlayerID   string
    PickupID   string
    PickupType string
    X          float64
    Y          float64
}
```

The rule output is:

```go
type CollectionResult struct {
    Collected    bool
    PlayerID     string
    PickupID     string
    PickupType   string
    X            float64
    Y            float64
    EffectIntent EffectIntent
}
```

`ResolveCollection` returns `Collected: false` only when the request is missing a player ID, pickup ID, or pickup type.

For a valid request, it returns `Collected: true` and attaches an effect intent based on pickup type.

Unknown pickup types resolve to a collected result with an empty no-op effect intent. In normal runtime, authoritative pickup spawning rejects unknown pickup types and pickup collision body lookup depends on known definitions, so unknown pickup collection is mostly a rule-level safety behavior.

## Effect intent model

Pickup collection rules do not directly mutate game state. They return an `EffectIntent`.

The current effect intent shape is:

```go
type EffectIntent struct {
    PlayerID   string
    PickupID   string
    PickupType string
    EffectType string
    Amount     int
    WeaponID   weapons.ID
    Slot       weapons.Slot
    Ammo       int
}
```

Implemented effect mappings are:

```text
1_up
-> EffectType: add_lives
-> Amount: 1

torpedo
-> EffectType: equip_weapon
-> WeaponID: torpedo
-> Slot: secondary
-> Ammo: 1
```

The rule package decides what should happen. The root game package applies the mutation because it owns player sessions, active player ships, weapon state, and event recording.

## Effect application

Effect application is implemented by:

```go
game.applyPickupEffectIntentLocked(...)
```

The current application rules are:

```text
empty effect type
-> no mutation
-> no pickup_effect_applied event

add_lives
-> mutate player session lives through addPlayerLivesLocked
-> record pickup_effect_applied when the player session exists

equip_weapon
-> require active player entity
-> require non-empty weapon id
-> equip limited-ammo weapon into the requested slot
-> add ammo to that slot
-> record pickup_effect_applied when application succeeds
```

`add_lives` mutates durable player session lives, not active ship health or avatar state.

`equip_weapon` currently mutates active runtime player weapon state. For torpedo pickups, the secondary weapon is set to torpedo and secondary ammo increases by one.

Weapon pickup ammo is additive. A torpedo pickup collected while torpedo is already equipped increases the secondary ammo count instead of replacing the ammo count.

## Event semantics

Pickup collection records events through the game-owned domain event adapter.

The current pickup collection event sequence for a successful `1_up` collection is:

```text
pickup_collected
pickup_effect_applied
```

`pickup_collected` means:

```text
the pickup entity was consumed and removed from authoritative pickup state
```

Its projected event fields are:

```text
type
player_id
pickup_id
pickup_type
x
y
```

`pickup_effect_applied` means:

```text
the gameplay mutation from the pickup effect intent succeeded
```

Its projected fields for `add_lives` include:

```text
type
player_id
pickup_id
pickup_type
effect_type
amount
lives_after
```

For `equip_weapon`, the current event includes the pickup identity, player identity, and effect type. The weapon/ammo result is primarily visible through later player and session state projection.

Events are queued for player sessions and delivered through `StatePacket.events`. The authoritative pickup removal is also visible through `StatePacket.pickups`, where the collected pickup is absent after collection.

## Data ownership

Pickup collection reads authoritative in-memory game runtime state.

It reads:

```text
game.entities.Players
game.entities.Pickups
game.collisionShapes
pickup type and position
player position and collision body
pickup collision body
```

It mutates:

```text
game.entities.Pickups
game.playerSessions lives for add_lives
active player ShipWeapons for equip_weapon
active player WeaponState ammo for equip_weapon
pending presentation/domain events
```

It does not persist account or local-profile data.

It does not write drop-table data, shared constants, packet schemas, or collision shape source files.

## Protocols and APIs

Pickup collection has no direct HTTP endpoint or direct client-callable WebSocket packet.

Clients observe pickup collection through normal game-server output:

```text
StatePacket.pickups
StatePacket.players
StatePacket.player_sessions
StatePacket.events
```

`StatePacket.pickups` no longer includes a collected pickup after the server removes it.

`StatePacket.events` can include:

```text
pickup_collected
pickup_effect_applied
```

The packet/event shape is generated from:

```text
shared/packets/gameplay.toml
```

Clients may trigger movement that eventually causes a player/pickup collision, but clients do not request pickup collection directly and do not decide collection outcomes.

## Invariants

Pickup collection must preserve these rules:

* The server is authoritative for pickup collection.
* Client rendering does not create, consume, or apply authoritative pickups.
* Player/pickup collision is collection, not damage.
* Pickup collision uses the pickup class collision shape, not a per-type collision shape.
* Collision checks use wrapped-local placement so cross-boundary collection can work.
* Frozen collisions prevent pickup collection.
* Pending-despawn players do not collect pickups.
* A collected pickup is removed from authoritative pickup state before effect application.
* Collection and effect application remain separate stages.
* `pickup_collected` and `pickup_effect_applied` remain separate events.
* Pure pickup rules return effect intents; they do not mutate game state.
* Game-owned code applies pickup effects because it owns player sessions and active runtime player state.
* `1_up` increments durable player session lives.
* Torpedo pickup ammo is additive.
* Pickup drop evaluation does not collect pickups or apply effects.
* Pickup expiry does not apply collection effects.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/pickup_effects.go
services/game-server/internal/game/pickups/collection.go
services/game-server/internal/game/collisions.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/events.go
services/game-server/internal/game/events/events.go
services/game-server/internal/game/player_counters.go
```

Pickup entity and definition files:

```text
services/game-server/internal/game/entities/pickups/types.go
services/game-server/internal/game/entities/pickups/definitions.go
services/game-server/internal/game/entities/pickups/pickup.go
```

Related runtime and generated files:

```text
services/game-server/internal/game/pickups.go
services/game-server/internal/game/pickup_lifecycle.go
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/constants/constants.go
```

Source-of-truth files:

```text
shared/packets/gameplay.toml
shared/constants/pickups.toml
shared/collisions/collision_shapes.json
```

Related tests:

```text
services/game-server/internal/game/pickups/collection_test.go
services/game-server/internal/game/pickup_effects_test.go
services/game-server/internal/game/entities/pickups/pickup_test.go
services/game-server/tests/game/pickups_test.go
services/game-server/tests/game/collision_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/drops/
services/game-server/internal/game/physics/
services/game-server/internal/game/weapons/
services/game-server/internal/networking/
client/
tools/data_sync/
```

`drops` evaluates pickup drops but does not collect pickups or apply pickup effects.

`physics` owns primitive collision math and collision shape loading, not pickup collection rules.

`weapons` owns weapon IDs, slots, profiles, and firing policy, not pickup collision.

`networking` owns transport and packet routing, not pickup outcomes.

`client` owns pickup rendering and presentation feedback only.

`tools/data_sync` owns generation from shared packet/constants data.

## Tests

Focused pickup rule tests cover:

* `1_up` collection intent.
* `torpedo` collection intent.
* unknown pickup types resolving to a no-op effect intent.

Game integration tests cover:

* spawned pickups being stored.
* pickup IDs and definition-backed type values.
* pickup health, age, and lifespan initialization.
* pickup expiry and expiry events.
* pickup removal.
* `StatePacket.pickups` projection.
* pickup age projection.
* player/pickup collision removing the pickup.
* frozen collisions preventing pickup collection.
* `1_up` incrementing player session lives.
* updated lives appearing in state packets.
* collection and effect events being emitted in order.

Effect application tests cover:

* torpedo pickups adding ammo to an already equipped secondary weapon.
* torpedo pickups equipping an empty secondary slot and adding ammo.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for pickup collection behavior:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/pickups ./tests/game -run 'Pickup|OneUp|Torpedo'
```

## Related docs

* [Game Server Simulation Pickups](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Pickup Effects](pickup-effects.md)
* [Pickup Entity Lifecycle](pickup-entity-lifecycle.md)
* [Pickup Drop Integration](pickup-drop-integration.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Collision Shapes](../world/collision-shapes.md)
* [Physics](../world/physics.md)
* [Toroidal Space And Motion](../world/toroidal-space-and-motion.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Player Pause And Suspension](../players/player-pause-and-suspension.md)
* [Data](../../../../data/!INDEX.md)
* [Protocol](../../../../protocol/!INDEX.md)
* [Devtools](../../../../devtools/!INDEX.md)

## Notes

The collection event position currently comes from the pickup position stored in the collection request. For tests where the pickup is spawned at the player position, that matches the player position.

The pickup collision fact stores an impact position, but the current collection adapter does not use that impact position when recording pickup collection events.

Pickup collection currently allows only one collected pickup per player per collision pass because the player loop breaks after the first detected pickup collision.
