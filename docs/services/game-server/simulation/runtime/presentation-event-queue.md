# Presentation Event Queue

Parent index: [Game Server Simulation Runtime](./!README.md)

## Purpose

This document describes the game-server presentation event queue.

It covers how simulation-owned gameplay facts are converted into packet-facing `EventState` values, queued per player, included in gameplay state packets, and drained after packet projection.

## Overview

The presentation event queue is a transient runtime lane on the game-server `Game` aggregate.

It carries short-lived client presentation facts such as bullet blasts, ship deaths, radial effect starts, pickup events, and damage presentation events. These facts are authoritative in the sense that the server decides when they happen, but the queue itself is not a durable event log and is not the domain event source of truth.

Current flow:

```text
simulation system records events.Event
-> game.recordDomainEvent
-> eventStateForDomainEvent
-> game.broadcastEvent
-> pendingPresentationEvents[playerID]
-> Game.StatePacket(playerID)
-> StatePacket.events
-> pendingPresentationEvents[playerID] = nil
-> networking encodes state packet
-> client gameplay event presentation
```

The queue is intentionally named:

```text
pendingPresentationEvents
```

It stores packet-facing `EventState` values for client presentation. It is not a domain event queue.

The server currently fans each event out to every player session that exists at event-recording time. A player with a durable session but no live ship can still receive the event. A player added after the event is recorded will not receive that event.

Events are drained per player when `Game.StatePacket(playerID)` is built. The first state packet after an event includes the queued events for that player. A later state packet for the same player no longer includes them unless new events were recorded.

