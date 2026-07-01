## Target Selection And Status

Parent index: [Game Server Simulation Targeting](./!INDEX.md)

## Purpose

This document describes game-server target selection and target status handling.

It covers how the authoritative game simulation accepts or rejects target requests, stores selected target identity, validates position-based target selection, classifies selected target availability, and clears targets when their referenced entity is removed.

## Overview

Target selection is the game-server simulation boundary that turns client target intent into authoritative target state.

The current target identity model is:

```text
target_kind
target_id
```

Supported target kinds are:

```text
player
enemy
pickup
asteroid
bullet
```

Client-side targeting may choose a candidate from local presentation state, but the game server owns whether that request is accepted. The server checks the requesting player session, validates the requested target reference, and mutates session-owned targeting state only when the request passes the relevant server-side checks.

Position-based target selection does not ask the server to choose from all overlapping targets. The client sends:

```text
x
y
target_kind
target_id
```

The server then verifies that the submitted target exists, has a target candidate collision body, and contains the submitted point. If validation fails, the existing selected target remains unchanged.

Target status is an internal read model used to classify a referenced target as:

```text
active
inactive
missing
```

This status is not currently projected as a standalone packet field. It is used by server targeting code to preserve inactive-but-existing targets, such as pending-respawn players, while clearing targets that reference removed entities.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/targeting/
services/game-server/internal/game/player/
```

## Responsibilities

Target selection and status owns:

* Defining canonical runtime target references through `TargetRef`.
* Defining current target kinds.
* Accepting target selection requests from game-server networking adapters.
* Rejecting target selection from missing requesting player sessions.
* Rejecting empty target refs for position-based selection.
* Allowing explicit target clear requests.
* Validating that requested target entities exist before storing them.
* Validating position-based selection against the target candidate collision body.
* Storing selected target identity in `playerSession.Targeting`.
* Applying selected target identity to the active `runtime.Ship` copy when one exists.
* Reapplying session-owned target identity when a ship is created or respawned.
* Preserving existing target state when a new invalid selection request fails.
* Classifying target refs as active, inactive, or missing.
* Clearing targets that reference removed/missing entities during player removal cleanup.

## Does not own

Target selection and status does not own:

* Client-side candidate picking.
* Client-side mouse input routing.
* Client-side target highlighting or presentation.
* WebSocket transport.
* Packet encoding or decoding.
* Packet schema source files.
* Combat targeting mechanics.
* Lock-on behavior.
* Auto-targeting behavior.
* Weapon eligibility.
* Damage resolution.
* Scoring.
* Pickup collection or pickup effects.
* Devtools command-specific target resolution.
* Player lifecycle ownership.
* Room membership or lobby state.
* Durable profile, account, inventory, or progression data.

## Domain roles

Target selection participates in gameplay by recording what target a player has selected.

The current server-side role is intentionally narrow:

```text
client target intent
-> server validation
-> session-owned selected target ref
-> active ship target copy
-> world lane ship record readback through ShipState target fields
```

The selected target is gameplay state, but it is not automatically combat behavior. A target being selected does not mean the target is damaged, locked, collectable, or otherwise affected. Other systems must explicitly consume target state and enforce their own rules.

Target status participates in player/session lifecycle by distinguishing these cases:

```text
active target
= exists and is currently targetable or not pending despawn

inactive target
= exists but is not currently targetable, such as a pending-respawn player or pending-despawn entity

