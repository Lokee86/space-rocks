## Canonical Target State

Parent index: [Game Server Simulation Targeting](./!README.md)

## Purpose

This document describes canonical target state in the game-server simulation.

It explains which game-server state owns a player's selected gameplay target, how that state is copied into active avatar state, how target identity is exposed through gameplay packets, and why player-only debug target fields must remain outside normal gameplay targeting.

## Overview

Canonical gameplay target state is the server-owned target identity stored for a player during a match.

The canonical identity shape is:

```text
target_kind
target_id
```

`target_kind` is the target type discriminator. `target_id` is the identifier inside that target type.

Current supported target kinds are:

```text
player
enemy
pickup
asteroid
bullet
```

The authoritative match-local target is stored on the player session:

```text
playerSession.Targeting
```

Active ships carry a copied packet-facing view:

```text
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

This split matters because a player may have a session without an active ship. A pending-respawn or temporarily dead player can still hold target state, clear target state, and later respawn with the same selected target copied onto the new ship.

The current server flow is:

```text
client target intent packet
-> networking inbound gameplay handler
-> Game.SelectTargetAtPosition / Game.SetTarget / Game.ClearTarget
-> playerSession.Targeting
-> optional copy to active runtime.Ship
-> runtime.Ship.State()
-> StatePacket.players[player_id].target_kind / target_id
-> client authoritative readback
```

Client clicks and local target candidates are requests. The server accepts or rejects them. Confirmed target state is the state projected back through authoritative gameplay state packets.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/targeting/
services/game-server/internal/game/player/
services/game-server/internal/networking/inbound/
shared/packets/
```

## Responsibilities

Canonical target state owns:

* Defining canonical gameplay target identity as `target_kind` plus `target_id`.
* Keeping the match-local target selection on `playerSession.Targeting`.
* Applying session-owned target state to a live `runtime.Ship` when one exists.
* Copying session-owned target state onto newly created or respawned ships.
* Exposing active ship target copies through `ShipState.target_kind` and `ShipState.target_id`.
* Validating target selection requests against existing server-side target entities.
* Validating point-based target selection against server collision bodies.
* Clearing a player's target by storing the empty target.
* Preserving target state for players that have a session but no active ship.
* Clearing targets that refer to removed/missing player sessions.
* Keeping inactive pending-respawn player targets distinct from missing removed-player targets.
* Keeping normal gameplay targeting on the canonical generic target pair rather than player-only compatibility fields.

## Does not own

Canonical target state does not own:

* Client-side click picking or local target candidate presentation.
* Client input routing or mouse action priority.
* WebSocket transport mechanics.
* Packet encoding mechanics.
* Packet schema source generation.
* Devtools command-specific target interpretation.
* `target_player_id` player-only debug command compatibility.
* Weapon fire policy.
* Damage resolution.
* Collision damage consequences.
* Score, lives, respawn, or match outcome rules.
* Spectate camera target selection.
* HUD or devtools telemetry rendering.
* Durable account or profile targeting persistence.

Those concerns belong to client input, networking, protocol/data, devtools, combat, players, rooms, or presentation documentation.

## Domain roles

Canonical target state participates in gameplay as a match-local player intent/read-model boundary.

It answers:

```text
What target has this player selected in authoritative game state?
```

It does not answer:

```text
Can this selected target be damaged by this weapon?
Should this target receive a command?
Should the client render this as selected?
Should the spectate camera follow this entity?
```

Those decisions belong to the consuming system.

A selected target can be useful for readout, presentation, devtools inspection, or future gameplay behavior without automatically making that entity damageable, commandable, or eligible for every target-consuming action.

## Runtime state model

The canonical target reference type is:

```go
type TargetRef struct {
    Kind TargetKind
    ID   string
}
```

The target kinds are:

```go
TargetKindPlayer
TargetKindEnemy
TargetKindPickup
TargetKindAsteroid
TargetKindBullet
```

The player-session target storage type is:

```go
type PlayerTargeting struct {
    Kind string
    ID   string
}
```

The empty target is represented by empty fields. A `TargetRef` is empty when either `Kind` or `ID` is empty.

The session-owned target is converted into a target ref through:

```go
func (targeting PlayerTargeting) TargetRef() targetpolicy.TargetRef
```

It is copied onto an active ship through:

```go
func (targeting PlayerTargeting) ApplyToShip(ship *runtime.Ship)
```

The active ship fields are packet-facing runtime copies:

```text
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

They are projected by `runtime.Ship.State()` into:

```text
ShipState.target_kind
ShipState.target_id
```

## Target mutation flow

The generic target setter is:

```go
func (game *Game) SetTarget(playerID string, target targetpolicy.TargetRef) bool
```

It requires the requesting player session to exist.

When the target is empty, the server clears the player's canonical target.

When the target is non-empty, the server validates that the referenced target exists before storing it. Invalid non-empty targets are rejected and do not overwrite the previous target.

Point-based selection is handled by:

```go
func (game *Game) SelectTargetAtPosition(playerID string, x float64, y float64, target targetpolicy.TargetRef) bool
```

It accepts only when:

```text
requesting player session exists
requested target ref is non-empty
target exists in authoritative server state
target appears in current server target candidates
click point is inside the matched candidate collision body
session targeting update succeeds
```

This means the client cannot make a target canonical merely by sending an ID. For point-based selection, the server re-checks existence and spatial overlap against server-side collision bodies.

Clearing is handled by:

```go
func (game *Game) ClearTarget(playerID string)
```

The player-only compatibility wrapper is:

```go
func (game *Game) SetPlayerTarget(playerID string, targetPlayerID string) bool
```

That wrapper converts a player ID into the generic canonical target shape:

```text
target_kind = player
target_id = targetPlayerID
```

New generic gameplay systems should use `SetTarget`, `SelectTargetAtPosition`, `ClearTarget`, and `TargetRef` rather than adding new player-only target fields.

## Session and active-avatar split

Canonical target state is session-owned.

The active ship only carries the current packet-facing copy.

When target state changes, the server updates:

```text
playerSession.Targeting
```

If the player currently has an active ship, the server also copies the same target onto:

```text
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

When a new ship is created from a session, `session.NewShip(position)` calls:

```go
session.Targeting.ApplyToShip(ship)
```

This preserves target state across respawn.

The important ownership split is:

```text
playerSession.Targeting
= authoritative match-local selected target

runtime.Ship.TargetKind / TargetID
= active-avatar projection copy

StatePacket.players[*].target_kind / target_id
= packet-facing read model for active ships
```

A player with no active ship can still have target state in the session. That target is not visible through `StatePacket.players` until the player has an active ship again, because `StatePacket.players` contains active avatar state only.

## Target existence and status

Target existence is checked against authoritative server state by target kind.

Current existence rules:

```text
player
= player world state exists

enemy
= game.entities.Enemies[target_id] exists and is non-nil

pickup
= game.entities.Pickups[target_id] exists and is non-nil

asteroid
= game.entities.Asteroids[target_id] exists and is non-nil

bullet
= game.entities.Projectiles[target_id] exists and is non-nil
```

Target status is separate from target existence.

Current target status values are:

```text
missing
inactive
active
```

For players, status is derived from `player.WorldState`:

```text
missing
= no player session/world state exists

active
= player world state exists and is targetable

inactive
= player world state exists but is not targetable
```

Pending-respawn players are inactive, not missing. Removed players are missing.

For asteroids, bullets, and enemies, pending-despawn entities are inactive. Missing or nil entities are missing. Present non-pending entities are active.

For pickups, present pickups are currently active. Missing or nil pickups are missing.

`clearTargetsForMissingPlayersLocked` clears targets only when the selected target has become missing. It does not clear a target merely because the selected player is inactive or pending respawn.

## Target candidates