This means the queue is best-effort presentation state. It does not guarantee delivery across encode failures, transport failures, disconnects, reconnects, or late joins.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/events/
```

Generated packet shapes come from:

```text
shared/packets/gameplay.toml
services/game-server/internal/game/packets.go
```

Outbound delivery uses:

```text
services/game-server/internal/networking/outbound/
```

## Responsibilities

The presentation event queue owns:

* Holding pending packet-facing presentation events per game player.
* Creating the queue map when a `Game` is constructed.
* Initializing a player's queue entry when `Game.AddPlayer` creates a player.
* Removing a player's queue entry when `Game.RemovePlayer` removes the player.
* Translating supported `events.Event` domain facts into generated `EventState` packet values.
* Broadcasting each converted event to every current player session.
* Preserving event order within each player's pending event slice.
* Copying pending events into `StatePacket.Events`.
* Clearing a player's pending events after `Game.StatePacket(playerID)` projects them.
* Keeping presentation events transient and non-durable.

## Does not own

The presentation event queue does not own:

* Collision detection.
* Damage resolution.
* Score, lives, death, respawn, or pickup effect authority.
* Radial effect timing or target selection.
* Pickup entity lifecycle.
* Domain event vocabulary design outside the currently supported presentation facts.
* Packet schema source-of-truth files.
* JSON encoding or websocket writes.
* Network retry, guaranteed delivery, acknowledgements, or replay.
* Client-side event routing, visual effects, audio, HUD updates, or match-end presentation.
* Durable player-data reporting or persistence.
* Room lifecycle or room match-over state transitions.
* Devtools-only event behavior.

Those concerns belong to combat, pickups, radial effects, state packet projection, protocol/data, networking, client presentation, rooms, or player-data documentation.

## Domain roles

The queue participates in the server-authoritative gameplay presentation flow.

Its role is narrow:

```text
server gameplay fact
-> transient packet event
-> client presentation input
```

The producing gameplay systems still own the authority behind each event.

Examples:

* Combat owns projectile and ship collision consequences.
* Damage resolution owns pure damage outcome calculation.
* Player lifecycle owns death, lives, and respawn state mutation.
* Pickup systems own pickup collection, expiry, drop, and effect application.
* Radial effects own radial zone timing and hit generation.
* The event queue only adapts selected facts into client-visible packet events.

This keeps client presentation event delivery separate from the runtime systems that decide what happened.

## Runtime lifecycle

### Queue creation

`Game.New()` initializes:

```text
pendingPresentationEvents map[string][]EventState
```

The map is owned by the `Game` aggregate and guarded by the same game mutex used for simulation stepping, player mutation, and state packet projection.

### Player addition

`Game.AddPlayer()` creates the durable player session, active ship, camera view, and event queue entry:

```text
game.pendingPresentationEvents[playerID] = nil
```

This prepares the player to receive presentation events emitted after they join the game simulation.

### Event recording

Current producers call:

```text
game.recordDomainEvent(events.Event{...})
```

`recordDomainEvent` converts the event into a packet-facing `EventState` and passes it to `broadcastEvent`.

The method is package-local. It is not a public protocol surface. It assumes it is called from game-owned code paths that are already inside the simulation lock, such as `Game.Step`, collision handling, pickup handling, or same-package tests.

### Fanout

`broadcastEvent` iterates over:

```text
game.playerSessions
```

and appends the packet event to each player's pending queue:

```text
game.pendingPresentationEvents[playerID] = append(game.pendingPresentationEvents[playerID], event)
```

Fanout uses durable player sessions rather than active ship state. This is intentional because a player can be pending respawn or eliminated while still needing presentation events such as match, death, or world-event feedback.

### State packet inclusion

`Game.StatePacket(playerID)` builds a packet for one player.

During projection, `statePacket(playerID)` copies that player's current pending events into:

```text
StatePacket.Events
```

The copy is made before the queue is cleared so the returned packet owns its own event slice.

### Drain

After `statePacket(playerID)` returns, `Game.StatePacket(playerID)` clears that player's pending event slice:

```text
game.pendingPresentationEvents[playerID] = nil
```

The queue is drained even though network encoding and websocket writing happen later. Presentation events are therefore not retried if later outbound work fails.

### Player removal

`Game.RemovePlayer(playerID)` deletes:

```text
game.pendingPresentationEvents[playerID]
```

This is full simulation removal. It is separate from normal player death, pending despawn, or pending respawn.

## Event adaptation

`events.Event` is the simulation-facing event fact. `EventState` is the packet-facing presentation shape.

The adapter is:

```text
eventStateForDomainEvent(events.Event) EventState
```

Current mappings are:

| Domain event                 | Packet event type          | Packet-facing fields                                                            |
| ---------------------------- | -------------------------- | ------------------------------------------------------------------------------- |
| `EventBulletBlast`           | `bullet_blast`             | `x`, `y`                                                                        |
| `EventRadialEffectStarted`   | `radial_effect_started`    | `source_id`, `effect_type`, `x`, `y`                                            |
| `EventShipDeath`             | `ship_death`               | `player_id`, `lives`, `respawn_delay`, `x`, `y`                                 |
| `EventPickupCollected`       | `pickup_collected`         | `player_id`, `pickup_id`, `pickup_type`, `x`, `y`                               |
| `EventPickupEffectApplied`   | `pickup_effect_applied`    | `player_id`, `pickup_id`, `pickup_type`, `effect_type`, `amount`, `lives_after` |
| `EventPickupExpired`         | `pickup_expired`           | `pickup_id`, `pickup_type`, `x`, `y`                                            |
| `EventPickupDropped`         | `pickup_dropped`           | `pickup_id`, `pickup_type`, `source_type`, `source_id`, `table_id`, `x`, `y`    |
| `EventDamageApplied`         | `damage_applied`           | `source_type`, `source_id`, `effect_type`, `amount`, `x`, `y`                   |
| `EventDamageOverTimeStarted` | `damage_over_time_started` | `source_type`, `source_id`, `effect_type`, `amount`                             |
| `EventDamageOverTimeTick`    | `damage_over_time_tick`    | `source_type`, `source_id`, `effect_type`, `amount`, `x`, `y`                   |

The adapter intentionally narrows event facts to the fields currently available in generated `EventState`.

For damage events, the packet-facing event currently preserves source identity, damage type through `effect_type`, amount, and coordinates where relevant. It does not project every damage result field such as target identity, cause, base amount, shield absorption, or remaining health.

## Protocols and APIs

The queue is an internal game-server runtime surface, not a standalone network protocol.

The network-visible surface is the `events` array on the generated gameplay state packet:

```text
StatePacket.events
```

That field is for client presentation. The client consumes it after receiving normal realtime gameplay state and routes supported events into local effects, audio, HUD, death, and match-end presentation.

Authority behind the events remains on the server:

```text
server simulation owns event production
Game owns event-to-packet queueing
StatePacket owns packet projection
networking owns JSON encode and websocket write
client owns presentation response
```

The queue does not own packet schema. Packet fields are generated from shared packet data.

The current internal surfaces are:

```text
recordDomainEvent(events.Event)
eventStateForDomainEvent(events.Event) EventState
broadcastEvent(EventState)
Game.StatePacket(playerID)
```

`BuildGameplayPresentationStateResponse` in networking calls `Game.StatePacket(playerID)`, stamps `server_sent_msec`, encodes the packet, and returns websocket bytes. It does not own event production or queue semantics.

## Data ownership

### Domain event facts

`services/game-server/internal/game/events.Event` carries simulation-facing event facts.

It can contain more fields than are currently projected into packets. The event value is transient and exists only while the game code records and adapts it.

### Packet event state

`EventState` is generated packet data.

Current generated fields are:

```text
type
player_id
lives
respawn_delay
x
y
pickup_id
pickup_type
source_type
source_id
table_id
lives_after
effect_type
amount
```

`EventState` is the value stored in `pendingPresentationEvents` and later included in `StatePacket.Events`.

### Pending queue state

The pending queue is:

```text
map[string][]EventState
```

The key is the game player ID.

The value is that player's not-yet-projected presentation event slice.

The queue is runtime-only. It is not persisted, not replayed, not mirrored to player-data, and not stored in room state.

## Code map

### Queue owner and lifecycle

```text
services/game-server/internal/game/game.go
```

Defines the `Game` aggregate field:

```text
pendingPresentationEvents map[string][]EventState
```

and initializes it in `New()`.

```text
services/game-server/internal/game/players.go
```

Creates a player's event queue entry in `AddPlayer` and deletes it in `RemovePlayer`.

### Event vocabulary and adapter

```text
services/game-server/internal/game/events/events.go
```

Defines the internal event vocabulary and the data shape used by simulation producers.

```text
services/game-server/internal/game/events.go
```

Owns `recordDomainEvent`, `eventStateForDomainEvent`, and `broadcastEvent`.

### State packet projection and drain

```text
services/game-server/internal/game/state_packet.go
```

Copies pending events into `StatePacket.Events` and drains the player's pending queue after projection.

```text
services/game-server/internal/game/packets.go
```

Generated packet definitions for `EventState` and `StatePacket`.

### Current event producers

```text
services/game-server/internal/game/combat.go
```

Records bullet blast, ship death, and damage-applied presentation events from combat consequences.

```text
services/game-server/internal/game/radial_spawning.go
```

Records radial-effect-start presentation events when projectile impact effects spawn radial effects.

```text
services/game-server/internal/game/simulation_radial_effects.go
```

Records damage-applied presentation events from radial hits.

```text
services/game-server/internal/game/pickup_drops.go
```

Records pickup-drop presentation events.

```text
services/game-server/internal/game/pickup_collisions.go
```

Records pickup-collected presentation events.

```text
services/game-server/internal/game/pickup_effects.go
```

Records pickup-effect-applied presentation events.

```text
services/game-server/internal/game/pickup_lifecycle.go
```

Records pickup-expired presentation events.

### Packet source and generated output

```text
shared/packets/gameplay.toml
```

Source-of-truth packet schema for `EventState` and `StatePacket.events`.

```text
services/game-server/internal/game/packets.go
```

Generated Go packet output consumed by the game and networking paths.

### Outbound networking consumer

```text
services/game-server/internal/networking/outbound/gameplay_presentation.go
```

Calls `Game.StatePacket(playerID)`, stamps `server_sent_msec`, encodes the packet through `packetcodec`, and returns the outbound websocket payload.

### Client presentation consumer

```text
client/scripts/gameplay/events/
client/scripts/gameplay/effects/
```

Consume packet events after state packet normalization and turn the supported event subset into local client presentation.

### Important non-ownership boundaries

```text
services/game-server/internal/game/damage/
```

Owns pure damage resolution. It does not record packet events directly.

```text
services/game-server/internal/game/effects/radial/
```

Owns radial effect timing and hit selection. It does not own packet queue storage.

```text
services/game-server/internal/protocol/packetcodec/
```

Owns JSON packet encode/decode helpers, not event queue semantics.

```text
shared/packets/
```

Owns packet schema source data.

```text
client/
```

Owns rendering, audio, effect nodes, HUD response, and match-end presentation.

## Tests

Relevant current tests include:

```text
services/game-server/internal/game/events_test.go
```

Covers:

* Event-to-packet conversion for bullet blasts, ship deaths, pickup events, damage events, and damage-over-time events.
* Queueing a recorded event for the current player.
* Queueing an event for a durable player session that has no live ship.
* Draining queued events through `Game.StatePacket(playerID)`.
* Creating damage presentation events only when damage results are not ignored and not no-op.

```text
services/game-server/internal/game/pickup_drops_test.go
```

Covers pickup drop behavior that later projects pickup state through gameplay state packets.

Suggested verification command from `services/game-server`:

```text
go test -buildvcs=false ./internal/game
```

## Related docs

* [Game Server Simulation Runtime](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Player Lifecycle Overview](../players/player-lifecycle-overview.md)
* [Player Death And Despawn](../players/player-death-and-despawn.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Pickup Effects](../pickups/pickup-effects.md)
* [Pickup Drop Integration](../pickups/pickup-drop-integration.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Damage Resolution](../combat/damage-resolution.md)
* [Radial Effects](../combat/radial-effects.md)
* [Gameplay Events And Effects](../../../client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Gameplay packets](../../../../protocol/stubs/gameplay-packets.md) - Gameplay realtime packet documentation.
* [Realtime websocket protocol](../../../../protocol/stubs/realtime-websocket-protocol.md) - Realtime websocket protocol documentation.
* [Packet Schema Pipeline](../../../../data/stubs/packet-schema-pipeline.md) - Packet schema generation documentation.

## Notes

Legacy documentation correctly identified the important naming rule: `pendingPresentationEvents` stores generated packet `EventState` values for client effects. It is not a domain event queue.

The event adapter currently returns a zero-value `EventState` for unsupported event types. Producers should not call `recordDomainEvent` for a new event type until `eventStateForDomainEvent`, packet schema, tests, and client presentation handling are updated as needed.

This document lives under simulation runtime because it documents a runtime queue on the `Game` aggregate. The concrete queue implementation is in the root `internal/game` package, not in the Go `internal/game/runtime` package.