missing target
= no longer exists in the owning runtime/session store
```

This distinction prevents cleanup from treating an inactive player as removed. A pending-respawn player can remain a valid selected target identity even though the player has no active ship candidate.

## Protocols and APIs

Target selection has no HTTP API.

The runtime surface is a game-server service API consumed by inbound realtime packet routing:

```go
func (game *Game) SelectTargetAtPosition(playerID string, x float64, y float64, target targeting.TargetRef) bool
func (game *Game) SetTarget(playerID string, target targeting.TargetRef) bool
func (game *Game) ClearTarget(playerID string)
func (game *Game) Target(playerID string) targeting.TargetRef
func (game *Game) SetPlayerTarget(playerID string, targetPlayerID string) bool
func (game *Game) ClearPlayerTarget(playerID string) bool
func (game *Game) PlayerTarget(playerID string) string
```

The realtime packet surface is consumed by the game-server inbound gameplay adapter. It exists so the client can send target selection intent to the authoritative simulation.

Current inbound target request packet types are:

```text
set_target_player_request
select_target_at_position_request
clear_target_request
```

`select_target_at_position_request` carries:

```text
x
y
target_kind
target_id
```

The inbound adapter converts those fields into a `TargetRef` and calls `Game.SelectTargetAtPosition`.

`clear_target_request` clears the requesting player's selected target by calling `Game.ClearTarget`.

`set_target_player_request` is a player-target compatibility path. The inbound adapter passes `packet.TargetID` to `Game.SetPlayerTarget`, which wraps it as a player target. The server-side adapter does not use `packet.TargetKind` for this packet type.

The outbound readback surface is `world lane ship records[*].target_kind` and `world lane ship records[*].target_id`, projected from each active ship's `runtime.ShipState`.

Target requests are not authority. They are client intent. The server owns whether the target is accepted and later projects accepted target state back through authoritative gameplay state.

## Target reference model

The target package defines:

```go
type TargetKind string

const (
    TargetKindPlayer   TargetKind = "player"
    TargetKindEnemy    TargetKind = "enemy"
    TargetKindPickup   TargetKind = "pickup"
    TargetKindAsteroid TargetKind = "asteroid"
    TargetKindBullet   TargetKind = "bullet"
)

type TargetRef struct {
    Kind TargetKind
    ID   string
}
```

An empty target is represented by an empty kind or empty ID.

```go
func EmptyTarget() TargetRef
func (target TargetRef) IsEmpty() bool
```

A target ref is valid only when the owning server code can resolve the referenced entity or player session according to the current target kind.

Current target existence checks are:

```text
player
-> player world state exists for target player ID

enemy
-> game.entities.Enemies[target_id] exists and is non-nil

pickup
-> game.entities.Pickups[target_id] exists and is non-nil

asteroid
-> game.entities.Asteroids[target_id] exists and is non-nil

bullet
-> game.entities.Projectiles[target_id] exists and is non-nil
```

For players, existence is session/world-state based rather than active-ship-map based. A player can therefore exist as a target identity while inactive or pending respawn.

## Selection flow

### Generic set

`Game.SetTarget` stores a target ref directly after validation.

The current flow is:

```text
lock game
-> require requester player session exists
-> if target is empty, clear targeting
-> otherwise require target exists
-> store target in session targeting
-> apply target to active ship if present
-> unlock game
```

If the requester is missing, the call fails.

If the requested target is missing, the call fails and existing target state is not overwritten.

If the requested target is empty and the requester exists, the call clears the selected target.

### Position-based selection

`Game.SelectTargetAtPosition` validates a client-selected target against authoritative server collision bodies.

The current flow is:

```text
lock game
-> require requester player session exists
-> reject empty target ref
-> require target exists
-> build current target candidates
-> find candidate matching target_kind + target_id
-> require submitted point to be inside candidate collision body
-> store target in session targeting
-> apply target to active ship if present
-> unlock game
```

Position-based selection validates the target ref the client sent. It does not choose the highest-priority target from the click point.

Current target candidates are built from:

```text
game.entities.Players
game.entities.Asteroids
game.entities.Projectiles
game.entities.Pickups
game.entities.Enemies
```

A candidate is included only when the entity is non-nil and can produce a collision body. Active player candidates additionally skip pending-despawn ships.

Candidate body checks use:

```go
physics.BodyContainsPoint(matchedCandidate.Body, clickPoint)
```

If the target is missing, not present in the candidate list, or does not contain the submitted point, the request fails and the previous target remains unchanged.

### Clear

`Game.ClearTarget` clears the selected target by setting the player's target to `EmptyTarget`.

Clearing updates:

```text
playerSession.Targeting
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