Point-based target selection uses server-built target candidates.

Current candidate kinds are:

```text
players
asteroids
bullets
pickups
enemies
```

The server builds candidates from authoritative entity maps and collision bodies.

Player candidates skip nil players and pending-despawn ships.

Asteroid, bullet, pickup, and enemy candidates skip nil entities. Their collision body must be available from the collision shape catalog or entity collision body method.

The candidate stores:

```text
TargetRef
CollisionBody
```

`SelectTargetAtPosition` then compares the requested target ref to the candidate ref and checks whether the requested click point is inside the candidate body.

Target candidate generation is server-side validation support. It is not the same as client visual picking and does not own presentation.

## Protocols and APIs

Canonical target state is affected by realtime gameplay packets.

The current target request packet types are:

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

The coordinates are the server-space point being claimed for the selected target. The target fields identify the candidate the client is asking the server to select.

`clear_target_request` clears the canonical target for the current game player.

`set_target_player_request` is a legacy-compatible player-target request path that still routes to canonical generic target state by converting the requested target into `target_kind = player` and `target_id = <player id>`.

Inbound routing is owned by networking. The gameplay inbound handler routes target packets to the current room's game instance only when the session has a current room and current game-player ID.

State readback is through `StatePacket.players`.

The server projects active ship target copies as:

```text
players[player_id].target_kind
players[player_id].target_id
```

The client should treat those fields as authoritative confirmation. Local click state is only request state until the server projects the accepted target back.

## Data ownership

Canonical target state is in-memory match runtime state.

It reads:

```text
player sessions
active player ships
enemy entities
pickup entities
asteroid entities
projectile entities
collision shape catalog
camera/player world state for inactive player status
target request packet fields
```

It mutates:

```text
playerSession.Targeting
runtime.Ship.TargetKind
runtime.Ship.TargetID
```

It does not persist target state outside the current game instance.

Packet shape source data lives in:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
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

Canonical target state must preserve these rules:

* Generic gameplay targeting uses `target_kind` plus `target_id`.
* `playerSession.Targeting` is the authoritative match-local target owner.
* `runtime.Ship.TargetKind` and `runtime.Ship.TargetID` are active-avatar copies, not durable ownership.
* Target state must survive player death or missing active ship when the player session remains.
* Respawned ships must receive the session-owned target copy.
* Empty target clears must be accepted for existing players.
* Invalid non-empty targets must not overwrite existing target state.
* Point-based selection must validate both target existence and server collision-body overlap.
* Pending-respawn player targets are inactive, not missing.
* Removed player targets are missing and should be cleared from remaining active players.
* Targeting state alone does not make an entity damageable, collidable, commandable, or eligible for a player-only devtools command.
* New gameplay packets and read models must not introduce `target_player_id`.

## Devtools and `target_player_id` boundary

`target_player_id` is not canonical gameplay target state.

It is a legacy player-only devtools/debug command compatibility surface. It may remain in debug packet schemas, debug handlers, generated debug packet code, and tests for those debug commands.

Normal gameplay targeting must use:

```text
target_kind
target_id
```

Player-only systems that already know they are operating on a player may use direct player IDs. They should not create a second targeting model.

Devtools may resolve a player-only command from the current canonical game target only when the canonical target is a player:

```text
target_kind == "player"
```

Non-player canonical targets such as `asteroid`, `bullet`, `enemy`, or `pickup` are valid canonical targets for readout/inspection, but they are not valid `target_player_id` sources for player-only debug commands.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/targeting.go
```

Owns `Game.SetTarget`, `Game.SelectTargetAtPosition`, `Game.ClearTarget`, target existence checks, target lookup status, target candidate construction, and missing-player target cleanup.

```text
services/game-server/internal/game/player_targeting.go
```

Owns `PlayerTargeting`, conversion to and from `targeting.TargetRef`, empty session targeting, and applying session target state onto an active ship.

```text
services/game-server/internal/game/session.go
```

Stores `playerSession.Targeting` and copies targeting onto newly created ships through `session.NewShip`.

```text
services/game-server/internal/game/runtime/ship.go
```

Projects `runtime.Ship.TargetKind` and `runtime.Ship.TargetID` into `runtime.ShipState`.

```text
services/game-server/internal/game/targeting/targeting.go
```

Defines canonical target kinds, `TargetRef`, `TargetCandidate`, empty target behavior, target kind priority, and legacy requested-player target validation support.

```text
services/game-server/internal/game/player/state.go
services/game-server/internal/game/player/target_status.go
services/game-server/internal/game/player_world_state.go
```

Define player world state and target status classification used to distinguish active, inactive, and missing player targets.

```text
services/game-server/internal/networking/inbound/gameplay.go
```

Routes target request packets from the current network session to the current room game instance.

Generated/source packet files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Related active-avatar and packet projection files:

```text
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/runtime/state.go
```

Important non-ownership boundaries:

```text
client/scripts/gameplay/input/
client/scripts/gameplay/targeting/
services/game-server/internal/devtools/
services/game-server/internal/game/damage/
services/game-server/internal/game/weapons/
services/game-server/internal/game/rules/
services/game-server/internal/networking/
services/game-server/internal/protocol/packetcodec/
```

`client` owns local input, candidate presentation, and request sending.

`devtools` owns command-specific debug target interpretation.

`damage` owns damage result math.

`weapons` owns fire policy and projectile spawn intent.

`rules` owns match lifecycle classification.

`networking` owns session routing and transport.

`packetcodec` owns JSON encode/decode mechanics.

## Tests and verification

Relevant game-server tests include:

```text
services/game-server/internal/game/targeting_test.go
services/game-server/internal/game/targeting/targeting_test.go
services/game-server/internal/game/player/target_status_test.go
services/game-server/internal/game/player/state_test.go
services/game-server/internal/game/player_world_state_test.go
services/game-server/tests/networking/rooms_test.go
```

Current coverage includes:

* player target selection stores `target_kind = player` and the target player ID
* `StatePacket.players` includes `target_kind` and `target_id`
* generic `SetTarget` stores player, asteroid, and bullet targets
* point-based selection stores player, asteroid, bullet, and pickup targets when the click overlaps the authoritative collision body
* missing targets do not overwrite an existing target
* non-overlapping point selection does not overwrite an existing target
* clearing target stores the empty target
* target lookup distinguishes active, inactive pending-respawn, and missing player targets
* missing removed player targets are cleared
* inactive pending-respawn player targets are preserved
* dead players can update and clear session target state without an active ship
* respawned ships inherit session-owned target state
* target request packets route through networking into server target selection

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/... ./tests/networking/...
```

Focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game -run 'Target'
go test -buildvcs=false ./internal/game/targeting
go test -buildvcs=false ./internal/game/player -run 'TargetStatus|WorldState'
go test -buildvcs=false ./tests/networking -run 'Target'
```

Run generated packet checks when target packet fields or packet output generation changes:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Game Server Simulation Targeting](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Session State](../players/player-session-state.md)
* [Player Death And Despawn](../players/player-death-and-despawn.md)
* [Player Respawn](../players/player-respawn.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Input And Targeting](../../../client/input-and-targeting.md)
* [Realtime Protocol](../../../../protocol/!README.md)
* [Data](../../../../data/!README.md)
* [Devtools](../../../../devtools/!README.md)

## Notes

The most important migrated legacy rule is the quarantine boundary: `target_player_id` must not leak back into normal gameplay targeting. Canonical gameplay target identity is `target_kind` plus `target_id`.

Current `StatePacket` projection exposes target state only through active ship state. Session-owned target state can still exist when a player has no active ship, but it is not separately projected in `player_sessions`.

`enemy` is a supported canonical target kind in the server targeting package and target candidate builder. Current enemy projection and full enemy gameplay behavior are not documented here.