when the active ship exists.

A dead or pending-respawn requester can clear session-owned targeting even when no active ship exists.

## Session and active ship storage

Durable match-local target selection is stored in:

```go
playerSession.Targeting PlayerTargeting
```

`PlayerTargeting` stores string copies of the canonical target identity:

```go
type PlayerTargeting struct {
    Kind string
    ID   string
}
```

The active ship carries a packet-facing copy:

```go
runtime.Ship.TargetKind string
runtime.Ship.TargetID   string
```

When targeting changes, `setPlayerTargetingLocked` updates the session first. If an active ship exists for the same player ID, it also applies the target to the ship.

When a new active ship is created from a session, `session.NewShip` calls:

```go
session.Targeting.ApplyToShip(ship)
```

This means target selection survives death and respawn at the session level. The respawned ship receives the session-owned target copy during ship creation.

## Target status model

Target status is defined under the player package:

```go
type TargetStatus string

const (
    TargetStatusMissing  TargetStatus = "missing"
    TargetStatusInactive TargetStatus = "inactive"
    TargetStatusActive   TargetStatus = "active"
)
```

For player targets, status comes from `player.WorldState`:

```text
missing
= player world state does not exist

active
= player world state exists and Targetable is true

inactive
= player world state exists and Targetable is false
```

`player.WorldState.Targetable` is true only for active players with an active ship.

For asteroid, bullet, and enemy targets, status is:

```text
missing
= entity map entry is absent or nil

inactive
= entity exists but IsPendingDespawn() is true

active
= entity exists and is not pending despawn
```

For pickup targets, current status is:

```text
missing
= pickup map entry is absent or nil

active
= pickup exists
```

Pickup target status currently does not classify a separate pending-despawn state inside `targetLookupStatusLocked`.

## Missing target cleanup

`clearTargetsForMissingPlayersLocked` removes target refs that point to missing targets.

The cleanup flow is:

```text
for each active player ship in game.entities.Players
-> read ship TargetKind and TargetID
-> skip empty target
-> classify target lookup status
-> if status is missing, clear targeting
-> keep active or inactive targets
```

This cleanup currently runs from `Game.RemovePlayer` after the removed player session and active ship are deleted.

The important behavior is:

```text
missing target
-> clear selected target

inactive target
-> keep selected target
```

This preserves target identity for existing but temporarily inactive targets, such as pending-respawn players.

## Data ownership

Target selection reads:

```text
game.playerSessions
game.entities.Players
game.entities.Enemies
game.entities.Pickups
game.entities.Asteroids
game.entities.Projectiles
game.collisionShapes
player.WorldState
runtime collision bodies
```

Target selection mutates:

```text
playerSession.Targeting
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

Target selection does not mutate the target entity.

It does not persist account, profile, inventory, progression, or match-result data.

Packet shape source data for target request fields and target readback fields lives in:

```text
shared/packets/gameplay.toml
```

Generated server packet output includes:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

Generated client packet output includes:

```text
client/scripts/generated/networking/packets/packets.gd
```

## Invariants

Target selection and status must preserve these rules:

* Canonical gameplay target identity is `target_kind` plus `target_id`.
* New normal gameplay paths must not introduce `target_player_id`.
* `target_player_id` remains devtools/player-only compatibility data, not the gameplay target model.
* Client-side candidate selection is not authority.
* Server-side position selection must validate both target identity and click containment.
* Invalid selection requests must not overwrite existing selected target state.
* Empty target refs clear through explicit clear paths, not through failed position selection.
* Session targeting is the durable match-local target owner.
* Active ship target fields are packet-facing copies of session targeting.
* Respawned ships inherit session-owned target state.
* Target status must distinguish inactive existing targets from missing targets.
* Missing-target cleanup must clear missing targets without clearing inactive existing targets.
* Targeting state must not imply combat effects, damage, scoring, pickup collection, or devtools command eligibility.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/targeting.go
```

Owns game-level target selection, target clearing, target lookup, target candidate building, target status lookup, and missing-target cleanup.

```text
services/game-server/internal/game/player_targeting.go
```

Owns the session-to-ship target copy model.

```text
services/game-server/internal/game/targeting/targeting.go
```

Defines target kinds, `TargetRef`, target candidates, empty target behavior, target kind priority, and legacy requested-player validation helpers.

```text
services/game-server/internal/game/player/target_status.go
```

Defines target status values and player world-state-to-target-status conversion.

```text
services/game-server/internal/game/player/state.go
services/game-server/internal/game/player_world_state.go
```

Define player world state and the `Targetable` flag used by player target status.

```text
services/game-server/internal/game/session.go
```

Stores `playerSession.Targeting` and applies session-owned target state when creating active ships.

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/packets_generated.go
```

Define active ship target fields and project them into `ShipState`.

Networking adapter:

```text
services/game-server/internal/networking/inbound/gameplay.go
```

Adapts inbound target request packets into game targeting API calls.

Generated/source packet files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Related collision and entity support:

```text
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/entities/pickups/pickup.go
```

Important non-ownership boundaries:

```text
client/
services/game-server/internal/networking/
services/game-server/internal/rooms/
services/game-server/internal/game/damage/
services/game-server/internal/game/weapons/
services/game-server/internal/game/pickups/
services/game-server/internal/devtools/
```

`client` owns local input, candidate picking, target presentation, and request sending.

`networking` owns packet routing and transport.

`rooms` owns room membership and room-to-game-instance access.

`damage`, `weapons`, and `pickups` own their own gameplay consequences.

`devtools` owns command-specific debug target interpretation.

## Tests

Relevant game-server tests include:

```text
services/game-server/internal/game/targeting_test.go
services/game-server/internal/game/targeting/targeting_test.go
services/game-server/internal/game/player/target_status_test.go
```

Current coverage includes:

* player target storage
* target clearing
* missing target rejection without overwrite
* `world lane ship records[*].target_kind`
* `world lane ship records[*].target_id`
* generic player, asteroid, bullet, and pickup target refs
* position-based target selection for player, asteroid, bullet, and pickup targets
* missing target rejection for position-based selection
* non-overlapping click rejection without overwrite
* player target status for active, inactive, and missing players
* missing target cleanup
* inactive pending-respawn target preservation
* removed target cleanup
* dead requester target selection without panic
* dead requester target clear without panic
* respawn target copy from session to new ship
* target ref empty/non-empty behavior
* target kind priority ordering
* requested player target validation helper behavior

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/... ./internal/networking/inbound/...
```

Focused targeting verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game -run 'Target|Targeting'
go test -buildvcs=false ./internal/game/targeting
go test -buildvcs=false ./internal/game/player -run 'TargetStatus'
```

Run packet generation checks when target packet fields or target world lane packet fields change:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Game Server Simulation Targeting](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Lane Packet Projection](../runtime/lane-packet-projection.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Session State](../players/player-session-state.md)
* [Player Camera View State](../players/player-camera-view-state.md)
* [Pickup Entity Lifecycle](../pickups/pickup-entity-lifecycle.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Input And Targeting](../../../client/input-and-targeting.md)
* [Outbound Packet Sending](../../../client/networking-flow/outbound-packet-sending.md)
* [Gameplay Packets](../../../../protocol/gameplay-packets.md) - Gameplay packet documentation.
* [Packet Schema Pipeline](../../../../data/packet-schemas.md) - Packet schema and generated output documentation.
* [Canonical Target State](canonical-target-state.md) - Canonical target state documentation.

## Notes

The legacy targeting docs correctly identified the core quarantine rule: normal gameplay targeting uses `target_kind` plus `target_id`, while `target_player_id` is limited to devtools/player-only compatibility paths.

`TargetKindPriority` exists in the target policy package and is covered by tests, but current game-server position selection validates the client-submitted target ref rather than choosing the highest-priority candidate from overlapping bodies.

`set_target_player_request` remains a player-target compatibility request. New generic gameplay targeting should prefer `select_target_at_position_request` or game APIs that use `TargetRef`.

